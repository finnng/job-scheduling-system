package common

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/push"
    "log"
    "os"
)

func CollectMetric(c *prometheus.GaugeVec, metricName string, value float64) {
    go func(name string, value float64) {
        c.WithLabelValues(name).Set(value)
        if err := push.New(os.Getenv("PUSH_GATEWAY_ENDPOINT"), "job_"+metricName).Collector(c).Push(); err != nil {
            log.Println("Could not push completion time to Push gateway:", err)
        }
    }(metricName, value)
}
