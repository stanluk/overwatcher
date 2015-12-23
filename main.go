package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

/*
overwatch start
overwatch stop
overwatch update --day="12121" --reason="MSG" --start="hour" --stop="hour"
overwarch status --announce --workday=8h --day="2015-12-11"
overwatch report --template="<path>" --from="" --to="" --workday=8h --after=1h --gran=30m
*/

func runStart() {
	err := StartWork(time.Now())
	if err != nil {
		log.Fatal(err)
	}
}

func runEnd() {
	err := EndWork(time.Now())
	if err != nil {
		log.Fatal(err)
	}
}

func runUpdate(dayFlag, startFlag, stopFlag, reasonFlag *string) {
	var day time.Time
	var start, stop *time.Time
	var err error
	if *dayFlag == "" {
		day = time.Now()
	} else {
		day, err = time.Parse("2006-Jan-02", *dayFlag)
		if err != nil {
			log.Fatal("failed: ", err)
		}
	}
	if *startFlag != "" {
		sta, err := time.Parse(time.Kitchen, *startFlag)
		if err != nil {
			log.Fatal(err)
		}
		start = &sta
	}
	if *stopFlag != "" {
		sto, err := time.Parse(time.Kitchen, *stopFlag)
		if err != nil {
			log.Fatal(err)
		}
		stop = &sto
	}
	err = UpdateLog(day, start, stop, reasonFlag)
	if err != nil {
		log.Fatal(err)
	}
}

func runStatus(dayFlag2 *string) {
	var day time.Time
	var err error
	if *dayFlag2 == "" {
		day = time.Now()
	} else {
		day, err = time.Parse("2006-Jan-02", *dayFlag2)
		if err != nil {
			log.Fatal("failed: ", err)
		}
	}
	logs, err := QueryLogs(day, day)
	if err != nil {
		log.Fatal(err)
	}
	if len(logs) == 0 {
		fmt.Println("No worklog")
		return
	}
	fmt.Println("Worktime:\t ", logs[0].End.Sub(logs[0].Start).String())
	fmt.Println("Start:\t", logs[0].Start.Format(time.Kitchen))
	fmt.Println("End:\t", logs[0].End.Format(time.Kitchen))
}

func main() {
	workdir := os.Getenv("HOME")
	if workdir == "" {
		workdir, _ = os.Getwd()
	}
	err := InitSQLDb(path.Join(workdir, ".overwatcher.db"))
	if err != nil {
		log.Fatal(err)
	}
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	dayFlag := updateCmd.String("day", time.Now().Format("2006-Jan-02"), "day to update (YYYY-Month-DD)")
	startFlag := updateCmd.String("start", "", "work start hour HH:MM")
	stopFlag := updateCmd.String("stop", "", "work stop hour HH:MM")
	reasonFlag := updateCmd.String("reason", "", "reson of overtime")

	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	dayFlag2 := statusCmd.String("day", time.Now().Format("2006-Jan-02"), "day to query (YYYY-Month-DD)")
	//announceFlag := statusCmd.Bool("announce", false, "print message on all pseudo terminals")
	//workdayFlag := statusCmd.String("workday", "8h", "workday length (default=8h)")

	if len(os.Args) == 1 {
		fmt.Println("overwatcher <command>")
		fmt.Println("")
		fmt.Println("work time logging")
		fmt.Println("\tstart - log workday start")
		fmt.Println("\tend - log workday end (can be called multiples times a day)")
		fmt.Println("\tupdate - add info about overtime")
		fmt.Println("\tstatus - log info about work day")
		fmt.Println("\treport - genereate overtimes report")
		os.Exit(1)
	}

	err = nil

	switch os.Args[1] {
	case "start":
		err = startCmd.Parse(os.Args[2:])
	case "stop":
		err = stopCmd.Parse(os.Args[2:])
	case "update":
		err = updateCmd.Parse(os.Args[2:])
	case "status":
		err = statusCmd.Parse(os.Args[2:])
	default:
		log.Fatalf("%q is not valid command", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}

	if startCmd.Parsed() {
		runStart()
	}
	if stopCmd.Parsed() {
		runEnd()
	}
	if updateCmd.Parsed() {
		runUpdate(dayFlag, startFlag, stopFlag, reasonFlag)
	}
	if statusCmd.Parsed() {
		runStatus(dayFlag2)
	}
	err = ShutdownSQLDb()
	if err != nil {
		log.Fatal(err)
	}
}
