package setting

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

func LoadEnv() {
    if os.Getenv("DOCKER_ENV") == "" { // kalau bukan di Docker
        if err := godotenv.Load(); err != nil {
            log.Println("⚠️ Tidak bisa load .env file, pakai environment bawaan")
        }
    }
}
