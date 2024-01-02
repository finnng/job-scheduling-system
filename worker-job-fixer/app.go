package main

import (
    "fmt"
    _ "github.com/lib/pq"
    . "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
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
    m := GetEnvInt("JOB_MAXIMUM_PROCESSING_TIME_IN_SECONDS", 15)
    maxTimeProcessing := fmt.Sprintf("%d seconds", m)

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
        // NOW() is at utc already
        query := fmt.Sprintf(`
          UPDATE jobs 
          SET status = $1, 
              due_at = NOW()
          WHERE id IN (
              SELECT id FROM jobs
              WHERE (status = $2 AND NOW() - jobs.due_at > INTERVAL '%s') 
                OR status = $3
          )`, maxTimeProcessing) // Use string formatting to include the interval in the query

        updRes, err := conn.Exec(query,
            entity.JobStatusInitialized,
            entity.JobStatusInProgress,
            entity.JobStatusFailed,
        )
        if err != nil {
            log.Fatal("Failed to update jobs", err)
        }

        updated, err := updRes.RowsAffected()
        log.Println("Updated jobs: ", updated, err)

        log.Printf("Sleeping... for %s seconds\n", maxTimeProcessing)
        time.Sleep(time.Duration(m) * time.Second)
    }
}
