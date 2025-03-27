package menu

import (
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/user"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type JoinNetwork struct {
	InviteCode string
	ActiveStep uint
}

func HandleJoinNetwork(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	botCtx.UpdateUserLevel(4)

	var Text string
	if botCtx.Message != nil {
		Text = botCtx.Message.Text
		deleteMsg1 := tgbotapi.DeleteMessageConfig{
			ChatID:    botCtx.TelegramID,
			MessageID: botCtx.Message.MessageID,
		}
		botCtx.Ctx.BotAPI.Send(deleteMsg1)
	}

	JoinData, exist := state.Data["JoinNetwork"].(JoinNetwork)
	if !exist {
		JoinData = JoinNetwork{}
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🚀 Главное меню", "back")))
	switch JoinData.ActiveStep {
	case 0:
		text := "🔑 Введите код приглашения:"
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
		JoinData.ActiveStep++
	case 1:
		if len(Text) < 8 {
			text := "❌ <b> Некорректный код. Попробуйте ещё раз. </b>"
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
			return
		}

		bytes, err := hex.DecodeString(Text)
		if err != nil || len(bytes) < 4 {
			text := "❌ <b>Некорректный формат кода. Введите заново.</b>"
			msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			botCtx.SendMessage(msg)
			return
		}

		NetworkIdPlusBillion := binary.BigEndian.Uint32(bytes)
		if NetworkIdPlusBillion < 1_000_000_000 {
			text := "❌ <b>Некорректный код. Введите заново.</b>"
			msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			botCtx.SendMessage(msg)
			return
		}

		NetworkId := NetworkIdPlusBillion - 1_000_000_000
		Network := user.GetNetworkById(db.DB, uint(NetworkId))
		if Network == nil {
			text := "❌ Ошибка: сеть не найдена."
			msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			botCtx.SendMessage(msg)
			return
		}

		if err := Network.NewUser(db.DB, botCtx.User, false); err != nil {
			text := fmt.Sprintf("❌ Ошибка: %s", err.Error())
			msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			botCtx.SendMessage(msg)
			return
		}
		text := fmt.Sprintf("✅ Заявка на вступление в сеть \"%s\" успешно отправлена!", Network.Title)
		msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
		msg.DisableWebPagePreview = true
		msg.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msg)
		delete(state.Data, "JoinNetwork")
		HandleStartCommand(botCtx)
		return
	}

	state.Data["JoinNetwork"] = JoinData
}

func CancelToJoinNetwork(botCtx *user.BotContext) {
	if botCtx.User.UserNetwork == nil || botCtx.User.UserNetwork.Confirmed {
		return
	}
	network := botCtx.User.UserNetwork.Network(db.DB)
	if network == nil {
		return
	}
	err := network.RemoveUser(db.DB, botCtx.User, botCtx.Ctx.BotAPI, "Вы отменили заявку на вступление! 😓")
	if err != nil {
		fmt.Print(err)
	}
	HandleStartCommand(botCtx)
}
