package main

import (
	"fmt"
	"time"
)

type WorkLog struct {
	enter          time.Time
	leave          time.Time
	Breaks         time.Duration
	OvertimeReason string
}

type Overtime struct {
	Start    time.Time
	Duration time.Duration
}

func NewWorkLog() *WorkLog {
	return &WorkLog{enter: time.Now(), leave: time.Now(), Breaks: 0, OvertimeReason: ""}
}

func (this *WorkLog) EnterTime() time.Time {
	return this.enter
}

func (this *WorkLog) LeaveTime() time.Time {
	return this.leave
}

func (this *WorkLog) TotalTime() time.Duration {
	ret := this.leave.Sub(this.enter) - this.Breaks
	if ret > 0 {
		return ret
	} else {
		return 0
	}
}

func (this *WorkLog) SetLeaveTime(tm time.Time) error {
	if tm.Year() == this.enter.Year() &&
		tm.Day() == this.enter.Day() &&
		tm.After(this.enter) {
		this.leave = tm
		return nil
	}
	return fmt.Errorf("Invalid date: Leave time before enter time: %q", this.enter)
}

func (this *WorkLog) SetEnterTime(tm time.Time) error {
	if tm.Year() == this.leave.Year() &&
		tm.Day() == this.leave.Day() &&
		tm.Before(this.leave) {
		this.enter = tm
		return nil
	}
	return fmt.Errorf("Invalid date: Enter time after leave time: %q", this.leave)
}

func (this *WorkLog) Overtime(workday_len time.Duration) Overtime {
	var ret Overtime

	if this.TotalTime() > workday_len {
		ret.Duration = this.TotalTime() - workday_len
		ret.Start = this.enter.Add(workday_len).Add(this.Breaks)
	}
	return ret
}
