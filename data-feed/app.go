package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "math/rand"
    "net/http"
    "time"
)

type Job struct {
    Type     string `json:"type"`
    Metadata string `json:"metadata"`
}

type Payload struct {
    Type        string `json:"type"`
    Steps       []Job  `json:"steps"`
    Subscribers int    `json:"subscribers"`
}

func sendRequest() {
    url := "http://localhost:8081/schedule-job"
    rand.Seed(time.Now().UnixNano())

    for {
        // Create random payload
        payload := Payload{
            Type: "sequence",
            Steps: []Job{
                {
                    Type:     "job",
                    Metadata: "{ 'any': 'thing 1' }",
                },
                {
                    Type:     "job",
                    Metadata: "{ 'any': 'thing 2' }",
                },
            },
            Subscribers: rand.Intn(10000) + 1, // Random number between 1 and 1000
        }

        payloadBytes, err := json.Marshal(payload)
        if err != nil {
            fmt.Println("Error marshaling payload:", err)
            continue
        }

        // Send POST request
        resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
        if err != nil {
            fmt.Println("Error sending POST request:", err)
            continue
        }

        // Close the response body when done reading
        defer resp.Body.Close()

        // Random delay between 10ms and 1000ms
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(991)+10))

        fmt.Println("Request sent with subscribers count:", payload.Subscribers)
    }
}

func main() {
    sendRequest()
}
