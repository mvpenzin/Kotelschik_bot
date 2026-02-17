package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Конфигурация (лучше вынести в переменные окружения)
const (
	BotToken = "7898354076:AAG5T8kdUKP2G-kV0zblHVi-XkZwTn2rvQQ"
	//DBConnString      = "postgres://user:pass@localhost:5432/dbname"
	DBConnString      = "postgresql://postgres:password@helium/heliumdb?sslmode=disable"
	OpenWeatherAPIKey = "YOUR_OPENWEATHER_KEY"
	AdminID           = 466588600
)

// --- Менеджеры для работы с таблицами БД ---

// UserManager управляет таблицей snt_users
type UserManager struct {
	db *pgxpool.Pool
}

func (m *UserManager) Init(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS snt_users (
		created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		user_id BIGINT NOT NULL PRIMARY KEY,
		user_name VARCHAR(64) NOT NULL,
		user_fio VARCHAR(255),
		user_phone VARCHAR(10),
		comment TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_snt_users_user_name ON snt_users(user_name);
	`
	_, err := m.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create snt_users: %w", err)
	}

	// Триггер для автоматического обновления modified
	trigger := `
	CREATE OR REPLACE FUNCTION update_modified_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.modified = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_snt_users_modtime') THEN
			CREATE TRIGGER update_snt_users_modtime
				BEFORE UPDATE ON snt_users
				FOR EACH ROW
				EXECUTE FUNCTION update_modified_column();
		END IF;
	END
	$$;
	`
	_, err = m.db.Exec(ctx, trigger)
	return err
}

func (m *UserManager) AddUser(ctx context.Context, userID int64, userName string) error {
	query := `
	INSERT INTO snt_users (user_id, user_name)
	VALUES ($1, $2)
	ON CONFLICT (user_id) DO UPDATE
	SET user_name = EXCLUDED.user_name
	`
	_, err := m.db.Exec(ctx, query, userID, userName)
	return err
}

func (m *UserManager) UpdateFio(ctx context.Context, userID int64, fio string) error {
	query := `UPDATE snt_users SET user_fio = $1 WHERE user_id = $2`
	_, err := m.db.Exec(ctx, query, fio, userID)
	return err
}

func (m *UserManager) UpdatePhone(ctx context.Context, userID int64, phone string) error {
	query := `UPDATE snt_users SET user_phone = $1 WHERE user_id = $2`
	_, err := m.db.Exec(ctx, query, phone, userID)
	return err
}

func (m *UserManager) GetUserInfo(ctx context.Context, userID int64) (map[string]interface{}, error) {
	query := `SELECT user_id, user_name, user_fio, user_phone FROM snt_users WHERE user_id = $1`
	row := m.db.QueryRow(ctx, query, userID)
	var (
		uid    int64
		uname  string
		ufio   *string
		uphone *string
	)
	err := row.Scan(&uid, &uname, &ufio, &uphone)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"user_id":    uid,
		"user_name":  uname,
		"user_fio":   ufio,
		"user_phone": uphone,
	}, nil
}

// DetailsManager управляет таблицей snt_details (исправлено название)
type DetailsManager struct {
	db *pgxpool.Pool
}

func (m *DetailsManager) Init(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS snt_details (
		created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		id VARCHAR(8) NOT NULL PRIMARY KEY,
		name VARCHAR(120) NOT NULL,
		inn VARCHAR(10) NOT NULL,
		kpp VARCHAR(9) NOT NULL,
		personal_acc VARCHAR(20) NOT NULL,
		bank_name VARCHAR(120) NOT NULL,
		bik VARCHAR(9) NOT NULL,
		corresp_acc VARCHAR(20) NOT NULL,
		comment TEXT
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_snt_details_id ON snt_details(id);
	`
	_, err := m.db.Exec(ctx, query)
	if err != nil {
		return err
	}
	// Триггер для modified (аналогично UserManager)
	trigger := `
	CREATE OR REPLACE FUNCTION update_modified_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.modified = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_snt_details_modtime') THEN
			CREATE TRIGGER update_snt_details_modtime
				BEFORE UPDATE ON snt_details
				FOR EACH ROW
				EXECUTE FUNCTION update_modified_column();
		END IF;
	END
	$$;
	`
	_, err = m.db.Exec(ctx, trigger)
	return err
}

// GetAll возвращает все записи реквизитов
func (m *DetailsManager) GetAll(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := m.db.Query(ctx, `SELECT id, name, inn, kpp, personal_acc, bank_name, bik, corresp_acc, comment FROM snt_details ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var id, name, inn, kpp, personalAcc, bankName, bik, correspAcc, comment string
		err = rows.Scan(&id, &name, &inn, &kpp, &personalAcc, &bankName, &bik, &correspAcc, &comment)
		if err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"id":           id,
			"name":         name,
			"inn":          inn,
			"kpp":          kpp,
			"personal_acc": personalAcc,
			"bank_name":    bankName,
			"bik":          bik,
			"corresp_acc":  correspAcc,
			"comment":      comment,
		})
	}
	return result, nil
}

// ContactsManager управляет таблицей snt_contacts
type ContactsManager struct {
	db *pgxpool.Pool
}

func (m *ContactsManager) Init(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS snt_contacts (
		created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		prior INT NOT NULL,
		type VARCHAR(20) NOT NULL PRIMARY KEY,
		value VARCHAR(120) NOT NULL,
		adds VARCHAR(240),
		comment TEXT
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_snt_contacts_type ON snt_contacts(type);
	`
	_, err := m.db.Exec(ctx, query)
	if err != nil {
		return err
	}
	trigger := `
	CREATE OR REPLACE FUNCTION update_modified_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.modified = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_snt_contacts_modtime') THEN
			CREATE TRIGGER update_snt_contacts_modtime
				BEFORE UPDATE ON snt_contacts
				FOR EACH ROW
				EXECUTE FUNCTION update_modified_column();
		END IF;
	END
	$$;
	`
	_, err = m.db.Exec(ctx, trigger)
	return err
}

// GetAllOrdered возвращает все контакты, отсортированные по prior
func (m *ContactsManager) GetAllOrdered(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := m.db.Query(ctx, `SELECT prior, type, value, adds, comment FROM snt_contacts ORDER BY prior`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var prior int
		var typ, value, adds, comment string
		var addsPtr, commentPtr *string
		err = rows.Scan(&prior, &typ, &value, &addsPtr, &commentPtr)
		if err != nil {
			return nil, err
		}
		if addsPtr != nil {
			adds = *addsPtr
		}
		if commentPtr != nil {
			comment = *commentPtr
		}
		result = append(result, map[string]interface{}{
			"prior":   prior,
			"type":    typ,
			"value":   value,
			"adds":    adds,
			"comment": comment,
		})
	}
	return result, nil
}

// --- Структура бота ---

type Bot struct {
	api      *tgbotapi.BotAPI
	db       *pgxpool.Pool
	users    *UserManager
	details  *DetailsManager
	contacts *ContactsManager
}

func NewBot(token, connString string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	api.Debug = true // можно отключить в проде

	ctx := context.Background()
	db, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		api:      api,
		db:       db,
		users:    &UserManager{db: db},
		details:  &DetailsManager{db: db},
		contacts: &ContactsManager{db: db},
	}

	// Инициализация таблиц
	if err := bot.users.Init(ctx); err != nil {
		return nil, fmt.Errorf("users init: %w", err)
	}
	if err := bot.details.Init(ctx); err != nil {
		return nil, fmt.Errorf("details init: %w", err)
	}
	if err := bot.contacts.Init(ctx); err != nil {
		return nil, fmt.Errorf("contacts init: %w", err)
	}

	return bot, nil
}

// replyKeyboard возвращает основную клавиатуру с кнопками
func replyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := []tgbotapi.KeyboardButton{
		{Text: "Прогноз погоды"},
		{Text: "Расписание электричек"},
		{Text: "Контакты"},
		{Text: "Реквизиты для оплаты"},
		{Text: "Цитату!"},
		{Text: "Анекдот!"},
		{Text: "Баш!"},
	}
	var rows [][]tgbotapi.KeyboardButton
	for _, btn := range buttons {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(btn))
	}
	return tgbotapi.NewReplyKeyboard(rows...)
}

