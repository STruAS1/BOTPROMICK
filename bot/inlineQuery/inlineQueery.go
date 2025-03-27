package InlineQuery

import (
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/user"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleInlineQuery(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.InlineQuery == nil {
		return
	}
	User, err := user.GetUser(db.DB, update.InlineQuery.From.ID, update.InlineQuery.From.UserName, update.InlineQuery.From.FirstName, update.InlineQuery.From.LastName)
	if err != nil {
		return
	}
	if User.UserNetwork == nil || !User.UserNetwork.CanInviteUser {
		return
	}
	network := User.UserNetwork.Network(db.DB)
	if network == nil {
		return
	}
	var results []any
	text := "Пригалшение в сеть: " + network.Title
	invite, err := network.CreateInvite(db.DB, uint32(User.ID))
	if err != nil {
		return
	}
	randomArticle := tgbotapi.InlineQueryResultArticle{
		Type:  "article",
		ID:    string(network.ID + uint(time.Now().Unix())),
		Title: "Пригласить в сеть: " + network.Title,
		InputMessageContent: tgbotapi.InputTextMessageContent{
			Text:      text,
			ParseMode: tgbotapi.ModeHTML,
		},
		ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonURL("Вступить в сеть 📥", fmt.Sprintf("https://t.me/%s?start=invite_%s", bot.Self.UserName, invite)),
				},
			},
		},
	}
	results = append(results, randomArticle)

	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
		IsPersonal:    true,
		CacheTime:     0,
	}
	bot.Send(inlineConfig)
}
