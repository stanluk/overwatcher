package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type WorkLog struct {
	Day    time.Time
	Start  time.Time
	End    time.Time
	Reason string
}

var db *sql.DB

func (p *WorkLog) String() string {
	diff := p.End.Sub(p.Start)
	return fmt.Sprintf(diff.String())
}

func InitSQLDb(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path+"?parseTime=true")
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

func ShutdownSQLDb() error { return db.Close() }

func StartWork(now time.Time) error {
	res, err := db.Exec("INSERT INTO overtimes SELECT DATE(?), ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM overtimes WHERE day=DATE(?));", now, now, now, "", now)
	if err != nil {
		return err
	}
	_, err = res.LastInsertId()
	return err
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
		return fmt.Errorf(fmt.Sprintf("No work record from day %s", time.Now().String()))
	}
	return nil
}

func UpdateLog(day time.Time, start, stop *time.Time, reason *string) error {
	if reason != nil && *reason != "" {
		_, err := db.Exec("UPDATE overtimes SET reason=? WHERE day=DATE(?)", *reason, day)
		if err != nil {
			return err
		}
	}
	if start != nil {
		_, err := db.Exec("UPDATE overtimes SET start=? WHERE day=DATE(?)", *start, day)
		if err != nil {
			return err
		}
	}
	if stop != nil {
		_, err := db.Exec("UPDATE overtimes SET end=? WHERE day=DATE(?)", *stop, day)
		if err != nil {
			return err
		}
	}
	return nil
}

func QueryLogs(from, to time.Time) ([]*WorkLog, error) {
	var ret []*WorkLog
	rows, err := db.Query("SELECT date(day), time(start), time(end), reason FROM overtimes WHERE day>=DATE(?) AND day<=DATE(?)", from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		ot := &WorkLog{}
		var tms, tme, day string
		err = rows.Scan(&day, &tms, &tme, &ot.Reason)
		if err != nil {
			return nil, err
		}
		ot.Day, _ = time.Parse("2006-01-02", day)
		ot.Start, _ = time.Parse("15:04:05", tms)
		ot.End, _ = time.Parse("15:04:05", tme)
		ret = append(ret, ot)
	}
	return ret, nil
}
