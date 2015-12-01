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
	dayFlag := reasonCmd.String("day", "", "2015-Jan-01")

	if len(os.Args) == 1 {
		fmt.Println("Invalid usage")
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
		log.Fatal("%q is not valid command")
	}

	if startCmd.Parsed() {
		err = StartWork()
		if err != nil {
			log.Fatal(err)
		}
	}
	if stopCmd.Parsed() {
		err = EndWork()
		if err != nil {
			log.Fatal(err)
		}
	}
	if reasonCmd.Parsed() {
		var day time.Time
		if *dayFlag == "" {
			day = time.Now()
		} else {
			day, err = time.Parse("2015-Jan-01", *dayFlag)
			if err != nil {
				log.Fatal(err)
			}
		}
		err = GiveReason("", day)
		if err != nil {
			log.Fatal(err)
		}
	}

	ShutdownSqlDb()
}
