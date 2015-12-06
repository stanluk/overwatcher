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
overwatch overtime --day="12121" --reason="MSG"
overwarch status --announce --day-length=8h --day="2015-12-11"
overwatch report --template="<path>" --from="" --to="" --day-length=8h --after=1h
*/

func main() {
	var workdir string = os.Getenv("HOME")
	if workdir == "" {
		workdir, _ = os.Getwd()
	}
	err := InitSqlDb(path.Join(workdir, ".overwatcher.db"))
	if err != nil {
		log.Fatal(err)
	}
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)

	overtimeCmd := flag.NewFlagSet("overtime", flag.ExitOnError)
	dayFlag := overtimeCmd.String("day", time.Now().Format("2006-Jan-02"), "day to update in YYYY-Month-DD format")
	reasonFlag := overtimeCmd.String("reason", "", "reson of overtime")

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

	switch os.Args[1] {
	case "start":
		startCmd.Parse(os.Args[2:])
	case "stop":
		stopCmd.Parse(os.Args[2:])
	case "overtime":
		overtimeCmd.Parse(os.Args[2:])
	default:
		log.Fatal("%q is not valid command", os.Args[1])
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
	if overtimeCmd.Parsed() {
		var day time.Time
		if *dayFlag == "" {
			day = time.Now()
		} else {
			day, err = time.Parse("2006-Jan-02", *dayFlag)
			if err != nil {
				log.Fatal("failed: ", err)
			}
		}
		err = GiveReason(*reasonFlag, day)
		if err != nil {
			log.Fatal(err)
		}
	}
	ShutdownSqlDb()
}
