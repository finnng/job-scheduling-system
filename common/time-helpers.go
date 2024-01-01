package common

import "time"

func GetCurrentUtcTime() time.Time {
    t := time.Now()
    _, offsetInSecs := t.Zone()
    offsetInHours := offsetInSecs / 60 / 60
    utcZeroLocation, _ := time.LoadLocation("UTC")
    return t.In(utcZeroLocation).Add(time.Hour * time.Duration(offsetInHours))
}
