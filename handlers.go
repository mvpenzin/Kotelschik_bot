package main

import (
        "context"
        "encoding/json"
        "fmt"
        "io"
        "log"
        "net/http"
        "strings"
        "time"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
        menuKeyboard = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                        tgbotapi.NewKeyboardButton("Прогноз погоды"),
                        tgbotapi.NewKeyboardButton("Расписание электричек"),
                ),
                tgbotapi.NewKeyboardButtonRow(
                        tgbotapi.NewKeyboardButton("Контакты"),
                        tgbotapi.NewKeyboardButton("Реквизиты для оплаты"),
                ),
                tgbotapi.NewKeyboardButtonRow(
                        tgbotapi.NewKeyboardButton("Цитату!"),
                        tgbotapi.NewKeyboardButton("Анекдот!"),
                        tgbotapi.NewKeyboardButton("Баш!"),
                ),
        )
)

func StartBot(cfg TelegramConfig) {
        bot, err := tgbotapi.NewBotAPI(cfg.Token)
        if err != nil {
                log.Fatalf("Failed to create bot: %v", err)
        }

        bot.Debug = cfg.Debug

        deleteWebhook := tgbotapi.DeleteWebhookConfig{DropPendingUpdates: false}
        if _, err := bot.Request(deleteWebhook); err != nil {
                log.Printf("Warning: failed to delete webhook: %v", err)
        }

        log.Printf("Authorized on account %s", bot.Self.UserName)
        LogBotAction("INFO", "Bot started", fmt.Sprintf("Account: %s", bot.Self.UserName))

        u := tgbotapi.NewUpdate(0)
        u.Timeout = 60

        updates := bot.GetUpdatesChan(u)

        for update := range updates {
                if update.Message == nil {
                        continue
                }

                handleUpdate(bot, update)
        }
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
        msg := update.Message
        if msg == nil {
                return
        }

        // Ensure user is in DB on ANY message
        addUserToDB(msg.From)

        // Reset/Start timer for keyboard closing
        resetUserTimer(bot, msg.Chat.ID, msg.From.ID)

        // Handle new chat members
        if msg.NewChatMembers != nil {
                for _, user := range msg.NewChatMembers {
                        if user.ID == bot.Self.ID {
                                // Bot added to group
                                reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет всем! Я бот СНТ. Готов помочь!")
                                bot.Send(reply)
                                LogBotAction("INFO", "Bot added to group", fmt.Sprintf("ChatID: %d", msg.Chat.ID))
                        } else {
                                // User added to group
                                reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Привет, @%s! Добро пожаловать! Чем могу помочь?", user.UserName))
                                bot.Send(reply)
                                LogBotAction("INFO", "User joined group", fmt.Sprintf("User: %s, ChatID: %d", user.UserName, msg.Chat.ID))
                        }
                }
                return
        }

        // Handle exit command/text
        if msg.IsCommand() && msg.Command() == "exit" || msg.Text == "Выйти" {
                handleExit(bot, msg)
                return
        }

        // Handle commands
        if msg.IsCommand() {
                switch msg.Command() {
                case "start":
                        handleStart(bot, msg)
                case "show":
                        handleShow(bot, msg)
                case "me":
                        handleUserInfo(bot, msg)
                case "fio":
                        handleFioEdit(bot, msg)
                case "phone":
                        handlePhoneEdit(bot, msg)
                default:
                        reply := tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда")
                        bot.Send(reply)
                }
                return
        }

        // Handle text messages (Menu)
        switch msg.Text {
        case "Прогноз погоды":
                handleWeather(bot, msg)
        case "Расписание электричек":
                handleTrains(bot, msg)
        case "Контакты":
                handleContacts(bot, msg)
        case "Реквизиты для оплаты":
                handlePaymentDetails(bot, msg)
        case "Цитату!":
                handleQuote(bot, msg)
        case "Анекдот!":
                handleJoke(bot, msg)
        case "Баш!":
                handleBash(bot, msg)
        }
}

