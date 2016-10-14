package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var db *sql.DB
var date_format = "2006-01-02"

func InitDb(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path+"?parseTime=true")
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS worklog (local_date string PRIMARY KEY, enter timestamp,
		leave timestamp, extra INT DEFAULT 0, reason varchar(256));
		`)
	return err
}

func ShutdownDb() error { return db.Close() }

func StoreWorkLog(wl *WorkLog) error {
	res, err := db.Exec(`
			INSERT OR IGNORE INTO worklog VALUES(?,?,?,?,?);
			UPDATE worklog
			SET enter=?, leave=?, extra=?, reason=? WHERE local_date=?;`,
		wl.EnterTime().Format(date_format), wl.EnterTime(), wl.LeaveTime(), wl.Breaks, wl.OvertimeReason,
		wl.EnterTime(), wl.LeaveTime(), wl.Breaks, wl.OvertimeReason, wl.EnterTime().Format(date_format))
	if err != nil {
		return fmt.Errorf("StoreWorkLog failed: ", err)
	}
	_, err = res.LastInsertId()
	return err
}

func QueryWorkLogs(from, to time.Time) ([]*WorkLog, error) {
	var ret []*WorkLog

	rows, err := db.Query(`
			SELECT strftime('%s', enter), strftime('%s', leave), extra, reason FROM worklog
			WHERE enter>=? AND enter<=?
			ORDER BY enter;
		`, from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		wl := WorkLog{}
		var l, e int64
		err := rows.Scan(&e, &l, &wl.Breaks, &wl.OvertimeReason)
		if err != nil {
			return nil, err
		}
		wl.enter = time.Unix(e, 0).Local()
		wl.leave = time.Unix(l, 0).Local()
		ret = append(ret, &wl)
	}
	return ret, nil
}

func QueryWorkLog(day time.Time) (*WorkLog, error) {
	y, m, d := day.Date()
	rounded := time.Date(y, m, d, 0, 0, 0, 0, day.Location())
	logs, err := QueryWorkLogs(rounded, rounded.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, nil
	}
	if len(logs) > 1 {
		panic("Invalid worklogs count.")
	}
	return logs[0], nil

}
