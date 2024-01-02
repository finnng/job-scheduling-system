package common

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
    "log"
    "os"
    "sync"
)

var (
    db   *sql.DB
    once sync.Once
)

func GetDBConnection() *sql.DB {
    once.Do(func() {
        var err error
        db, err = sql.Open("postgres", os.Getenv("POSTGRES_CONNECTION_STRING"))
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Successfully connected to database!")

        if err = EnsureTable(db); err != nil {
            log.Fatal(err)
        }
    })
    return db
}

func EnsureTable(db *sql.DB) error {
    _, err := db.Exec(`
     CREATE TABLE IF NOT EXISTS PUBLIC.jobs
     (
         id        serial CONSTRAINT jobs_pk PRIMARY KEY,
         due_at    TIMESTAMP DEFAULT NOW() NOT NULL,
         priority  INTEGER   DEFAULT 0,
         tenant_id INTEGER   DEFAULT 1,
         status    INTEGER   DEFAULT 0,
         metadata  VARCHAR(100)
     );
     
     ALTER TABLE public.jobs
         OWNER TO postgres;
     
     CREATE INDEX IF NOT EXISTS jobs_due_at_index
         ON PUBLIC.jobs (due_at);
     
     CREATE INDEX IF NOT EXISTS jobs_priority_index
         ON PUBLIC.jobs (priority);
     
     CREATE INDEX IF NOT EXISTS jobs_status_index
         ON PUBLIC.jobs (status);
    `)
    if err != nil {
        return fmt.Errorf("error creating tables: %v", err)
    }
    log.Println("Tables ensured!")
    return nil
}
