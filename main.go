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
overwarch status --announce --day-length=8h --day="2015-12-11"
overwatch report --template="<path>" --from="" --to="" --day-length=8h --after=1h
*/

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
	dayFlag := updateCmd.String("day", time.Now().Format("2006-Jan-02"), "day to update in YYYY-Month-DD format")
	startFlag := updateCmd.String("start", "", "work start hour HH:MM")
	stopFlag := updateCmd.String("stop", "", "work stop hour HH:MM")
	reasonFlag := updateCmd.String("reason", "", "reson of overtime")

	if len(os.Args) == 1 {
		fmt.Println("overwatcher <command>")
		fmt.Println("")
		fmt.Println("work time logging")
		fmt.Println("\tstart - log workday start")
		fmt.Println("\tend - log workday end (can be called multiples times a day)")
		fmt.Println("\tovertime - add info about overtime")
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
	default:
		log.Fatalf("%q is not valid command", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}

	if startCmd.Parsed() {
		err = StartWork(time.Now())
		if err != nil {
			log.Fatal(err)
		}
	}
	if stopCmd.Parsed() {
		err = EndWork(time.Now())
		if err != nil {
			log.Fatal(err)
		}
	}
	if updateCmd.Parsed() {
		var day time.Time
		var start, stop time.Time
		if *dayFlag == "" {
			day = time.Now()
		} else {
			day, err = time.Parse("2006-Jan-02", *dayFlag)
			if err != nil {
				log.Fatal("failed: ", err)
			}
		}
		if *startFlag != "" {
			start, err = time.Parse(time.Kitchen, *startFlag)
			if err != nil {
				log.Fatal(err)
			}
		}
		if *stopFlag != "" {
			stop, err = time.Parse(time.Kitchen, *stopFlag)
			if err != nil {
				log.Fatal(err)
			}
		}
		err = UpdateLog(day, &start, &stop, reasonFlag)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = ShutdownSQLDb()
	if err != nil {
		log.Fatal(err)
	}
}
