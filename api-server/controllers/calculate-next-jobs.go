package controllers

import (
    "errors"
    "fmt"
    "go-pg-bench/entity"
    "time"
)

func CalculateNextJobs(sequence entity.Sequence, startedAt time.Time) ([]entity.Job, error) {
    jobs := make([]entity.Job, 0)
    for _, step := range sequence.Steps {
        if step.StepType() == entity.StepTypeWaitCertainPeriod {
            s := step.(*entity.StepWaitCertainPeriod)
            startedAt = startedAt.Add(time.Duration(s.DelayPeriod) * s.DelayUnit.ToDuration())
            continue
        }
        if step.StepType() == entity.StepTypeWaitWeekDay {
            s := step.(*entity.StepWaitWeekDay)
            startedAt = GetNearestWeekDay(s.WeekDays, startedAt)
            continue
        }
        if step.StepType() == entity.StepTypeWaitSpecificDate {
            s := step.(*entity.StepWaitSpecificDate)
            var err error
            startedAt, err = time.Parse(time.RFC3339, s.Date)
            if err != nil {
                return []entity.Job{}, errors.New(fmt.Sprintf("failed to parse date %v, error: %s", s.Date, err.Error()))
            }
            continue
        }

        if step.StepType() == entity.StepTypeJob {
            s := step.(*entity.StepJob)
            // schedule job at this time
            job := entity.Job{
                DueAt:    startedAt,
                Status:   entity.JobStatusInitialized,
                Metadata: s.Metadata,
                Priority: 1, // TODO: calculate priority
                TenantId: 1, // TODO: calculate tenant id
            }
            jobs = append(jobs, job)
        }
    }
    return jobs, nil
}

func GetNearestWeekDay(weekdays []entity.WeekDay, now time.Time) time.Time {
    today := int(now.Weekday())

    minDays := 7
    for _, wd := range weekdays {
        wdInt := wd.ToInt()
        daysUntil := (wdInt - today + 7) % 7
        if daysUntil == 0 { // If today is one of the specified weekdays
            return now // Return today
        }

        if daysUntil < minDays {
            minDays = daysUntil
        }
    }

    return now.AddDate(0, 0, minDays)
}
