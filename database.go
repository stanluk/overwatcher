package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type WorkLog struct {
	Start, End     time.Time
	NetLen         time.Duration
	TotalLen       time.Duration
	OvertimeReason string
}

var db *sql.DB

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
		`CREATE TABLE IF NOT EXISTS worklog (start timestamp, end timestamp,
			PRIMARY KEY(start, end),
			CHECK (start<end));
		 CREATE TABLE IF NOT EXISTS overtimes (day date PRIMARY KEY, reason varchar(256));`)
	return err
}

func ShutdownSQLDb() error { return db.Close() }

func CreateWorkLog(start, end time.Time) error {
	res, err := db.Exec("INSERT INTO worklog VALUES(?, ?);", start.UTC(), end.UTC())
	if err != nil {
		return err
	}
	_, err = res.LastInsertId()
	return err
}

func UpdateOvertime(day time.Time, reason *string) error {
	if reason != nil && *reason != "" {
		_, err := db.Exec("REPLACE INTO overtimes VALUES(?, ?)", day, *reason)
		if err != nil {
			return err
		}
	}
	return nil
}

func QueryLogs(from, to time.Time) ([]*WorkLog, error) {
	return nil, nil
}
