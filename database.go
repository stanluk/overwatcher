package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Overtime struct {
	Start  time.Time
	End    time.Time
	Reason string
}

var db *sql.DB

func InitSqlDb(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS overtimes (day date PRIMARY KEY, start time, end time, reason varchar(256));`)

	return err
}

func ShutdownSqlDb() error {
	return db.Close()
}

func StartWork(now time.Time) error {
	res, err := db.Exec("INSERT INTO overtimes VALUES (DATE(?), ?, ?, ?);", now, now, now, "")
	if err != nil {
		return err
	}
	ra, err := res.LastInsertId()
	if err != nil {
		return err
	}
	if ra <= 0 {
		return errors.New(fmt.Sprintf("Unable to insert record %s", now.String()))
	}
	return nil
}

func EndWork(now time.Time) error {
	res, err := db.Exec("UPDATE overtimes SET end=? WHERE day=DATE(?)", now, now)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra <= 0 {
		return errors.New(fmt.Sprintf("No work record from day %s", time.Now().String()))
	}
	return nil
}

func GiveReason(reason string, when time.Time) error {
	res, err := db.Exec("UPDATE overtimes SET reason=? WHERE day=DATE(?)", reason, when)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows <= 0 {
		return errors.New(fmt.Sprintf("No work record from day %s", when.String()))
	}
	return nil
}

func OvertimesReport(start, end time.Timer) ([]*Overtime, error) {
	var ret []*Overtime
	rows, err := db.Query("SELECT * FROM overtimes")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		ot := &Overtime{}
		err = rows.Scan(&ot.Start, &ot.End, &ot.Reason)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ot)
	}
	return ret, nil
}
