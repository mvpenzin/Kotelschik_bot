package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func InitDB(cfg DatabaseConfig) error {
	config, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return fmt.Errorf("unable to parse database config: %w", err)
	}

	config.MaxConns = int32(cfg.MaxConns)

	db, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	// Create tables if they don't exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS snt_users (
												id SERIAL PRIMARY KEY,
                        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        user_id BIGINT NOT NULL,
                        user_name VARCHAR(64) NOT NULL,
                        user_fio VARCHAR(255),
                        user_phone VARCHAR(10),
                        user_birth DATE,
                        comment TEXT
                )`,
		`CREATE INDEX IF NOT EXISTS idx_snt_users_user_name ON snt_users(user_name)`,
		`CREATE TABLE IF NOT EXISTS snt_details (
												id SERIAL PRIMARY KEY,
                        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        name VARCHAR(120) NOT NULL,
                        inn VARCHAR(10) NOT NULL,
                        kpp VARCHAR(9) NOT NULL,
                        personal_acc VARCHAR(20) NOT NULL,
                        bank_name VARCHAR(120) NOT NULL,
                        bik VARCHAR(9) NOT NULL,
                        corresp_acc VARCHAR(20) NOT NULL,
                        comment TEXT
                )`,
		`CREATE TABLE IF NOT EXISTS snt_contacts (
												id SERIAL PRIMARY KEY,
                        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        prior INTEGER NOT NULL,
                        type VARCHAR(20) NOT NULL,
                        value VARCHAR(120) NOT NULL,
                        adds VARCHAR(240),
                        comment TEXT
                )`,
		`CREATE INDEX IF NOT EXISTS idx_snt_contacts_type ON snt_contacts(type)`,
		`CREATE TABLE IF NOT EXISTS snt_logs (
												id SERIAL PRIMARY KEY,
                        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        level VARCHAR(10),
                        message TEXT,
                        details TEXT
                )`,
		`CREATE TABLE IF NOT EXISTS snt_debts (
												id SERIAL PRIMARY KEY,
                        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        area VARCHAR(12) NOT NULL,
                        debt DECIMAL(10,2) NOT NULL,
                        debt_prev DECIMAL(10,2),
                        comment TEXT
                )`,
		`CREATE TABLE IF NOT EXISTS snt_prices (
												id SERIAL PRIMARY KEY,
                        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        deposit DECIMAL(10,2) NOT NULL,
                        electricity DECIMAL(10,2) NOT NULL,
                        comment TEXT
                )`,
	}

	for _, q := range queries {
		if _, err := db.Exec(context.Background(), q); err != nil {
			log.Printf("Warning: failed to execute initialization query: %v", err)
		}
	}

	// Default record for snt_details if empty
	var count int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM snt_details").Scan(&count)
	if err == nil && count == 0 {
		_, err = db.Exec(context.Background(), `
                        INSERT INTO snt_details (name, inn, kpp, personal_acc, bank_name, bik, corresp_acc, comment)
                        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
                `, "СНТ \"КОТЕЛЬЩИК\"", "2263006486", "226301001", "40703810202140010272", "АЛТАЙСКОЕ ОТДЕЛЕНИЕ N8644 ПАО СБЕРБАНК", "040173604", "30101810200000000604", "Первоначальное значение")
		if err != nil {
			log.Printf("Warning: failed to insert default snt_details: %v", err)
		}
	}

	// Default record for snt_prices if empty
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM snt_prices").Scan(&count)
	if err == nil && count == 0 {
		_, err = db.Exec(context.Background(), `
                        INSERT INTO snt_prices (deposit, electricity, comment)
                        VALUES ($1, $2, $3)
                `, 1000, 5.04, "Первоначальные значения")
		if err != nil {
			log.Printf("Warning: failed to insert default snt_prices: %v", err)
		}
	}

	// Default records for snt_contacts if empty
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM snt_contacts").Scan(&count)
	if err == nil && count == 0 {
		_, err = db.Exec(context.Background(), `
                        INSERT INTO snt_contacts (prior, type, value, adds, comment)
                        VALUES ($1, $2, $3, $4, $5),
                               ($6, $7, $8, $9, $10),
                               ($11, $12, $13, $14, $15),
                               ($16, $17, $18, $19, $20),
                               ($21, $22, $23, $24, $25)
                     `,
			1, "Адрес", "г. Барнаул, ул. Малахова, 61", "вход с торца", "Первоначальное значение",
			2, "Председатель", "+79132296605", "Вячеслав Юрьевич", "Первоначальное значение",
			3, "Трактор", "+79132296605", "Вячеслав Юрьевич", "Первоначальное значение",
			4, "Сварщик", "+79609488031", "Евгений Владимирович", "Первоначальное значение",
			5, "Охрана", "+79635181173", nil, "Первоначальное значение",
		)
		if err != nil {
			log.Printf("Warning: failed to insert default snt_contacts: %v", err)
		}
	}

	log.Println("Connected to database and initialized tables")
	return nil
}

func LogBotAction(level, message, details string) {
	if db == nil {
		log.Println("Database not initialized, skipping log:", message)
		return
	}
	_, err := db.Exec(context.Background(), "INSERT INTO snt_logs (level, message, details) VALUES ($1, $2, $3)", level, message, details)
	if err != nil {
		log.Println("Failed to write log to DB:", err)
	}
}
