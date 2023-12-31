package main

import (
    "go-pg-bench/entity"
    "log"
    "testing"
    "time"
)

func TestCalculateNextJobs(t *testing.T) {
    // Define a start time for the sequence
    startedAt := time.Date(2023, 12, 28, 12, 0, 0, 0, time.UTC)

    // Setup Sequence with steps
    sequence := entity.Sequence{
        Steps: []entity.Step{
            &entity.StepWaitCertainPeriod{DelayPeriod: 1, DelayUnit: entity.DelayUnitMinute},
            &entity.StepJob{Metadata: "{ 'any': 'thing' }"},
            &entity.StepWaitWeekDay{WeekDays: []entity.WeekDay{entity.Monday, entity.Tuesday, entity.Wednesday, entity.Friday}},
            &entity.StepJob{Metadata: "job 2"},
            &entity.StepWaitSpecificDate{Date: "2023-12-29T18:48:34.200Z"},
            &entity.StepJob{Metadata: "job 3"},
        },
        Subscribers: 2,
    }

    // Expected due dates for jobs
    expectedDates := []time.Time{
        startedAt.Add(1 * time.Minute),                           // 1 minute from startedAt (Job 1)
        time.Date(2023, 12, 29, 12, 1, 0, 0, time.UTC),           // Next weekday (Friday) for Job 2
        time.Date(2023, 12, 29, 18, 48, 34, 200000000, time.UTC), // Specific date for Job 3
    }

    got, err := calculateNextJobs(sequence, startedAt)
    if err != nil {
        t.Fatalf("calculateNextJobs() error = %v", err)
    }

    if len(got) != len(expectedDates) {
        t.Fatalf("Expected %d jobs, got %d", len(expectedDates), len(got))
    }

    for i, job := range got {
        log.Print(i, job)
        if !job.DueAt.Equal(expectedDates[i]) {
            t.Errorf("Job %d due at %v, want %v", i, job.DueAt, expectedDates[i])
        }
    }
}
