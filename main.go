package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Quote представляет структуру ответа от ZenQuotes API
type Quote struct {
	Q string `json:"q"` // текст цитаты
	A string `json:"a"` // автор
}

func main() {
	// Получаем токен бота из переменной окружения
	token := "7898354076:AAG5T8kdUKP2G-kV0zblHVi-XkZwTn2rvQQ"
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Создаём нового бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panicf("Ошибка создания бота: %v", err)
	}

	bot.Debug = true // полезно при разработке, в проде можно отключить
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Настройка получения обновлений (long polling)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Обрабатываем обновления в цикле
	for update := range updates {
		// Игнорируем пустые сообщения
		if update.Message == nil {
			continue
		}

		// Обрабатываем команды
		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Привет, %s! Я бот для цитат.\n"+
					"Отправь /quote, чтобы получить случайную цитату.", update.Message.From.UserName))
			bot.Send(msg)

		case "/help":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Доступные команды:\n"+
					"/start - приветствие\n"+
					"/quote - случайная цитата\n"+
					"/help - эта справка")
			bot.Send(msg)

		case "/quote":
			// Получаем цитату и отправляем
			quoteText, err := getRandomQuote()
			if err != nil {
				log.Printf("Ошибка получения цитаты: %v", err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Извините, не удалось получить цитату. Попробуйте позже.")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, quoteText)
				bot.Send(msg)
			}

		default:
			// Неизвестная команда – можно проигнорировать или ответить
			// msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Напишите /help")
			// bot.Send(msg)
		}
	}
}

// getRandomQuote делает запрос к ZenQuotes API и возвращает отформатированную цитату
func getRandomQuote() (string, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://zenquotes.io/api/random")
	if err != nil {
		return "", fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API вернул статус %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// API возвращает массив из одного объекта
	var quotes []Quote
	if err := json.Unmarshal(body, "es); err != nil {
		return "", fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	if len(quotes) == 0 {
		return "", fmt.Errorf("пустой ответ от API")
	}

	q := quotes[0]
	return fmt.Sprintf("«%s»\n— %s", q.Q, q.A), nil
}
