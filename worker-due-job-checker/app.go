package main

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/lib/pq"
    _ "github.com/lib/pq"
    "github.com/prometheus/client_golang/prometheus"
    . "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

var (
    collector = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "due_job_checker_metric_collector",
            Help: "Total delay time from the moment job was due to the moment it was taken out",
        },
        []string{"count"},
    )
)

func main() {
    LoadEnv()
    conn := GetDBConnection()
    defer conn.Close()

    prometheus.MustRegister(collector)
    dueJobBatchSize := GetEnvInt("DUE_JOB_CHECKER_BATCH_SIZE", 1000)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigs
        log.Println("Gracefully shutting down...")
        cancel()
    }()

    for {
        select {
        case <-ctx.Done():
            log.Println("Shutting down...")
            return
        default:
            start := time.Now()

            if _, lockErr := conn.Exec(`SELECT PG_ADVISORY_LOCK($1)`, GetEnvInt("JOB_CHECKER_LOCK_KEY", 1)); lockErr != nil {
                log.Println("Failed to acquire lock", lockErr)
                continue
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
            rows, err := conn.Query(updateQuery)
            if err != nil {
                log.Printf("Failed to update jobs: %v\n", err)
                continue
            }

            jobIDs := extractJobIds(rows)

            // Release lock
            if _, lockErr := conn.Exec(`SELECT PG_ADVISORY_UNLOCK($1)`, GetEnvInt("JOB_CHECKER_LOCK_KEY", 1)); lockErr != nil {
                log.Println("Failed to release lock", lockErr)
                continue
            }

            postProcessJobs(jobIDs, start)
        }
    }
}

func postProcessJobs(jobIDs []int, start time.Time) {
    processRate := float64(len(jobIDs)) / time.Now().Sub(start).Seconds()
    CollectMetric(collector, "processed_job_per_second", processRate)

    if len(jobIDs) > 0 {
        doSomething(jobIDs)
        sendJobsNextService(jobIDs)
        log.Printf("Processed batch of %d jobs. Total time %dms", len(jobIDs), time.Since(start).Milliseconds())

        // track average processing time
        avgProcessingTime := (int64)(len(jobIDs)) / time.Since(start).Milliseconds()
        CollectMetric(collector, "average_delay", (float64)(avgProcessingTime))
    }
}

func extractJobIds(rows *sql.Rows) []int {
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
    return jobIDs
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
