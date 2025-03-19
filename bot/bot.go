package bot

import (
	"BOTPROMICK/bot/handlers"
	"BOTPROMICK/config"
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/user"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot(cfg *config.Config) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}
	log.Printf("Authorized on account %s", botAPI.Self.UserName)

	Bctx := user.NewContext(botAPI, cfg)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)
	updateQueue := make(chan tgbotapi.Update, 100)

	workerCount := 50

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for update := range updateQueue {
				processUpdate(Bctx, update)
			}
		}()
	}

	for update := range updates {
		select {
		case updateQueue <- update:
		default:
			log.Println("Очередь обновлений переполнена, пропуск сообщения")
		}
	}

	close(updateQueue)
	wg.Wait()
}

func processUpdate(ctx *user.Context, update tgbotapi.Update) {
	var TelegramID int64
	var message *tgbotapi.Message
	var callbackQuery *tgbotapi.CallbackQuery
	var lastAction string
	var UserName, FirstName, LastName string
	if update.Message != nil {
		TelegramID = update.Message.Chat.ID
		message = update.Message
		UserName = update.Message.From.UserName
		FirstName = update.Message.From.FirstName
		LastName = update.Message.From.LastName
		lastAction = "message"
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Message.Chat.Type != "private" {
			return
		}
		UserName = update.CallbackQuery.From.UserName
		FirstName = update.CallbackQuery.From.FirstName
		LastName = update.CallbackQuery.From.LastName
		TelegramID = update.CallbackQuery.Message.Chat.ID
		callbackQuery = update.CallbackQuery
		lastAction = "callback"
	} else {
		return
	}
	u, err := user.GetUser(db.DB, TelegramID, UserName, FirstName, LastName)
	if err != nil {
		msg := tgbotapi.NewMessage(TelegramID, "Неизвестная ошибка!")
		ctx.BotAPI.Send(msg)
		return
	}
	botCtx := &user.BotContext{
		Ctx:           ctx,
		TelegramID:    TelegramID,
		User:          u,
		Message:       message,
		CallbackQuery: callbackQuery,
	}

	state := botCtx.GetUserState()
	state.Data["LastAction"] = lastAction

	handlers.HandleUpdate(botCtx)
}
