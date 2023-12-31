package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/push"
    "go-pg-bench/common"
    "go-pg-bench/entity"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

var (
    insertJobCounter = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "inserted_jobs_total",
            Help: "Total number of jobs inserted",
        },
        []string{"count"},
    )
)

var db *sql.DB

func main() {
    common.LoadEnv()
    db = common.GetDBConnection()

    prometheus.MustRegister(insertJobCounter)
    http.HandleFunc("/ping", pingHandler)
    http.HandleFunc("/schedule-job", scheduleJobHandler)

    fmt.Println("Starting server at port 8081")
    if err := http.ListenAndServe(":8081", nil); err != nil {
        log.Fatal(err)
    }

    defer func() {
        if err := db.Close(); err != nil {
            log.Fatal(err)
        }
    }()
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

type ScheduleJobRequest struct {
    Steps       []map[string]interface{} `json:"steps"`
    Subscribers int                      `json:"subscribers"`
}

func scheduleJobHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
    }

    var body ScheduleJobRequest
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

    sequence, err := parseSequence(body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    log.Println("Sequence steps: ", sequence.Steps)

    jobs, err := calculateNextJobs(*sequence, time.Now())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    log.Println("Jobs: ", jobs)

    if err = insertJobs(jobs, *sequence); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if _, err := fmt.Fprintf(w, "Jobs scheduled for sequence"); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func parseSequence(body ScheduleJobRequest) (*entity.Sequence, error) {
    sequence := entity.Sequence{
        Subscribers: body.Subscribers,
        Steps:       []entity.Step{},
    }
    for _, stepInterface := range body.Steps {
        step, err := unmarshalStep(stepInterface)
        if err != nil {
            return &entity.Sequence{}, err
        }
        sequence.Steps = append(sequence.Steps, step)
    }

    return &sequence, nil
}

func unmarshalStep(stepInterface interface{}) (entity.Step, error) {
    stepMap, ok := stepInterface.(map[string]interface{})
    if !ok {
        return nil, errors.New("invalid step format")
    }

    stepType, ok := stepMap["type"].(string)
    if !ok {
        return nil, errors.New("step type is not a string")
    }

    var step entity.Step
    var err error
    jsonData, err := json.Marshal(stepMap)
    if err != nil {
        return nil, err
    }

    switch stepType {
    case "wait_certain_period":
        step = &entity.StepWaitCertainPeriod{}
    case "wait_weekday":
        step = &entity.StepWaitWeekDay{}
    case "wait_specific_date":
        step = &entity.StepWaitSpecificDate{}
    case "job":
        step = &entity.StepJob{}
    default:
        return nil, fmt.Errorf("unsupported step type: %s", stepType)
    }

    if err = json.Unmarshal(jsonData, step); err != nil {
        return nil, err
    }

    return step, nil
}

func calculateNextJobs(sequence entity.Sequence, startedAt time.Time) ([]entity.Job, error) {
    jobs := make([]entity.Job, 0)
    for _, step := range sequence.Steps {
        if step.StepType() == entity.StepTypeWaitCertainPeriod {
            s := step.(*entity.StepWaitCertainPeriod)
            startedAt = startedAt.Add(time.Duration(s.DelayPeriod) * s.DelayUnit.ToDuration())
            continue
        }
        if step.StepType() == entity.StepTypeWaitWeekDay {
            s := step.(*entity.StepWaitWeekDay)
            startedAt = getNearestWeekDay(s.WeekDays, startedAt)
            continue
        }
        if step.StepType() == entity.StepTypeWaitSpecificDate {
            s := step.(*entity.StepWaitSpecificDate)
            var err error
            startedAt, err = time.Parse(time.RFC3339, s.Date)
            if err != nil {
                return []entity.Job{}, errors.New(fmt.Sprintf("failed to parse date %v, error: %s", s.Date, err.Error()))
            }
            continue
        }

        if step.StepType() == entity.StepTypeJob {
            s := step.(*entity.StepJob)
            // schedule job at this time
            job := entity.Job{
                DueAt:    startedAt,
                Status:   entity.JobStatusInitialized,
                Metadata: s.Metadata,
                Priority: 1, // TODO: calculate priority
                TenantId: 1, // TODO: calculate tenant id
            }
            jobs = append(jobs, job)
        }
    }
    return jobs, nil
}

func getNearestWeekDay(weekdays []entity.WeekDay, now time.Time) time.Time {
    today := int(now.Weekday())

    minDays := 7
    for _, wd := range weekdays {
        wdInt := wd.ToInt()
        daysUntil := (wdInt - today + 7) % 7
        if daysUntil == 0 { // If today is one of the specified weekdays
            return now // Return today
        }

        if daysUntil < minDays {
            minDays = daysUntil
        }
    }

    return now.AddDate(0, 0, minDays)
}

func insertJobs(jobTemplates []entity.Job, sequence entity.Sequence) error {
    start := time.Now()
    jobTemplateCount := len(jobTemplates)
    if jobTemplateCount == 0 || sequence.Subscribers == 0 {
        log.Println("No jobs to insert or no subscribers")
        return nil
    }

    totalJobs := jobTemplateCount * sequence.Subscribers
    log.Printf("Inserting (%d) jobTemplates * total subscribers (%d) = (%d) jobs\n", jobTemplateCount, sequence.Subscribers, totalJobs)

    insertJobBatchSize, err := strconv.Atoi(os.Getenv("INSERT_JOB_BATCH_SIZE"))
    if err != nil {
        panic(err)
    }

    for batchSizeIndex := 0; batchSizeIndex < totalJobs; batchSizeIndex += insertJobBatchSize {
        endBatchIndex := min(batchSizeIndex+insertJobBatchSize, totalJobs)

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
            placeholderStartIndex := (batchItemIndex-batchSizeIndex)*4 + 1
            placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)",
                placeholderStartIndex, placeholderStartIndex+1, placeholderStartIndex+2, placeholderStartIndex+3))

            // Append job details to args slice for query execution
            args = append(args, job.DueAt, job.Status, job.Priority, job.Metadata)
        }

        query.WriteString(strings.Join(placeholders, ", "))
        finalQuery := query.String()

        // log.Println(finalQuery, args)
        _, err := db.Exec(finalQuery, args...)
        if err != nil {
            log.Printf("Failed to insert batch: %v\n", err)
            return err
        }

        insertRate := float64(endBatchIndex-batchSizeIndex) / time.Now().Sub(start).Seconds()
        log.Println("Insert rate: ", insertRate)
        insertJobCounter.WithLabelValues("new_inserted").Set(insertRate)
        // Push metrics to PushGateway
        if err := push.New(os.Getenv("PUSH_GATEWAY_ENDPOINT"), "track_insert_rate_job").Collector(insertJobCounter).Push(); err != nil {
            log.Println("Could not push completion time to Push gateway:", err)
        }
    }
    return nil
}
