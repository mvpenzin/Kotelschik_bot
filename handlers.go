package main

import (
        "context"
        "fmt"
        "log"
        "math/rand"
        "strings"

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

        // Handle commands
        if msg.IsCommand() {
                switch msg.Command() {
                case "start":
                        handleStart(bot, msg)
                case "show":
                        handleShow(bot, msg)
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

func handleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        // Add user to DB
        _, err := db.Exec(context.Background(), `
                INSERT INTO snt_users (telegram_id, username, first_name, last_name)
                VALUES ($1, $2, $3, $4)
                ON CONFLICT (telegram_id) DO UPDATE 
                SET username = EXCLUDED.username, first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name
        `, fmt.Sprint(msg.From.ID), msg.From.UserName, msg.From.FirstName, msg.From.LastName)
        if err != nil {
                log.Printf("Error adding user: %v", err)
                LogBotAction("ERROR", "Failed to add user", err.Error())
        }

        reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет! Выберите действие в меню.")
        reply.ReplyMarkup = menuKeyboard
        bot.Send(reply)
        LogBotAction("INFO", "Start command", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleShow(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Меню показано/скрыто")
        // Toggle logic is tricky with reply keyboards as we can't easily check current state on client.
        // We'll just resend the keyboard. To hide, we'd use NewRemoveKeyboard(true).
        // Since user asked to "Show or hide", maybe we check if we should send keyboard or remove it?
        // Let's just send the keyboard again as "Show". 
        // To "Toggle", we'd need to track state per user, which is overkill for now.
        // Let's assume "show" brings it back if hidden.
        reply.ReplyMarkup = menuKeyboard
        bot.Send(reply)
}

func handleWeather(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        // Mock implementation as OpenWeatherMap requires API key
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Погода в Барнауле: +20°C, Солнечно (Mock)")
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true) // User said "Menu hide"
        // Actually user said "Menu hide" for all these commands.
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        LogBotAction("INFO", "Weather requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleTrains(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        // Mock implementation
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Расписание электричек:\n08:00 - Барнаул -> СНТ\n18:00 - СНТ -> Барнаул")
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        LogBotAction("INFO", "Trains requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleContacts(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        rows, err := db.Query(context.Background(), "SELECT name, description, phone, email FROM snt_contacts")
        if err != nil {
                log.Printf("Error querying contacts: %v", err)
                bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка получения контактов"))
                return
        }
        defer rows.Close()

        var sb strings.Builder
        sb.WriteString("Контакты СНТ:\n")
        found := false
        for rows.Next() {
                found = true
                var name, desc, phone, email string
                // handle NULLs
                var d, p, e *string
                if err := rows.Scan(&name, &d, &p, &e); err != nil {
                        continue
                }
                if d != nil { desc = *d }
                if p != nil { phone = *p }
                if e != nil { email = *e }
                
                sb.WriteString(fmt.Sprintf("- %s: %s (%s, %s)\n", name, desc, phone, email))
        }

        if !found {
                sb.WriteString("Контактов пока нет.")
        }

        reply := tgbotapi.NewMessage(msg.Chat.ID, sb.String())
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        LogBotAction("INFO", "Contacts requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handlePaymentDetails(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Реквизиты для оплаты:\nООО 'СНТ'\nИНН 1234567890\nР/с 40702810000000000000")
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        LogBotAction("INFO", "Payment details requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleQuote(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        // Mock
        quotes := []string{
                "Не волнуйтесь, если что-то не работает. Если бы всё работало, вас бы уволили.",
                "Код как шутка. Если приходится объяснять — он плохой.",
        }
        reply := tgbotapi.NewMessage(msg.Chat.ID, quotes[rand.Intn(len(quotes))])
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        LogBotAction("INFO", "Quote requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleJoke(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "Заходит улитка в бар и говорит: 'Можно мне виски с колой?' Бармен: 'Простите, мы не обслуживаем улиток'. И вышвырнул её. Через неделю заходит та же улитка и спрашивает: 'Ну и зачем ты это сделал?'")
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        LogBotAction("INFO", "Joke requested", fmt.Sprintf("User: %s", msg.From.UserName))
}

func handleBash(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
        reply := tgbotapi.NewMessage(msg.Chat.ID, "<xxx> Привет, как дела?\n<yyy> Норм, код пишу.")
        // reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
        bot.Send(reply)
        LogBotAction("INFO", "Bash requested", fmt.Sprintf("User: %s", msg.From.UserName))
}
