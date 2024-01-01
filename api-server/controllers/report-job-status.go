package controllers

import (
    "database/sql"
    "github.com/prometheus/client_golang/prometheus"
    "go-pg-bench/common"
    "go-pg-bench/entity"
)

func ReportJobStatus(db *sql.DB, collector *prometheus.GaugeVec) error {
    var count int
    err := db.QueryRow(`
      SELECT COUNT(id) AS scheduling_jobs
      FROM jobs
      WHERE status = $1`, entity.JobStatusInitialized).Scan(&count)
    if err != nil {
        return err
    }

    collector.WithLabelValues("scheduling_jobs").Set(float64(count))
    common.CollectMetric(collector, "scheduling_jobs", float64(count))
    return nil
}
