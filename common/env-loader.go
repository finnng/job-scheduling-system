package common

import (
    "github.com/joho/godotenv"
    "log"
    "os"
    "regexp"
    "sync"
)

var loadEnvOnce sync.Once

func LoadEnv() {
    loadEnvOnce.Do(func() {
        if os.Getenv("ENV") != "" {
            // already loaded
            return
        }
        projectDirName := "go-pg-bench"
        re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
        cwd, _ := os.Getwd()
        rootPath := re.Find([]byte(cwd))

        err := godotenv.Load(string(rootPath) + "/" + "local.env")
        if err != nil {
            log.Fatal("Failed to get .env file", err)
        }
    })
}
