package entity

import "time"

type Step interface {
    StepType() StepType
}

type DelayUnit string

const (
    DelayUnitMinute DelayUnit = "minute"
    DelayUnitHour   DelayUnit = "hour"
    DelayUnitDay    DelayUnit = "day"
)

type StepType string

const (
    StepTypeWaitCertainPeriod = "wait_certain_period"
    StepTypeWaitWeekDay       = "wait_weekday"
    StepTypeWaitSpecificDate  = "wait_specific_date"
    StepTypeJob               = "job"
)

type StepWaitCertainPeriod struct {
    DelayPeriod int       `json:"delay_period"`
    DelayUnit   DelayUnit `json:"delay_unit"`
}

func (s StepWaitCertainPeriod) StepType() StepType {
    return StepTypeWaitCertainPeriod
}
func (s DelayUnit) ToDuration() time.Duration {
    switch s {
    case DelayUnitMinute:
        return time.Minute
    case DelayUnitHour:
        return time.Hour
    case DelayUnitDay:
        return time.Hour * 24
    }
    return 0
}

type StepWaitSpecificDate struct {
    Date string `json:"date"`
}

func (s StepWaitSpecificDate) StepType() StepType {
    return StepTypeWaitSpecificDate
}

type StepJob struct {
    Metadata string `json:"metadata,omitempty"`
}

func (s StepJob) StepType() StepType {
    return StepTypeJob
}

type StepWaitWeekDay struct {
    WeekDays []WeekDay `json:"weekdays"`
}

func (s StepWaitWeekDay) StepType() StepType {
    return StepTypeWaitWeekDay
}

type WeekDay string

const (
    Monday    WeekDay = "monday"
    Tuesday   WeekDay = "tuesday"
    Wednesday WeekDay = "wednesday"
    Thursday  WeekDay = "thursday"
    Friday    WeekDay = "friday"
    Saturday  WeekDay = "saturday"
    Sunday    WeekDay = "sunday"
)

func (w *WeekDay) ToInt() int {
    switch *w {
    case Monday:
        return 1
    case Tuesday:
        return 2
    case Wednesday:
        return 3
    case Thursday:
        return 4
    case Friday:
        return 5
    case Saturday:
        return 6
    case Sunday:
        return 0
    }
    return 0
}
