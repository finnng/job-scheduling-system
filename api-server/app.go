package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
    "go-pg-bench/api-server/controllers"
    "go-pg-bench/common"
    "io"
    "log"
    "net/http"
    "time"
)

var (
    collector = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "api_server_metric_collector",
            Help: "Collect metric related inserting job in api server",
        },
        []string{"count"},
    )
)

var db *sql.DB

func main() {
    common.LoadEnv()
    db = common.GetDBConnection()
    defer func() {
        if err := db.Close(); err != nil {
            log.Fatal(err)
        }
    }()

    prometheus.MustRegister(collector)
    http.HandleFunc("/ping", pingHandler)
    http.HandleFunc("/schedule-job", scheduleJobHandler)

    // go func() {
    //     if err := controllers.ReportJobStatus(db, collector); err != nil {
    //         log.Print(err)
    //     }
    //     time.Sleep(7 * time.Second)
    // }()

    fmt.Println("Starting server at port 8081")
    if err := http.ListenAndServe(":8081", nil); err != nil {
        log.Fatal(err)
    }
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
    }

    if _, err := fmt.Fprintf(w, "pong"); err != nil {
        return
    }
}

func scheduleJobHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
    }

    var body controllers.ScheduleJobRequest
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer func(Body io.ReadCloser) {
        err := Body.Close()
        if err != nil {
            panic(err)
        }
    }(r.Body)

    sequence, err := controllers.ParseSequence(body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    jobs, err := controllers.CalculateNextJobs(*sequence, time.Now())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err = controllers.InsertJobs(jobs, *sequence, db, collector); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if _, err := fmt.Fprintf(w, "Jobs scheduled for sequence"); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