// removeKeyboard клавиатура для скрытия
func removeKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}

// --- Обработчики ---

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	ctx := context.Background()

	// Обработка добавления бота в группу / нового участника
	if update.MyChatMember != nil {
		b.handleMyChatMember(update.MyChatMember)
		return
	}
	if update.ChatMember != nil {
		b.handleChatMember(update.ChatMember)
		return
	}

	// Обработка сообщений
	if update.Message == nil {
		return
	}
	msg := update.Message
	chat := msg.Chat
	user := msg.From

	// Если это группа и команда /start, предлагаем перейти в личку
	if chat.IsGroup() || chat.IsSuperGroup() {
		if msg.IsCommand() && msg.Command() == "start" {
			b.sendMessage(chat.ID, "Привет! Я бот СНТ. Для работы со мной перейдите, пожалуйста, в личный чат: @", removeKeyboard())
		}
		// Другие команды в группе игнорируем (кроме /start)
		return
	}

	// Личный чат
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			b.handleStart(ctx, user, chat.ID)
		case "show":
			b.handleShow(chat.ID)
		case "admin":
			b.handleAdmin(chat.ID, user.ID)
		default:
			b.sendMessage(chat.ID, "Неизвестная команда. Используйте /start", removeKeyboard())
		}
		return
	}

	// Обработка текстовых сообщений (нажатий на кнопки)
	if msg.Text != "" {
		b.handleButton(ctx, user, chat.ID, msg.Text)
	}
}

