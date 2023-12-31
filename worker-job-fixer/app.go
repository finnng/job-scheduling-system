package main

import (
    _ "github.com/lib/pq"
    . "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
    "os"
    "strconv"
    "time"
)

func main() {
    LoadEnv()
    conn := GetDBConnection()
    defer func() {
        if err := conn.Close(); err != nil {
            log.Fatal(err)
        }
    }()
    maxTimeProcessing, err := strconv.Atoi(os.Getenv("JOB_MAXIMUM_PROCESSING_TIME_IN_SECONDS"))
    if err != nil {
        log.Fatal("Failed to parse env value to int", err)
    }

    for {
        // DELETE completed jobs
        // We should archive completed jobs instead of deleting them
        // But this is testing code, so we just delete them
        delRes, err := conn.Exec(`DELETE FROM jobs WHERE status = $1`, entity.JobStatusCompleted)
        if err != nil {
            log.Fatal("Failed to delete completed jobs", err)
        }
        deleted, err := delRes.RowsAffected()
        log.Println("Deleted jobs: ", deleted, err)

        // Select job exceeding processing time limit and update them to Initialized status to get  reprocessed
        updRes, err := conn.Exec(`
          UPDATE jobs 
          SET status = $1, 
              due_at = NOW()
          WHERE id IN (
              SELECT id FROM jobs
              WHERE (status = $2 OR status = $3) 
                    AND NOW() - jobs.due_at > INTERVAL '1 second' * $4)`,
            entity.JobStatusInitialized,
            entity.JobStatusInProgress,
            entity.JobStatusFailed,
            maxTimeProcessing)
        if err != nil {
            log.Fatal("Failed to update jobs", err)
        }

        updated, err := updRes.RowsAffected()
        log.Println("Updated jobs: ", updated, err)

        log.Printf("Sleeping... for %d seconds\n", maxTimeProcessing)
        time.Sleep(time.Duration(maxTimeProcessing) * time.Second)
    }
}
