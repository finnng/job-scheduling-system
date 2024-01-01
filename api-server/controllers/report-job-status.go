package controllers

import (
    "database/sql"
    "github.com/prometheus/client_golang/prometheus"
    "go-pg-bench/common"
    "go-pg-bench/entity"
    "log"
)

func ReportJobStatus(db *sql.DB, collector *prometheus.GaugeVec) error {
    var count int
    err := db.QueryRow(`
      SELECT COUNT(id) AS count
      FROM jobs
      WHERE status = $1`, entity.JobStatusInitialized).Scan(&count)
    if err != nil {
        return err
    }

    log.Println("Jobs in queue: ", count)
    common.CollectMetric(collector, "jobs_in_queue", float64(count))
    return nil
}
