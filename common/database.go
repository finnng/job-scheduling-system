package common

import (
    "database/sql"
    _ "github.com/lib/pq"
    "log"
    "sync"
)

var (
    db        *sql.DB
    once      sync.Once
    dbConnStr = "postgres://postgres:postgres@localhost:5432/test?sslmode=disable"
)

func GetDBConnection() *sql.DB {
    once.Do(func() {
        var err error
        db, err = sql.Open("postgres", dbConnStr)
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Successfully connected to database!")
    })
    return db
}
