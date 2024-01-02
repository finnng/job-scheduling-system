package main

import (
    "context"
    "database/sql"
    "github.com/lib/pq"
    _ "github.com/lib/pq"
    "github.com/prometheus/client_golang/prometheus"
    . "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
    "os"
    "os/signal"
    "sort"
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

            rows, err := conn.Query(`
              UPDATE jobs 
              SET status = $1
              WHERE id IN (
                  SELECT id FROM jobs 
                  WHERE due_at <= NOW() AND status = $2
                  ORDER BY priority 
                  LIMIT $3
                  FOR UPDATE SKIP LOCKED
              )
              RETURNING id, due_at`, entity.JobStatusInProgress, entity.JobStatusInitialized, dueJobBatchSize)
            if err != nil {
                log.Printf("Failed to update jobs: %v\n", err)
                continue
            }

            jobs := extractJobs(rows)

            // Release lock
            if _, lockErr := conn.Exec(`SELECT PG_ADVISORY_UNLOCK($1)`, GetEnvInt("JOB_CHECKER_LOCK_KEY", 1)); lockErr != nil {
                log.Println("Failed to release lock", lockErr)
                continue
            }

            sendJobsNextService(jobs)
            collectMetrics(jobs, start)
        }
    }
}

func collectMetrics(jobs []entity.Job, start time.Time) {
    if len(jobs) == 0 {
        return
    }
    processRate := float64(len(jobs)) / time.Now().Sub(start).Seconds()
    CollectMetric(collector, "job_throughput_per_sec", processRate)

    log.Printf("Processed batch of %d jobs. Total time %dms", len(jobs), time.Since(start).Milliseconds())

    p95 := calculateP95(jobs)
    CollectMetric(collector, "job_post_process_p95", p95)
}

func calculateP95(jobs []entity.Job) float64 {
    if len(jobs) == 0 {
        // No jobs, so percentile is undefined. Handle as needed.
        return -1 // or math.NaN()
    }

    var delays []float64
    utcNow := time.Now().UTC()
    for _, job := range jobs {
        delay := utcNow.Sub(job.DueAt).Milliseconds()
        delays = append(delays, float64(delay))
    }
    // We may need to store the delays for later use, so we'll return them

    if len(delays) == 1 {
        // Only one job, so its delay is the p95
        return delays[0]
    }

    sort.Float64s(delays)

    p95Index := int(float64(len(delays)) * 0.95)
    if p95Index >= len(delays) {
        p95Index = len(delays) - 1
    }
    return delays[p95Index]
}

func extractJobs(rows *sql.Rows) []entity.Job {
    var jobs []entity.Job
    for rows.Next() {
        var id int
        var dueAt time.Time
        if rsError := rows.Scan(&id, &dueAt); rsError != nil {
            log.Fatalf("Error scanning row: %v", rsError)
        }
        job := entity.Job{
            Id:    id,
            DueAt: dueAt,
        }
        jobs = append(jobs, job)
    }
    if rCError := rows.Close(); rCError != nil {
        log.Fatal(rCError)
    }
    return jobs
}

func sendJobsNextService(jobs []entity.Job) {
    if len(jobs) == 0 {
        return
    }
    var completedJobs, failedJobs []int

    for _, job := range jobs {
        err := sendMessageToQueue(job.Id)
        if err != nil {
            failedJobs = append(failedJobs, job.Id)
        } else {
            completedJobs = append(completedJobs, job.Id)
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
    // track error rate
    CollectMetric(collector, "job_post_process_error_rate", (float64)(len(failedJobs)/len(jobs)))
}

func updateJobStatuses(jobIDs []int, status entity.JobStatus) error {
    conn := GetDBConnection()
    _, err := conn.Exec(`UPDATE jobs SET status = $1 WHERE id = ANY($2)`, status, pq.Array(jobIDs))
    return err
}

func sendMessageToQueue(jobId int) error {
    return nil
}
