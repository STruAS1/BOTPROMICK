package menu

import (
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/user"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EditNetwork struct {
	Title      string
	ActiveStep uint
}

func EditNetworkNameHandler(botCtx *user.BotContext) {
	botCtx.UpdateUserLevel(5)
	state := botCtx.GetUserState()
	if botCtx.User.UserNetwork != nil && botCtx.User.UserNetwork.CanEditNetwork {
		network := botCtx.User.UserNetwork.Network(db.DB)
		if network == nil {
			return
		}

		NetworkData, exist := state.Data["EditNetwork"].(EditNetwork)
		if !exist {
			NetworkData = EditNetwork{}
		}
		var rows [][]tgbotapi.InlineKeyboardButton
		switch NetworkData.ActiveStep {
		case 0:
			text := "🔄️Введите новое название сети:"
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🚫 Отмена", "back")))
			if state.MessageID == 0 {
				msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				botCtx.SendMessage(msg)
			} else {
				msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.TelegramID, state.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
				msg.DisableWebPagePreview = true
				msg.ParseMode = "HTML"
				botCtx.Ctx.BotAPI.Send(msg)
			}
			NetworkData.ActiveStep++
		case 1:
			if botCtx.Message != nil && botCtx.Message.Text != "" {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✅ Сохранить", "Save")))
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🚫 Отмена", "back")))
				NetworkData.Title = botCtx.Message.Text
				text := "Новое название:\n" + botCtx.Message.Text
				msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				botCtx.SendMessage(msg)
				NetworkData.ActiveStep++
			}
		case 2:
			if botCtx.CallbackQuery != nil && botCtx.CallbackQuery.Data == "Save" {
				network.Title = NetworkData.Title
				db.DB.Save(network)
				HandleStartCommand(botCtx)
				return
			}
		}
		state.Data["EditNetwork"] = NetworkData
	}
}