func (b *Bot) handleStart(ctx context.Context, user *tgbotapi.User, chatID int64) {
	// Добавляем пользователя в БД
	err := b.users.AddUser(ctx, user.ID, user.UserName)
	if err != nil {
		log.Printf("Ошибка добавления пользователя %d: %v", user.ID, err)
		b.sendMessage(chatID, "Произошла ошибка при регистрации.", removeKeyboard())
		return
	}

	// Показываем меню
	msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Выберите действие:")
	msg.ReplyMarkup = replyKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleShow(chatID int64) {
	// Просто показываем клавиатуру
	msg := tgbotapi.NewMessage(chatID, "Меню открыто.")
	msg.ReplyMarkup = replyKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleAdmin(chatID int64, userID int64) {
	if userID != AdminID {
		return // никак не реагируем
	}
	b.sendMessage(chatID, "Привет, админ!", removeKeyboard())
}

func (b *Bot) handleButton(ctx context.Context, user *tgbotapi.User, chatID int64, text string) {
	var response string
	switch text {
	case "Прогноз погоды":
		response = b.getWeather()
	case "Расписание электричек":
		response = b.getTimetable()
	case "Контакты":
		response = b.getContacts(ctx)
	case "Реквизиты для оплаты":
		response = b.getDetails(ctx)
	case "Цитату!":
		response = b.getQuote()
	case "Анекдот!":
		response = b.getAnekdot()
	case "Баш!":
		response = b.getBash()
	default:
		return
	}
	// После ответа скрываем клавиатуру
	b.sendMessage(chatID, response, removeKeyboard())
}

// Вспомогательный метод отправки с клавиатурой
func (b *Bot) sendMessage(chatID int64, text string, replyMarkup interface{}) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = replyMarkup
	msg.ParseMode = "HTML"
	b.api.Send(msg)
}

// --- Обработчики событий группы ---

func (b *Bot) handleMyChatMember(update *tgbotapi.ChatMemberUpdated) {
	// Бот добавлен в чат
	if update.NewChatMember.Status == "member" && update.OldChatMember.Status == "left" {
		greeting := "Всем привет! Я бот СНТ. Чем могу помочь? Напишите мне в личку: @"
		b.sendMessage(update.Chat.ID, greeting, removeKeyboard())
	}
}

func (b *Bot) handleChatMember(update *tgbotapi.ChatMemberUpdated) {
	// Новый участник добавлен в чат
	if update.NewChatMember.Status == "member" && update.OldChatMember.Status == "left" {
		user := update.NewChatMember.User
		greeting := fmt.Sprintf("Привет, %s! Добро пожаловать в чат СНТ. Я бот, могу помочь. Напиши мне в личку: @", user.FirstName)
		b.sendMessage(update.Chat.ID, greeting, removeKeyboard())
	}
}

// --- Заглушки для внешних API (реализовать по необходимости) ---

func (b *Bot) getWeather() string {
	// Реальный вызов OpenWeatherMap
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=Barnaul,ru&units=metric&lang=ru&appid=%s", OpenWeatherAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return "Не удалось получить погоду."
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "Ошибка обработки погоды."
	}
	// Пример парсинга (упрощённо)
	main, _ := data["main"].(map[string]interface{})
	temp, _ := main["temp"].(float64)
	weather, _ := data["weather"].([]interface{})
	desc := ""
	if len(weather) > 0 {
		w := weather[0].(map[string]interface{})
		desc, _ = w["description"].(string)
	}
	return fmt.Sprintf("Погода в Барнауле: %.1f°C, %s", temp, desc)
}