func addUserToDB(user *tgbotapi.User) {
        _, err := db.Exec(context.Background(), `
                INSERT INTO snt_users (user_id, user_name)
                VALUES ($1, $2)
                ON CONFLICT (user_id) DO UPDATE 
                SET user_name = EXCLUDED.user_name, modified = CURRENT_TIMESTAMP
        `, user.ID, user.UserName)
        if err != nil {
                log.Printf("Error ensuring user exists: %v", err)
        }
}

func handleExit(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Клавиатура закрыта.")
        reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        
        // Stop timer if exists
        if timer, ok := userTimers[msg.From.ID]; ok {
                timer.Stop()
                delete(userTimers, msg.From.ID)
        }
}

func resetUserTimer(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
        if timer, ok := userTimers[userID]; ok {
                timer.Stop()
        }

        userTimers[userID] = time.AfterFunc(time.Duration(kbTimeout)*time.Second, func() {
                reply := tgbotapi.NewMessage(chatID, "Время ожидания истекло. Клавиатура закрыта.")
                reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
                bot.Send(reply)
                delete(userTimers, userID)
        })
}

func handleFioEdit(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        args := msg.CommandArguments()
        if args == "" {
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Пожалуйста, укажите ФИО после команды. Пример: /fio Иванов Иван Иванович"))
                return
        }

        _, err := db.Exec(context.Background(), `
                UPDATE snt_users 
                SET user_fio = $1, modified = CURRENT_TIMESTAMP 
                WHERE user_id = $2
        `, args, msg.From.ID)

        if err != nil {
                log.Printf("Error updating FIO: %v", err)
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при обновлении ФИО."))
                return
        }

        bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "ФИО успешно обновлено!"))
}

func handlePhoneEdit(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        args := msg.CommandArguments()
        if args == "" {
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Пожалуйста, укажите номер телефона (10 цифр) после команды. Пример: /phone 9001234567"))
                return
        }

        // Basic validation: 10 digits
        cleanPhone := ""
        for _, r := range args {
                if r >= '0' && r <= '9' {
                        cleanPhone += string(r)
                }
        }

        if len(cleanPhone) != 10 {
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Номер телефона должен содержать 10 цифр."))
                return
        }

        _, err := db.Exec(context.Background(), `
                UPDATE snt_users 
                SET user_phone = $1, modified = CURRENT_TIMESTAMP 
                WHERE user_id = $2
        `, cleanPhone, msg.From.ID)

        if err != nil {
                log.Printf("Error updating phone: %v", err)
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при обновлении телефона."))
                return
        }

        bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Телефон успешно обновлен!"))
}

func handleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        // Add user to DB (user_add logic)
        _, err := db.Exec(context.Background(), `
                INSERT INTO snt_users (user_id, user_name)
                VALUES ($1, $2)
                ON CONFLICT (user_id) DO UPDATE 
                SET user_name = EXCLUDED.user_name, modified = CURRENT_TIMESTAMP
        `, msg.From.ID, msg.From.UserName)
        if err != nil {
                log.Printf("Error adding user: %v", err)
                LogBotAction("ERROR", "Failed to add user", err.Error())
        }

        reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет! Выберите действие в меню.")
        reply.ReplyMarkup = menuKeyboard
        bot.Send(reply)
        LogBotAction("INFO", "Start command", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleUserInfo(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        var userID int64
        var userName, userFio, userPhone *string
        err := db.QueryRow(context.Background(), `
                SELECT user_id, user_name, user_fio, user_phone 
                FROM snt_users 
                WHERE user_id = $1
        `, msg.From.ID).Scan(&userID, &userName, &userFio, &userPhone)

        if err != nil {
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Информация о пользователе не найдена. Нажмите /start"))
                return
        }

        fio := "не указано"
        if userFio != nil {
                fio = *userFio
        }
        phone := "не указано"
        if userPhone != nil {
                phone = *userPhone
        }

        text := fmt.Sprintf("Ваш профиль:\nID: %d\nЛогин: @%s\nФИО: %s\nТелефон: %s",
                userID, *userName, fio, phone)
        bot.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
}

func handleShow(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Меню показано")
        reply.ReplyMarkup = menuKeyboard
        bot.Send(reply)
}

// Структура ответа Open-Meteo (только нужные поля)
type Forecast struct {
        Daily struct {
                Time             []string  `json:"time"`
                Weathercode      []int     `json:"weathercode"`
                Temperature2MMax []float64 `json:"temperature_2m_max"`
                Temperature2MMin []float64 `json:"temperature_2m_min"`
                PrecipitationSum []float64 `json:"precipitation_sum"`
                WindSpeed10MMax  []float64 `json:"wind_speed_10m_max"`
        } `json:"daily"`
}

// Русское описание кода погоды (можно дополнить)
func weatherDescription(code int) string {
        switch code {
        case 0:
                return "Ясно"
        case 1, 2, 3:
                return "Облачно"
        case 45, 48:
                return "Туман"
        case 51, 53, 55:
                return "Морось"
        case 56, 57:
                return "Ледяная морось"
        case 61, 63, 65:
                return "Дождь"
        case 66, 67:
                return "Ледяной дождь"
        case 71, 73, 75:
                return "Снег"
        case 77:
                return "Снежная крупа"
        case 80, 81, 82:
                return "Ливень"
        case 85, 86:
                return "Снегопад"
        case 95:
                return "Гроза"
        case 96, 99:
                return "Гроза с градом"
        default:
                return "Неизвестно"
        }
}

func handleWeather(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        cfg, _ := LoadConfig("main.ini")
        url := fmt.Sprintf(cfg.Weather.URL, cfg.Weather.LAT, cfg.Weather.LON, cfg.Weather.PAST, cfg.Weather.DAYS, cfg.Weather.ZONE, cfg.Weather.WIND)

        // Выполняем запрос
        resp, err := http.Get(url)
        if err != nil {
                fmt.Printf("❌ Ошибка запроса: %v\n", err)
                return
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                fmt.Printf("❌ Ошибка чтения ответа: %v\n", err)
                return
        }

        // Парсим JSON
        var f Forecast
        if err := json.Unmarshal(body, &f); err != nil {
                fmt.Printf("❌ Ошибка парсинга JSON: %v\n", err)
                return
        }

        // Выводим данные
        var sb strings.Builder
        sb.WriteString("<b><u>Погода в СНТ:</u></b>")
        for i, date := range f.Daily.Time {
                // Пропускаем, если данных меньше ожидаемого (на всякий случай)
                if i >= len(f.Daily.Weathercode) {
                        break
                }

                code := f.Daily.Weathercode[i]
                tMin := f.Daily.Temperature2MMin[i]
                tMax := f.Daily.Temperature2MMax[i]
                precip := f.Daily.PrecipitationSum[i]
                wind := f.Daily.WindSpeed10MMax[i]

                sb.WriteString(fmt.Sprintf("\n - <b>%s:</b> %s, %.1f…%.1f°C, осадки %.1fмм, ветер %.1fм/с",
                        date, weatherDescription(code), tMin, tMax, precip, wind))
        }

        reply := tgbotapi.NewMessage(msg.Chat.ID, sb.String())
        reply.ParseMode = tgbotapi.ModeHTML
        bot.Send(reply)
        LogBotAction("INFO", "Weather requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleTrains(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Расписание электричек:\n08:00 - Барнаул -> СНТ\n18:00 - СНТ -> Барнаул")
        bot.Send(reply)
        LogBotAction("INFO", "Trains requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleContacts(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        rows, err := db.Query(context.Background(), "SELECT prior, type, value, adds FROM snt_contacts ORDER BY prior ASC")
        if err != nil {
                log.Printf("Error querying contacts: %v", err)
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка получения контактов"))
                return
        }
        defer rows.Close()

        var sb strings.Builder
        sb.WriteString("<b><u>Контакты СНТ:</u></b>")
        found := false
        for rows.Next() {
                found = true
                var prior int
                var cType, value string
                var adds *string
                if err := rows.Scan(&prior, &cType, &value, &adds); err != nil {
                        continue
                }
                addInfo := ""
                if adds != nil {
                        addInfo = " (" + *adds + ")"
                }
                sb.WriteString(fmt.Sprintf("\n - <b>%s:</b> %s%s", cType, value, addInfo))
        }

        if !found {
                sb.WriteString("\nКонтактов пока нет.")
        }

        reply := tgbotapi.NewMessage(msg.Chat.ID, sb.String())
        reply.ParseMode = tgbotapi.ModeHTML
        bot.Send(reply)
        LogBotAction("INFO", "Contacts requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handlePaymentDetails(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        rows, err := db.Query(context.Background(), "SELECT name, inn, kpp, personal_acc, bank_name, bik, corresp_acc FROM snt_details ORDER BY id DESC LIMIT 1")
        if err != nil {
                log.Printf("Error querying details: %v", err)
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка получения реквизитов"))
                return
        }
        defer rows.Close()

        var sb strings.Builder
        found := false
        for rows.Next() {
                found = true
                var name, inn, kpp, acc, bank, bik, corr string
                if err := rows.Scan(&name, &inn, &kpp, &acc, &bank, &bik, &corr); err != nil {
                        continue
                }
                sb.WriteString(fmt.Sprintf("Реквизиты:\nПолучатель: %s\nИНН: %s\nКПП: %s\nСчет: %s\nБанк: %s\nБИК: %s\nКорр. счет: %s\n\n",
                        name, inn, kpp, acc, bank, bik, corr))
        }

        if !found {
                sb.WriteString("Реквизиты еще не настроены.")
        }

        bot.Send(tgbotapi.NewMessage(msg.Chat.ID, sb.String()))
        LogBotAction("INFO", "Payment details requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

type ForismaticResponse struct {
        QuoteText   string `json:"quoteText"`
        QuoteAuthor string `json:"quoteAuthor"`
}

// GetRandomQuote получает случайную цитату на указанном языке
func GetRandomQuote() (*ForismaticResponse, error) {
        cfg, _ := LoadConfig("main.ini")
        url := cfg.Quote.URL

        resp, err := http.Get(url)
        if err != nil {
                return nil, err
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                return nil, err
        }

        // API может возвращать JSON с символами экранирования, убираем их
        cleaned := strings.ReplaceAll(string(body), "\\'", "'")
        cleaned = strings.ReplaceAll(cleaned, "\\\"", "\"")
        cleaned = strings.ReplaceAll(cleaned, "\\n", "\n")

        var quote ForismaticResponse
        err = json.Unmarshal([]byte(cleaned), &quote)
        if err != nil {
                return nil, err
        }

        return &quote, nil
}

func handleQuote(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        var sb strings.Builder
        sb.WriteString("<b><u>Цитата:</u></b>")

        quote, err := GetRandomQuote()
        if err != nil {
                sb.WriteString("\nОшибка получения цитаты.")
        } else {
                if quote.QuoteAuthor == "" {
                        quote.QuoteAuthor = "Автор неизвестен"
                }
                sb.WriteString(fmt.Sprintf("\n%s", quote.QuoteText))
                sb.WriteString(fmt.Sprintf("\n<i>%s</i>", quote.QuoteAuthor))
        }

        reply := tgbotapi.NewMessage(msg.Chat.ID, sb.String())
        reply.ParseMode = tgbotapi.ModeHTML
        bot.Send(reply)
        LogBotAction("INFO", "Quote requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleJoke(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Заходит улитка в бар и говорит: 'Можно мне виски с колой?' Бармен: 'Простите, мы не обслуживаем улиток'. И вышвырнул её. Через неделю заходит та же улитка и спрашивает: 'Ну и зачем ты это сделал?'")
        bot.Send(reply)
        LogBotAction("INFO", "Joke requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleBash(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "<xxx> Привет, как дела?\n<yyy> Норм, код пишу.")
        bot.Send(reply)
        LogBotAction("INFO", "Bash requested", fmt.Sprintf("User: %s", msg.From.UserName))
}
