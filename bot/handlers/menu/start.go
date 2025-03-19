package menu

import (
	"BOTPROMICK/Utilities"
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/user"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleStartCommand(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	botCtx.UpdateUserLevel(0)
	var rows [][]tgbotapi.InlineKeyboardButton
	if botCtx.Message != nil {
		MainArgument(botCtx, botCtx.Message.CommandArguments())
	}
	if !botCtx.User.Registered {
		HandleRegister(botCtx)
		return
	}
	MainText := "<b>🤖 </b>\n\n"
	MainText += "╭━━━━━━━━━➕\n"
	MainText += fmt.Sprintf("┃  👤 <b>%s</b>\n", botCtx.User.FullName)
	MainText += fmt.Sprintf("┃  💰 <b>Баланс: <code>%s</code></b>\n", Utilities.ConvertToFancyStringFloat(fmt.Sprintf("%f", float64(botCtx.User.Balance/100))))
	if botCtx.User.UserNetwork != nil {
		network := botCtx.User.UserNetwork.Network(db.DB)
		if network == nil {
			fmt.Print("Какого хуя?")
			return
		}
		if botCtx.User.UserNetwork.Confirmed {
			MainText += fmt.Sprintf("┃  ✍️ <b>Сеть: %s</b>\n", network.Title)
			MainText += "┃━━━━━━━━━➕\n"
			MainText += "┃  <b>📊Сегодня продаж:</b>\n"
			MainText += fmt.Sprintf("┃  ⭐️ <b>На сеть:</b> <code>%s</code>\n", Utilities.ConvertToFancyString(1))
			MainText += fmt.Sprintf("┃  👀 <b>Личных:</b> <code>%s</code>\n", Utilities.ConvertToFancyString(1))
			if botCtx.User.UserNetwork.CanSell {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Новая продажа", "NewSale"),
					tgbotapi.NewInlineKeyboardButtonData("Мои продажи", "MySales"),
				))
			} else {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Мои продажи", "MySales"),
				))
			}
			var row []tgbotapi.InlineKeyboardButton
			if botCtx.User.UserNetwork.CanInviteUser {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData("Агенты сети", "NetworkAgents"))
			}
			if botCtx.User.UserNetwork.CanViewAllSales {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData("Продажи сети", "NetworkSales"))
			}
			if len(row) != 0 {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(row...))
			}
			if botCtx.User.UserNetwork.CanEditNetwork {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Изменить название сети", "NetworkSettingsName"),
				))
			}
		} else {
			MainText += "┃━━━━━━━━━➕\n"
			MainText += fmt.Sprintf("┃  ✍️ <b>Сеть: %s</b>\n", network.Title)
			MainText += "┃  <code>⭕️ Ожидайте подтерждения</code>\n"
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Отменить заявку", "Cancel"),
			))
		}
	} else {
		MainText += "┃━━━━━━━━━➕\n"
		MainText += "┃  <code>⭕️ Вы не находитесь в сети</code>\n"
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вступить в сеть", "JoinNetwork"),
			tgbotapi.NewInlineKeyboardButtonData("Создать сеть", "NewNetwork"),
		))
	}
	MainText += "╰━━━━━━━━━➕\n"
	if state.MessageID == 0 {
		msg := tgbotapi.NewMessage(botCtx.TelegramID, MainText)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		botCtx.SendMessage(msg)
	} else {
		msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.TelegramID, state.MessageID, MainText, tgbotapi.NewInlineKeyboardMarkup(rows...))
		msg.DisableWebPagePreview = true
		msg.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msg)
	}
}

func MainArgument(botCtx *user.BotContext, arg string) {
	args := strings.Split(arg, "_")
	switch args[0] {
	case "netId":
		var msgText string
		if len(args) != 2 {
			msgText = "некорректное приглашение"
			msg := tgbotapi.NewMessage(botCtx.TelegramID, msgText)
			botCtx.SendMessage(msg)
			return
		}
		bytes, err := hex.DecodeString(args[1])
		if err != nil || len(bytes) < 4 {
			msgText = "некорректное приглашение"
			msg := tgbotapi.NewMessage(botCtx.TelegramID, msgText)
			botCtx.SendMessage(msg)
			return
		}
		NetworkIdPlusBillion := binary.BigEndian.Uint32(bytes)
		if NetworkIdPlusBillion < 1_000_000_000 {
			msgText = "некорректное приглашение"
			msg := tgbotapi.NewMessage(botCtx.TelegramID, msgText)
			botCtx.SendMessage(msg)
			return
		}
		NetworkId := NetworkIdPlusBillion - 1_000_000_000
		Network := user.GetNetworkById(db.DB, uint(NetworkId))
		if Network == nil {
			msgText = "неизвестная ошибка"
			msg := tgbotapi.NewMessage(botCtx.TelegramID, msgText)
			botCtx.SendMessage(msg)
			return
		}
		if err := Network.NewUser(db.DB, botCtx.User, false); err != nil {
			msgText = err.Error()
			msg := tgbotapi.NewMessage(botCtx.TelegramID, msgText)
			botCtx.SendMessage(msg)
			return
		}
		msgText = fmt.Sprintf("Вы успешно подали заявку на вступление в сеть: %s", Network.Title)
		msg := tgbotapi.NewMessage(botCtx.TelegramID, msgText)
		botCtx.SendMessage(msg)
		return
	}
}
