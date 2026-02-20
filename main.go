package main

import (
        "log"
        "time"

        "github.com/joho/godotenv"
        "gopkg.in/ini.v1"
)

var (
        userTimers = make(map[int64]*time.Timer)
        kbTimeout  = 60
)

func main() {
        // Load .env if exists (for development)
        godotenv.Load()

        cfg, err := LoadConfig("main.ini")
        if err != nil {
                log.Printf("Failed to load main.ini: %v, using defaults/env", err)
                cfg = &Config{}
        }

        // Load timeout from ini if available
        if iniFile, err := ini.Load("main.ini"); err == nil {
                kbTimeout = iniFile.Section("settings").Key("kb_timeout").MustInt(60)
        }

        // Initialize Database
        if err := InitDB(cfg.Database); err != nil {
                log.Printf("Database initialization failed: %v. Bot functionality might be limited.", err)
        }

        // Start API Server (Frontend Dashboard) in a goroutine
        go func() {
                log.Printf("Starting API Server on port %d...", cfg.Server.Port)
                StartAPIServer(cfg.Server)
        }()

        // Start Bot
        if cfg.Telegram.Token != "" {
                log.Println("Starting Telegram Bot...")
                StartBot(cfg.Telegram)
        } else {
                log.Println("Telegram token not provided. Bot not started. Dashboard available.")
                LogBotAction("WARN", "Bot not started", "Token missing")
                // Keep process alive for API server
                select {}
        }
}
