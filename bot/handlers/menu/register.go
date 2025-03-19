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
		msg := tgbotapi.NewMessage(botCtx.TelegramID, "Введите ваше ФИО:")
		botCtx.SendMessage(msg)
		RegisterData.ActiveStep++
	case 1:
		if botCtx.Message != nil {
			RegisterData.FullName = Text
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "Введите ваш стаж работы (в годах):")
			botCtx.SendMessage(msg)
			RegisterData.ActiveStep++
		}
	case 2:
		var experience int
		_, err := fmt.Sscanf(Text, "%d", &experience)
		if err == nil {
			RegisterData.Experience = experience
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "Отправьте ваш номер телефона, нажав на кнопку ниже.")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButtonContact("Отправить контакт"),
				),
			)
			msg.ReplyMarkup = keyboard
			botCtx.SendMessage(msg)
			RegisterData.ActiveStep++
		} else {
			botCtx.SendMessage(tgbotapi.NewMessage(botCtx.TelegramID, "Пожалуйста, введите стаж работы числом."))
		}
	case 3:
		if botCtx.Message != nil && botCtx.Message.Contact != nil {
			RegisterData.Phone = botCtx.Message.Contact.PhoneNumber
			msg := tgbotapi.NewMessage(botCtx.TelegramID, "Теперь отправьте свою геолокацию, используя кнопку ниже.")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButtonLocation("Отправить геолокацию"),
				),
			)
			msg.ReplyMarkup = keyboard
			botCtx.SendMessage(msg)
			RegisterData.ActiveStep++
		} else {
			botCtx.SendMessage(tgbotapi.NewMessage(botCtx.TelegramID, "Пожалуйста, используйте кнопку для отправки контакта."))
		}
	case 4:
		if botCtx.Message != nil && botCtx.Message.Location != nil {
			botCtx.User.RegisterData(db.DB, RegisterData.FullName, RegisterData.Phone, RegisterData.Experience, botCtx.Message.Location.Latitude, botCtx.Message.Location.Longitude)
			HandleStartCommand(botCtx)
			delete(state.Data, "Register")
			return
		} else {
			botCtx.SendMessage(tgbotapi.NewMessage(botCtx.TelegramID, "Пожалуйста, отправьте свою геолокацию, используя кнопку."))
		}
	}
	state.Data["Register"] = RegisterData
}
