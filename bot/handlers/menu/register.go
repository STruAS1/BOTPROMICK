package menu

import (
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/user"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Register struct {
	FullName   string
	Experience int
	Phone      string
	Latitude   float32
	Longitude  float32
	City       string
	ActiveStep uint
}

func HandleRegister(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	botCtx.UpdateUserLevel(1)

	var Text string
	if botCtx.Message != nil {
		Text = botCtx.Message.Text
	}

	RegisterData, exist := state.Data["Register"].(Register)
	if !exist {
		RegisterData = Register{}
	}

	switch RegisterData.ActiveStep {
	case 0:
		msg := tgbotapi.NewMessage(botCtx.TelegramID, "👤 Введите ваше ФИО:")
		botCtx.Ctx.BotAPI.Send(msg)
		RegisterData.ActiveStep++
	case 1:
		if botCtx.Message != nil {
			RegisterData.FullName = Text
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "⌛<b>Введите ваш стаж работы (в годах): </b>")
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)
			RegisterData.ActiveStep++
		}
	case 2:
		var experience int
		_, err := fmt.Sscanf(Text, "%d", &experience)
		if err == nil {
			RegisterData.Experience = experience
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "📱<b>Подтвердите ваш номер телефона, нажав на кнопку ниже.</b>")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButtonContact("Подтвердить ✅"),
				),
			)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = keyboard
			botCtx.Ctx.BotAPI.Send(msg)
			RegisterData.ActiveStep++
		} else {
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "⌛<b>Введите ваш стаж работы (в годах): </b>")
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)
		}
	case 3:
		if botCtx.Message != nil && botCtx.Message.Contact != nil {
			RegisterData.Phone = botCtx.Message.Contact.PhoneNumber
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "<b>Теперь отправьте свою геолокацию, используя кнопку ниже.</b>")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButtonLocation("Отправить геолокацию 🗺️"),
				),
			)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = keyboard
			botCtx.Ctx.BotAPI.Send(msg)
			RegisterData.ActiveStep++
		} else {
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "❌ Пожалуйста, подтвердите кнопку для отправки контакта.")
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)
		}
	case 4:
		if botCtx.Message != nil && botCtx.Message.Location != nil {
			botCtx.User.RegisterData(db.DB, RegisterData.FullName, RegisterData.Phone, RegisterData.Experience, botCtx.Message.Location.Latitude, botCtx.Message.Location.Longitude)
			HandleStartCommand(botCtx)
			delete(state.Data, "Register")
			return
		} else {
			botCtx.Ctx.BotAPI.Send(tgbotapi.NewMessage(botCtx.TelegramID, "❌ Пожалуйста, отправьте свою геолокацию, используя кнопку."))
		}
	}
	state.Data["Register"] = RegisterData
}
