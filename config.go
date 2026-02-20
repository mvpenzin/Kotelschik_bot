package main

import (
	"gopkg.in/ini.v1"
)

type Config struct {
	Settings SettingsConfig
	Telegram TelegramConfig
	Database DatabaseConfig
	Server   ServerConfig
	Weather  WeatherConfig
	Quote    QuoteConfig
}

type SettingsConfig struct {
	Timeout int
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

type WeatherConfig struct {
	URL  string
	LAT  float64
	LON  float64
	PAST int
	DAYS int
	ZONE string
	WIND string
}

type QuoteConfig struct {
	URL string
}

func LoadConfig(path string) (*Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	// Settings
	config.Settings.Timeout = cfg.Section("settings").Key("timeout").MustInt(60)

	// Telegram
	config.Telegram.Token = cfg.Section("telegram").Key("token").String()
	config.Telegram.Debug = cfg.Section("telegram").Key("debug").MustBool(false)

	// Database
	config.Database.URL = cfg.Section("database").Key("url").String()
	config.Database.MaxConns = cfg.Section("database").Key("max_conns").MustInt(10)

	// Server
	config.Server.Port = cfg.Section("server").Key("port").MustInt(5000)

	// Weather
	config.Weather.URL = cfg.Section("weather").Key("url").String()
	config.Weather.LAT = cfg.Section("weather").Key("lat").MustFloat64(53.327935)
	config.Weather.LON = cfg.Section("weather").Key("lon").MustFloat64(84.102975)
	config.Weather.PAST = cfg.Section("weather").Key("past").MustInt(1)
	config.Weather.DAYS = cfg.Section("weather").Key("days").MustInt(3)
	config.Weather.ZONE = cfg.Section("weather").Key("zone").String()
	config.Weather.WIND = cfg.Section("weather").Key("wind").String()

	// Quote
	config.Quote.URL = cfg.Section("quote").Key("url").String()

	return config, nil
}
