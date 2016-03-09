package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"time"
)

var lockFilePath, alarmLockFilePath string

const defaultTimeFormat string = "2006-Jan-02"

/*
overwatch start <hour>
overwatch stop <hour>
overwatch overtime --day="12121" --reason="MSG"
overwarch query --from="2015-12-11" --to="2015-12-11" --week --month
overwatch alarm enable <8h> | disable | check
overwatch report --template="<path>" --from="" --to="" --workday=8h --after=1h --gran=30m
*/

func runStart(when time.Time) {
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString(when.UTC().Format(time.RFC822Z))
	if err != nil {
		log.Fatal(err)
	}
}

func runEnd(when time.Time) {
	file, err := os.Open(lockFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	str, err := buf.ReadString('\n')
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	start, err := time.Parse(time.RFC822Z, str)
	if err != nil {
		log.Fatal(err)
	}
	err = CreateWorkLog(start, when.UTC())
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	os.Remove(lockFilePath)
}

func updateOvertime(dayFlag, reasonFlag *string) {
	var day time.Time
	var err error
	if *dayFlag == "" {
		day = time.Now()
	} else {
		day, err = time.Parse("2006-Jan-02", *dayFlag)
		if err != nil {
			log.Fatal("failed: ", err)
		}
	}
	err = UpdateOvertime(day, reasonFlag)
	if err != nil {
		log.Fatal(err)
	}
}

func queryLogs(out io.Writer, from, to time.Time) {
	var err error
	logs, err := QueryLogs(from, to)
	if err != nil {
		log.Fatal(err)
	}
	if len(logs) == 0 {
		fmt.Fprintln(out, "No worklog")
		return
	}
	for _, log := range logs {
		fmt.Fprintln(out, log.Start.Format(defaultTimeFormat), "\t",
			log.Start.Format(time.Kitchen), "\t", log.End.Format(time.Kitchen),
			"\t", log.End.Sub(log.Start).String(), "\t", log.NetLen.String())
	}
}

func dateFromTodayHour(hour time.Time) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour.Hour(), hour.Minute(), hour.Second(), 0, time.Local)
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	err = InitSQLDb(path.Join(usr.HomeDir, ".overwatcher.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer ShutdownSQLDb()

	lockFilePath = path.Join(usr.HomeDir, ".overwatcher.lock")
	alarmLockFilePath = path.Join(usr.HomeDir, ".overwatcher.alarm")

	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startTime := startCmd.String("time", time.Now().Format(time.Kitchen), "time of workday start")

	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	endTime := stopCmd.String("time", time.Now().Format(time.Kitchen), "time of workday end")

	overtimeCmd := flag.NewFlagSet("update", flag.ExitOnError)
	dayFlag := overtimeCmd.String("day", time.Now().Format(defaultTimeFormat), "day to update (YYYY-Month-DD)")
	reasonFlag := overtimeCmd.String("reason", "", "reson of overtime")

	queryCmd := flag.NewFlagSet("query", flag.ExitOnError)
	fromFlag := queryCmd.String("from", time.Now().Format(defaultTimeFormat), "first day to query (YYYY-Month-DD)")
	toFlag := queryCmd.String("to", time.Now().Format(defaultTimeFormat), "last day to query (YYYY-Month-DD)")
	weekFlag := queryCmd.Bool("week", false, "print worklog for current week")
	monthFlag := queryCmd.Bool("month", false, "print worklog for current month")

	enableAlarmSubCmd := flag.NewFlagSet("enable", flag.ExitOnError)
	durationLen := enableAlarmSubCmd.Duration("after", time.Hour*8, "Duration since first work start")

	disableAlarmSubCmd := flag.NewFlagSet("disable", flag.ExitOnError)
	checkAlarmSubCmd := flag.NewFlagSet("check", flag.ExitOnError)

	if len(os.Args) == 1 {
		fmt.Println("overwatcher <command>")
		fmt.Println("")
		fmt.Println("work time logging")
		fmt.Println("\tstart - log workday start")
		fmt.Println("\tstop - log workday end (can be called multiples times a day)")
		fmt.Println("\tovertime - add info about overtime")
		fmt.Println("\tquery - log info about work day")
		fmt.Println("\talarm - inform about work day end")
		fmt.Println("\treport - genereate overtimes report")
		os.Exit(1)
	}

	err = nil

	switch os.Args[1] {
	case "start":
		err = startCmd.Parse(os.Args[2:])
	case "stop":
		err = stopCmd.Parse(os.Args[2:])
	case "overtime":
		err = overtimeCmd.Parse(os.Args[2:])
	case "query":
		err = queryCmd.Parse(os.Args[2:])
	case "alarm":
		if len(os.Args) < 3 {
			log.Fatal("enable|disble|check expected")
		}
		switch os.Args[2] {
		case "enable":
			err = enableAlarmSubCmd.Parse(os.Args[3:])
		case "disable":
			err = disableAlarmSubCmd.Parse(os.Args[3:])
		case "check":
			err = checkAlarmSubCmd.Parse(os.Args[3:])
		default:
			log.Fatal("enable|disble|check expected")
		}
	default:
		log.Fatalf("%q is not valid command", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
	if startCmd.Parsed() {
		tm, err := time.Parse(time.Kitchen, *startTime)
		if err != nil {
			log.Fatal(err)
		}
		runStart(dateFromTodayHour(tm))
	}
	if stopCmd.Parsed() {
		tm, err := time.Parse(time.Kitchen, *endTime)
		if err != nil {
			log.Fatal(err)
		}
		runEnd(dateFromTodayHour(tm))
	}
	if overtimeCmd.Parsed() {
		updateOvertime(dayFlag, reasonFlag)
	}
	if queryCmd.Parsed() {
		var to, from time.Time
		if *weekFlag {
			to = time.Now()
			from = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-7, 0, 0, 0, 0, time.Local)
		} else if *monthFlag {
			from = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
			to = from.AddDate(0, 1, 0)
		} else {
			from, _ = time.Parse(defaultTimeFormat, *fromFlag)
			to, _ = time.Parse(defaultTimeFormat, *toFlag)
		}
		queryLogs(os.Stdout, from, to)
	}
	if enableAlarmSubCmd.Parsed() {
		log.Println(durationLen)
	}
	if disableAlarmSubCmd.Parsed() {
		log.Println("DISABLE")
	}
	if checkAlarmSubCmd.Parsed() {
		log.Println("CHECK")
	}
}
