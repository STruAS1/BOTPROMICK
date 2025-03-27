package user

import (
	"BOTPROMICK/config"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserState struct {
	mu        sync.Mutex
	Name      string
	Level     int
	Data      map[string]any
	MessageID int
	LastSeen  time.Time
}

type Context struct {
	BotAPI     *tgbotapi.BotAPI
	UserStates sync.Map
	Config     *config.Config
}

type BotContext struct {
	Ctx           *Context
	TelegramID    int64
	User          *User
	Message       *tgbotapi.Message
	CallbackQuery *tgbotapi.CallbackQuery
}

func NewContext(botAPI *tgbotapi.BotAPI, cfg *config.Config) *Context {
	ctx := &Context{
		BotAPI: botAPI,
		Config: cfg,
	}
	go ctx.CleanupOldUsers()
	return ctx
}

func (botCtx *BotContext) GetUserState() *UserState {
	if state, exists := botCtx.Ctx.UserStates.Load(botCtx.TelegramID); exists {
		userState := state.(*UserState)
		userState.mu.Lock()
		userState.LastSeen = time.Now()
		userState.mu.Unlock()
		return userState
	}

	newState := &UserState{
		Name:      "start",
		Level:     0,
		Data:      make(map[string]any),
		MessageID: 0,
		LastSeen:  time.Now(),
	}

	botCtx.Ctx.UserStates.Store(botCtx.TelegramID, newState)
	return newState
}

func (botCtx *BotContext) UpdateUserLevel(newLevel int) {
	state := botCtx.GetUserState()
	state.mu.Lock()
	state.Level = newLevel
	state.mu.Unlock()
}

func (botCtx *BotContext) UpdateUserName(newName string) {
	if len(newName) > 50 {
		newName = newName[:50]
	}
	state := botCtx.GetUserState()
	state.mu.Lock()
	state.Name = newName
	state.Level = 0
	state.mu.Unlock()
}

func (botCtx *BotContext) ClearAllUserData() {
	state := botCtx.GetUserState()
	state.mu.Lock()
	state.Data = make(map[string]any)
	state.mu.Unlock()
}

func (botCtx *BotContext) SaveMessageID(messageID int) {
	state := botCtx.GetUserState()
	state.mu.Lock()
	state.MessageID = messageID
	state.mu.Unlock()
}

func (botCtx *BotContext) SendMessage(msg tgbotapi.MessageConfig) (int, error) {
	sentMessage, err := botCtx.Ctx.BotAPI.Send(msg)
	if err != nil {
		return 0, err
	}

	botCtx.SaveMessageID(sentMessage.MessageID)
	return sentMessage.MessageID, nil
}

func (ctx *Context) CleanupOldUsers() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		var toDelete []int64

		ctx.UserStates.Range(func(key, value any) bool {
			userID := key.(int64)
			state := value.(*UserState)

			state.mu.Lock()
			inactive := now.Sub(state.LastSeen) > 1*time.Hour
			state.mu.Unlock()

			if inactive {
				toDelete = append(toDelete, userID)
			}
			return true
		})

		log.Printf("Удаляем %d неактивных пользователей", len(toDelete))

		for _, userID := range toDelete {
			ctx.UserStates.Delete(userID)
		}
	}
}
