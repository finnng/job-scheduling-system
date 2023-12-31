package entity

import "time"

type Job struct {
    Id       int       `json:"id"`
    DueAt    time.Time `json:"due_at"`
    Priority int       `json:"priority"`
    Status   JobStatus `json:"status"`
    Metadata string    `json:"metadata"`
    TenantId int       `json:"tenant_id"`
}

type JobStatus int

const (
    JobStatusInitialized JobStatus = iota
    JobStatusInProgress
    JobStatusCompleted
    JobStatusFailed
)

func (s JobStatus) String() string {
    if s == JobStatusInitialized {
        return "JobStatusInitialized"
    }
    if s == JobStatusInProgress {
        return "JobStatusInProgress"
    }
    if s == JobStatusCompleted {
        return "JobStatusCompleted"
    }
    if s == JobStatusFailed {
        return "JobStatusFailed"
    }
    return "JobStatusUnknown"
}
