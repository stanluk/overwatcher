package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

/*
overwatch start
overwatch stop
overwatch reason --day="12121" <MSG>
overwatch report --template="" --from="" --to=""
*/

func main() {
	err := InitSqlDb("/home/stanluk/.overtimes")
	if err != nil {
		log.Fatal(err)
	}
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)

	reasonCmd := flag.NewFlagSet("reason", flag.ExitOnError)
	dayFlag := reasonCmd.String("day", time.Now().String(), "day to update in YYYY-Month-DD format")
	reasonFlag := reasonCmd.String("r", "", "reson of overtime")

	if len(os.Args) == 1 {
		fmt.Println("overwatcher <command>")
		fmt.Println("")
		fmt.Println("work time logging")
		fmt.Println("\tstart - log workday start")
		fmt.Println("\tend - log workday end (can be called multiples times a day)")
		fmt.Println("")
		fmt.Println("work time report")
		fmt.Println("\treport - genereate overtime report")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		startCmd.Parse(os.Args[2:])
	case "stop":
		stopCmd.Parse(os.Args[2:])
	case "reason":
		reasonCmd.Parse(os.Args[2:])
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
	if reasonCmd.Parsed() {
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
