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

var lockFilePath string

const defaultTimeFormat string = "2006-Jan-02"

/*
overwatch start
overwatch stop
overwatch overtime --day="12121" --reason="MSG"
overwarch query --from="2015-12-11" --to="2015-12-11" --week --month
overwatch report --template="<path>" --from="" --to="" --workday=8h --after=1h --gran=30m
*/

func runStart() {
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString(time.Now().UTC().Format(time.RFC822Z))
	if err != nil {
		log.Fatal(err)
	}
}

func runEnd() {
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
	err = CreateWorkLog(start, time.Now().UTC())
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
		fmt.Println("No worklog")
		return
	}
	for _, log := range logs {
		fmt.Println("Start:\t ", log.Start.String())
		fmt.Println("End:\t ", log.End.String())
		fmt.Println("Net workime:\t ", log.NetLen.String())
	}
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

	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)

	overtimeCmd := flag.NewFlagSet("update", flag.ExitOnError)
	dayFlag := overtimeCmd.String("day", time.Now().Format(defaultTimeFormat), "day to update (YYYY-Month-DD)")
	reasonFlag := overtimeCmd.String("reason", "", "reson of overtime")

	queryCmd := flag.NewFlagSet("query", flag.ExitOnError)
	fromFlag := queryCmd.String("from", time.Now().Format(defaultTimeFormat), "first day to query (YYYY-Month-DD)")
	toFlag := queryCmd.String("to", time.Now().Format(defaultTimeFormat), "last day to query (YYYY-Month-DD)")
	//announceFlag := queryCmd.Bool("announce", false, "print message on all pseudo terminals")
	//nowFlag := queryCmd.Bool("now", false, "calculate worktime till now.")
	//workdayFlag := statusCmd.String("workday", "8h", "workday length (default=8h)")

	if len(os.Args) == 1 {
		fmt.Println("overwatcher <command>")
		fmt.Println("")
		fmt.Println("work time logging")
		fmt.Println("\tstart - log workday start")
		fmt.Println("\tstop - log workday end (can be called multiples times a day)")
		fmt.Println("\tovertime - add info about overtime")
		fmt.Println("\tquery - log info about work day")
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
	if overtimeCmd.Parsed() {
		updateOvertime(dayFlag, reasonFlag)
	}
	if queryCmd.Parsed() {
		from, _ := time.Parse(defaultTimeFormat, *fromFlag)
		to, _ := time.Parse(defaultTimeFormat, *toFlag)
		queryLogs(os.Stdout, from, to)
	}
}
