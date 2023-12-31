package main

import (
    "go-pg-bench/entity"
    "testing"
    "time"
)

func TestGetNearestWeekDay(t *testing.T) {
    // Mock a specific date for consistent testing
    now := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC) // December 31, 2023 is a Sunday

    tests := []struct {
        name         string
        weekdays     []entity.WeekDay
        now          time.Time
        expectedDate time.Time
    }{
        {
            name:         "Next Monday from Sunday",
            weekdays:     []entity.WeekDay{entity.Monday},
            now:          now,
            expectedDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), // Next Monday
        },
        {
            name:         "Next Friday from Sunday",
            weekdays:     []entity.WeekDay{entity.Friday},
            now:          now,
            expectedDate: time.Date(2024, time.January, 5, 0, 0, 0, 0, time.UTC), // Next Friday
        },
        {
            name:         "Next Sunday from Sunday (Same Day)",
            weekdays:     []entity.WeekDay{entity.Sunday},
            now:          now,
            expectedDate: now, // Today (Sunday)
        },
        {
            name:         "Next Wednesday from Sunday",
            weekdays:     []entity.WeekDay{entity.Wednesday},
            now:          now,
            expectedDate: time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC), // Next Wednesday
        },
        {
            name:         "Next Weekday from Sunday (Multiple Options)",
            weekdays:     []entity.WeekDay{entity.Wednesday, entity.Friday, entity.Monday},
            now:          now,
            expectedDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), // Next Monday
        },
        {
            name:         "Next Tuesday from Sunday",
            weekdays:     []entity.WeekDay{entity.Tuesday},
            now:          now,
            expectedDate: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC), // Next Tuesday
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := getNearestWeekDay(tt.weekdays, tt.now)
            if !got.Equal(tt.expectedDate) {
                t.Errorf("getNearestWeekDay() got %v, want %v", got, tt.expectedDate)
            }
        })
    }
}
