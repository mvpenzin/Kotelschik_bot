package main

import (
	"os"

	"gopkg.in/ini.v1"
)

type Config struct {
	Telegram TelegramConfig
	Database DatabaseConfig
	Server   ServerConfig
}

type TelegramConfig struct {
	Token string
	Debug bool
}

type DatabaseConfig struct {
	URL      string
	MaxConns int
}

type ServerConfig struct {
	Port int
}

func LoadConfig(path string) (*Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	// Telegram
	config.Telegram.Token = cfg.Section("telegram").Key("token").String()
	if config.Telegram.Token == "" {
		config.Telegram.Token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	config.Telegram.Debug = cfg.Section("telegram").Key("debug").MustBool(false)

	// Database
	config.Database.URL = cfg.Section("database").Key("url").String()
	if config.Database.URL == "" {
		config.Database.URL = os.Getenv("DATABASE_URL")
	}
	config.Database.MaxConns = cfg.Section("database").Key("max_conns").MustInt(10)

	// Server
	config.Server.Port = cfg.Section("server").Key("port").MustInt(5000)

	return config, nil
}
