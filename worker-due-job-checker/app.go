package main

import (
    "fmt"
    "github.com/lib/pq"
    _ "github.com/lib/pq"
    . "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
    "time"
)

const (
    dueJobBatchSize    = 1000
    jobCheckerLockName = 1
)

func main() {
    conn := GetDBConnection()
    defer func() {
        if err := conn.Close(); err != nil {
            log.Fatal(err)
        }
    }()

    for {
        tx, err := conn.Begin()
        if err != nil {
            log.Fatal(err)
        }

        start := time.Now()
        if _, lockErr := tx.Exec(`SELECT PG_ADVISORY_XACT_LOCK($1)`, jobCheckerLockName); lockErr != nil {
            log.Fatal("Failed to acquire lock", lockErr)
        }
        log.Printf("Acquired lock time %dms", time.Since(start).Milliseconds())

        // Select and update jobs
        updateQuery := fmt.Sprintf(`
          UPDATE jobs 
          SET status = %d 
          WHERE id IN (
              SELECT id FROM jobs 
              WHERE due_at >= NOW() AND status = %d
              ORDER BY priority 
              LIMIT %d 
              FOR UPDATE SKIP LOCKED
          )
          RETURNING id`, entity.JobStatusInProgress, entity.JobStatusInitialized, dueJobBatchSize)
        rows, err := tx.Query(updateQuery)
        if err != nil {
            log.Printf("Failed to update jobs: %v\n", err)
            if rbErr := tx.Rollback(); rbErr != nil {
                log.Fatal("Failed to roll back", rbErr)
            }
            continue
        }
        log.Printf("Updated jobs time %dms", time.Since(start).Milliseconds())

        var jobIDs []int
        for rows.Next() {
            var id int
            if rsError := rows.Scan(&id); rsError != nil {
                log.Fatalf("Error scanning row: %v", rsError)
            }
            jobIDs = append(jobIDs, id)
        }
        if rCError := rows.Close(); rCError != nil {
            log.Fatal(rCError)
        }

        if commitErr := tx.Commit(); commitErr != nil {
            log.Fatal(err)
        }

        // Check if no jobs were updated, then exit the loop
        if len(jobIDs) == 0 {
            log.Println("No more jobs to process")
            time.Sleep(500 * time.Millisecond)
        } else {
            doSomething(jobIDs)
            sendJobsNextService(jobIDs)
        }

        // Log progress
        log.Printf("Processed batch of %d jobs. Total time %dms", len(jobIDs), time.Since(start).Milliseconds())
    }
}

func doSomething(jobIDs []int) {
    // log.Println("Preparing context data")
}

func sendJobsNextService(jobIDs []int) {
    var completedJobs, failedJobs []int

    for _, jobId := range jobIDs {
        err := sendMessageToQueue(jobId)
        if err != nil {
            failedJobs = append(failedJobs, jobId)
        } else {
            completedJobs = append(completedJobs, jobId)
        }
    }

    // Update completed jobs
    if len(completedJobs) > 0 {
        err := updateJobStatuses(completedJobs, entity.JobStatusCompleted)
        if err != nil {
            log.Printf("Failed to update completed jobs: %v", err)
        }
    }

    // Update failed jobs
    if len(failedJobs) > 0 {
        err := updateJobStatuses(failedJobs, entity.JobStatusFailed)
        if err != nil {
            log.Printf("Failed to update failed jobs: %v", err)
        }
    }
}

func updateJobStatuses(jobIDs []int, status entity.JobStatus) error {
    conn := GetDBConnection()
    _, err := conn.Exec(`UPDATE jobs SET status = $1 WHERE id = ANY($2)`, status, pq.Array(jobIDs))
    return err
}

func sendMessageToQueue(jobId int) error {
    return nil
}