func (b *Bot) getTimetable() string {
	// Заглушка, можно реализовать через API Яндекс.Расписаний или парсинг
	return "Расписание электричек временно недоступно."
}

func (b *Bot) getContacts(ctx context.Context) string {
	contacts, err := b.contacts.GetAllOrdered(ctx)
	if err != nil || len(contacts) == 0 {
		return "Контакты не найдены."
	}
	var sb strings.Builder
	for _, c := range contacts {
		sb.WriteString(fmt.Sprintf("<b>%s</b>: %s\n", c["type"], c["value"]))
		if adds, ok := c["adds"].(string); ok && adds != "" {
			sb.WriteString(fmt.Sprintf("  <i>%s</i>\n", adds))
		}
	}
	return sb.String()
}

func (b *Bot) getDetails(ctx context.Context) string {
	details, err := b.details.GetAll(ctx)
	if err != nil || len(details) == 0 {
		return "Реквизиты не найдены."
	}
	var sb strings.Builder
	for _, d := range details {
		sb.WriteString(fmt.Sprintf("🏦 <b>%s</b>\n", d["name"]))
		sb.WriteString(fmt.Sprintf("ИНН: %s\n", d["inn"]))
		sb.WriteString(fmt.Sprintf("КПП: %s\n", d["kpp"]))
		sb.WriteString(fmt.Sprintf("Счёт: %s\n", d["personal_acc"]))
		sb.WriteString(fmt.Sprintf("Банк: %s\n", d["bank_name"]))
		sb.WriteString(fmt.Sprintf("БИК: %s\n", d["bik"]))
		sb.WriteString(fmt.Sprintf("К/с: %s\n", d["corresp_acc"]))
		if d["comment"] != nil && d["comment"].(string) != "" {
			sb.WriteString(fmt.Sprintf("Комментарий: %s\n", d["comment"]))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b *Bot) getQuote() string {
	// Заглушка
	return "Цитата дня: «Программирование — это искусство заставить компьютер делать то, что нужно, а не то, что вы сказали»."
}

func (b *Bot) getAnekdot() string {
	// Заглушка
	return "Анекдот: Штирлиц шёл по коридору и вдруг услышал шаги сзади. «За мной следят», — подумал Штирлиц и ускорил шаг. Шаги тоже ускорились. Тогда Штирлиц побежал. Шаги тоже побежали. Тогда Штирлиц остановился и закричал: «Кто здесь?». В ответ тишина. Тогда Штирлиц закурил и пошёл дальше. А сзади шли его шаги."
}

func (b *Bot) getBash() string {
	// Заглушка
	return "Цитата с Баша: – У вас есть план Б? – У нас есть план «Бля буду»."
}

// --- main ---

func main() {
	bot, err := NewBot(BotToken, DBConnString)
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	log.Printf("Бот @%s запущен", bot.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.api.GetUpdatesChan(u)

	for update := range updates {
		bot.handleUpdate(update)
	}
}
