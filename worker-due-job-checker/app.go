package main

import (
    "fmt"
    "github.com/lib/pq"
    _ "github.com/lib/pq"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/push"
    . "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
    "os"
    "strconv"
    "time"
)

const (
    jobCheckerLockName = 1
)

var (
    takeOutJobDelay = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "take_out_job_delay",
            Help: "Total delay time from the moment job was due to the moment it was taken out",
        },
        []string{"count"},
    )
)

func main() {
    LoadEnv()
    conn := GetDBConnection()
    defer func() {
        if err := conn.Close(); err != nil {
            log.Fatal(err)
        }
    }()
    prometheus.MustRegister(takeOutJobDelay)
    dueJobBatchSize, err := strconv.Atoi(os.Getenv("DUE_JOB_CHECKER_BATCH_SIZE"))
    if err != nil {
        log.Println("Failed to parse env value to int", err)
        dueJobBatchSize = 100000
    }

    for {
        tx, err := conn.Begin()
        if err != nil {
            log.Fatal(err)
        }

        start := time.Now()
        if _, lockErr := tx.Exec(`SELECT PG_ADVISORY_XACT_LOCK($1)`, jobCheckerLockName); lockErr != nil {
            log.Fatal("Failed to acquire lock", lockErr)
        }

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

        processRate := float64(len(jobIDs)) / time.Now().Sub(start).Seconds()

        go func(metric float64) {
            if metric == 0 {
                return
            }
            takeOutJobDelay.WithLabelValues("delayed").Set(metric)
            if err := push.New(os.Getenv("PUSH_GATEWAY_ENDPOINT"), "track_delay_time_job").Collector(takeOutJobDelay).Push(); err != nil {
                log.Println("Could not push completion time to Push gateway:", err)
            }

            log.Println("Processed jobs: ", len(jobIDs), "Rate: ", metric)
        }(processRate)

        // Check if no jobs were updated, then exit the loop
        if len(jobIDs) == 0 {
            time.Sleep(100 * time.Millisecond)
        } else {
            doSomething(jobIDs)
            sendJobsNextService(jobIDs)
            log.Printf("Processed batch of %d jobs. Total time %dms", len(jobIDs), time.Since(start).Milliseconds())
        }
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
