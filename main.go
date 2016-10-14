package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"text/template"
	"time"
)

var (
	startCmd  flag.FlagSet // overwatcher start
	stopCmd   flag.FlagSet // overwatcher stop
	updateCmd flag.FlagSet // overwatcher update
	queryCmd  flag.FlagSet // overwatcher query
	reportCmd flag.FlagSet // overwatcher report

	// overwatcher update flags
	day    string // shared with 'overwatcher report'
	reason string
	enter  string
	leave  string
	breaks string

	// overwatcher query flags
	week     bool
	month    bool
	fromDate string
	toDate   string

	// overwatcher report flags
	templatePath string
)

const defaultTimeFormat string = "2006-Jan-02"

const help string = `
overwatcher <command>

work time logging script.

	start		log workday start
	stop		log workday end (can be called multiples times a day)
	update		update info about workday (enter, leave time, overtime reason)
	query		display info about work workdays
	report		genereate overtimes report

Please type "overwatcher <command> -h" for help
`

const logTemplate = `{{"Day" | printf "%-15s"}}{{"Enter" | printf "%-15s"}}{{"Leave" | printf "%-15s"}}{{"TotalWorktime" | printf "%-15s"}}{{"Breaks" | printf "%-15s"}}{{"Reason" | printf "%-15s"}}
{{range .}}{{.EnterTime | format_date "2006-Jan-02" | printf "%-15s"}}{{.EnterTime | format_date "3:04PM" | printf "%-15s"}}{{.LeaveTime | format_date "3:04PM" | printf "%-15s"}}{{.TotalTime | printf "%-15s"}}{{.Breaks | printf "%-15s"}}{{.OvertimeReason | printf "%-15s"}}
{{end}}`

// functions needed for template formatting
var logTemplateFuncs = template.FuncMap{
	"format_date": func(format string, date time.Time) string {
		return date.Format(format)
	},
	"parse_duration": func(diff string) time.Duration {
		dur, err := time.ParseDuration(diff)
		if err != nil {
			log.Fatal("ParseDuration failed: %q", err)
		}
		return dur
	},
	"add": func(tm time.Time, diff time.Duration) time.Time {
		return tm.Add(diff)
	},
	"sub": func(diff1, diff2 time.Duration) time.Duration {
		return diff1 - diff2
	},
	"round_down": func(diff1, granulation time.Duration) time.Duration {
		return diff1 - diff1%granulation
	},
	"round_up": func(diff1, granulation time.Duration) time.Duration {
		return diff1 + (granulation - diff1%granulation)
	},
}

func handleStartCmd() {
	wl, err := QueryWorkLog(time.Now())
	if err != nil {
		log.Fatal(err)
	}
	if wl == nil {
		wl = NewWorkLog()
	}
	wl.SetEnterTime(time.Now())
	err = StoreWorkLog(wl)
	if err != nil {
		log.Fatal(err)
	}
}

func handleStopCmd() {
	wl, err := QueryWorkLog(time.Now())
	if err != nil {
		log.Fatal(err)
	}
	if wl == nil {
		log.Fatal("No worklogs for today...")
	}
	wl.SetLeaveTime(time.Now())
	err = StoreWorkLog(wl)
	if err != nil {
		log.Fatal(err)
	}
}

