package controllers

import (
    "database/sql"
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
    "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
    "os"
    "strconv"
    "strings"
    "time"
)

func InsertJobs(jobTemplates []entity.Job, sequence entity.Sequence, db *sql.DB, collector *prometheus.GaugeVec) error {
    start := time.Now()
    jobTemplateCount := len(jobTemplates)
    if jobTemplateCount == 0 || sequence.Subscribers == 0 {
        log.Println("No jobs to insert or no subscribers")
        return nil
    }

    totalJobs := jobTemplateCount * sequence.Subscribers
    log.Printf("Inserting (%d) jobTemplates * total subscribers (%d) = (%d) jobs\n", jobTemplateCount, sequence.Subscribers, totalJobs)

    maxParams, err := strconv.Atoi(os.Getenv("POSTGRES_SUPPORTED_BATCH_PARAMETERS"))
    insertParamsCount := 4 // according to the number of column in query below
    batchSize := maxParams / insertParamsCount
    if err != nil {
        panic(err)
    }

    for batchSizeIndex := 0; batchSizeIndex < totalJobs; batchSizeIndex += batchSize {
        endBatchIndex := min(batchSizeIndex+batchSize, totalJobs)

        var query strings.Builder
        query.WriteString("INSERT INTO jobs (due_at, status, priority, metadata) VALUES ")

        var placeholders []string
        var args []interface{}

        // Create a placeholder for each job and append its values to the args slice
        for batchItemIndex := batchSizeIndex; batchItemIndex < endBatchIndex; batchItemIndex++ {
            // Cycle through the jobs array for each subscriber
            jobIndex := batchItemIndex % jobTemplateCount
            job := jobTemplates[jobIndex]

            // Calculate placeholder indexes for SQL query
            placeholderStartIndex := (batchItemIndex-batchSizeIndex)*insertParamsCount + 1
            placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)",
                placeholderStartIndex, placeholderStartIndex+1, placeholderStartIndex+2, placeholderStartIndex+3))

            // Append job details to args slice for query execution
            args = append(args, job.DueAt, job.Status, job.Priority, job.Metadata)
        }

        query.WriteString(strings.Join(placeholders, ", "))
        finalQuery := query.String()

        _, err := db.Exec(finalQuery, args...)
        if err != nil {
            log.Printf("Failed to insert batch: %v\n", err)
            return err
        }

        insertRate := float64(endBatchIndex-batchSizeIndex) / time.Now().Sub(start).Seconds()
        common.CollectMetric(collector, "new_job_inserted_rate", insertRate)
    }
    return nil
}
