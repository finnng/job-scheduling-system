package controllers

import (
    "errors"
    "fmt"
    "go-pg-bench/entity"
    "math/rand"
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
                Priority: 0, // getPriority(), // TODO: calculate priority
                TenantId: 1, // TODO: join tenant id
            }
            jobs = append(jobs, job)
        }
    }
    return jobs, nil
}

// getPriority returns a random number between 1 and 3
// Assume we are getting this number from tenant type, 0 = new_tenants, 1 = sme, 2 = enterprise
// The priority will be managed by other service
func getPriority() int {
    return rand.Intn(3) + 1
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