func ParseHourAtDay(day time.Time, kitchenString string) (time.Time, error) {
	hourTime, err := time.ParseInLocation(time.Kitchen, kitchenString, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(day.Year(), day.Month(), day.Day(),
		hourTime.Hour(), hourTime.Minute(), hourTime.Second(), 0, day.Location()), nil
}

func handleUpdateCmd() {
	var dayTime, newTime time.Time
	var err error

	// parse day
	if day != "" {
		if dayTime, err = time.ParseInLocation(defaultTimeFormat, day, time.Local); err != nil {
			log.Fatal("Invalid date format: ", err)
		}
	} else {
		dayTime = time.Now()
	}

	worklog, err := QueryWorkLog(dayTime)
	if err != nil {
		log.Fatal(err)
	}

	if worklog == nil {
		worklog = &WorkLog{enter: dayTime, leave: dayTime}
	}

	if reason != "" {
		worklog.OvertimeReason = reason
	}
	if breaks != "" {
		dur, err := time.ParseDuration(breaks)
		if err != nil {
			log.Fatal("Invalid break duration: ", dur)
		}
		worklog.Breaks = dur
	}
	if enter != "" {
		if newTime, err = ParseHourAtDay(dayTime, enter); err != nil {
			log.Fatal("Invalid date format: ", err)
		}
		if err = worklog.SetEnterTime(newTime); err != nil {
			log.Fatal(err)
		}
	}
	if leave != "" {
		if newTime, err = ParseHourAtDay(dayTime, leave); err != nil {
			log.Fatal("Invalid date format: ", err)
		}
		if err = worklog.SetLeaveTime(newTime); err != nil {
			log.Fatal(err)
		}
	}
	if err = StoreWorkLog(worklog); err != nil {
		log.Fatalln(err)
	}
}

func handleReportCommand() {
	var dayTime time.Time
	var err error

	// parse day
	if day != "" {
		if dayTime, err = time.ParseInLocation(defaultTimeFormat, day, time.Local); err != nil {
			log.Fatal("Invalid date format: ", err)
		}
	} else {
		dayTime = time.Now()
	}

	// get template
	if templatePath == "" {
		log.Fatal("No template parameter, please check \"overwatcher report -h\"")
	}
	tmpl, err := template.New(templatePath).Funcs(logTemplateFuncs).ParseFiles(templatePath)
	if err != nil {
		log.Fatal("Unable to load template: %q", err)
	}

	worklog, err := QueryWorkLog(dayTime)
	if err != nil {
		log.Fatal(err)
	}

	if worklog == nil {
		log.Fatal("No worklog in day: ", dayTime)
	}

	err = tmpl.ExecuteTemplate(os.Stdout, templatePath, worklog)
	if err != nil {
		log.Fatal(err)
	}
}

func handleQueryCommand() {
	var to, from time.Time
	if week {
		to = time.Now()
		from = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-7, 0, 0, 0, 0, time.Local)
	} else if month {
		from = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
		to = from.AddDate(0, 1, 0)
	} else {
		from, _ = time.ParseInLocation(defaultTimeFormat, fromDate, time.Local)
		to, _ = time.ParseInLocation(defaultTimeFormat, toDate, time.Local)
	}
	logs, err := QueryWorkLogs(from, to)
	if err != nil {
		log.Fatal(err)
	}
	printLogs(os.Stdout, logs)
}

func printLogs(out io.Writer, logs []*WorkLog) {
	if len(logs) == 0 {
		fmt.Fprintln(out, "No worklogs")
		return
	}
	templ, err := template.New("log").Funcs(logTemplateFuncs).Parse(logTemplate)
	if err != nil {
		panic("unable to parse template")
	}
	err = templ.Execute(out, logs)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	updateCmd.StringVar(&day, "day", "", "day to update (eg. 2016-Oct-11)")
	updateCmd.StringVar(&reason, "reason", "", "reson of overtime")
	updateCmd.StringVar(&enter, "enter", "", "time of enter (eg. 7:22AM, 10:23PM)")
	updateCmd.StringVar(&leave, "leave", "", "time of leave (eg. 7:22AM, 10:23PM)")
	updateCmd.StringVar(&breaks, "break", "", "breaks made during workday (eg. 2h30m)")
	queryCmd.StringVar(&fromDate, "from", time.Now().Format(defaultTimeFormat), "day to query (YYYY-Month-DD)")
	queryCmd.StringVar(&toDate, "to", time.Now().AddDate(0, 0, 1).Format(defaultTimeFormat), "day to query (YYYY-Month-DD)")
	queryCmd.BoolVar(&week, "week", false, "print worklogs for this week")
	queryCmd.BoolVar(&month, "month", false, "print worklogs for this month")
	reportCmd.StringVar(&templatePath, "template", "", "path to report template.")
	reportCmd.StringVar(&day, "day", "", "day to update (eg. 2016-Oct-11)")
}

func initDb() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	err = InitDb(path.Join(usr.HomeDir, ".overwatcher.db"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	initDb()

	if len(os.Args) == 1 {
		fmt.Println(help)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		err = startCmd.Parse(os.Args[2:])
	case "stop":
		err = stopCmd.Parse(os.Args[2:])
	case "update":
		err = updateCmd.Parse(os.Args[2:])
	case "query":
		err = queryCmd.Parse(os.Args[2:])
	case "report":
		err = reportCmd.Parse(os.Args[2:])
	case "help":
		fmt.Println(help)
		os.Exit(0)
	default:
		log.Fatalf("%q is not valid command.", os.Args[1])
	}
	if err != nil {
		os.Exit(1)
	}
	if startCmd.Parsed() {
		handleStartCmd()
	}
	if stopCmd.Parsed() {
		handleStopCmd()
	}
	if updateCmd.Parsed() {
		handleUpdateCmd()
	}
	if queryCmd.Parsed() {
		handleQueryCommand()
	}
	if reportCmd.Parsed() {
		handleReportCommand()
	}
}
