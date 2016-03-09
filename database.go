package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type WorkLog struct {
	Start, End     time.Time
	NetLen         time.Duration
	OvertimeReason string
}

var db *sql.DB

func (wl *WorkLog) TotalLen() time.Duration {
	return wl.End.Sub(wl.Start)
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
		`CREATE TABLE IF NOT EXISTS worklog (start timestamp, end timestamp,
			PRIMARY KEY(start, end),
			CHECK (start<=end));
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
	var ret []*WorkLog

	rows, err := db.Query(`
	SELECT strftime('%s', MIN(start)), strftime('%s', MAX(end)), SUM(strftime('%s', end) - strftime('%s', start))
	FROM worklog
	WHERE DATE(start)>=DATE(?) AND DATE(end)<=DATE(?)
	GROUP BY DATE(start), DATE(end);
	`, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var start, end, netsum int64
		err := rows.Scan(&start, &end, &netsum)
		if err != nil {
			return nil, err
		}
		wl := &WorkLog{Start: time.Unix(start, 0), End: time.Unix(end, 0), NetLen: time.Duration(netsum) * time.Second}
		ret = append(ret, wl)
	}
	return ret, nil
}
