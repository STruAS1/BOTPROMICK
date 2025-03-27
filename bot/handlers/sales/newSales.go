package sales

import (
	"BOTPROMICK/Utilities"
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/product"
	"BOTPROMICK/db/models/user"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetAllProductsHandler(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	botCtx.UpdateUserName("sales")
	botCtx.UpdateUserLevel(0)
	products, err := product.GetProducts(db.DB)
	var rows [][]tgbotapi.InlineKeyboardButton
	if err != nil || products == nil || len(products) == 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "main")))
		text := "❌ Неизвестная ошибка"
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
	for _, product := range products {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(product.Title, fmt.Sprintf("product_%d", product.ID))))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "main")))
	text := "📚 Выберите продукт!"
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
}

type SaleData struct {
	ActiveStep       uint
	Sale             *product.Sale
	ActiveInputIndex uint
	ActiveInputValue string
	ActivePhotoIndex uint
	ActivePhotoId    string
}

func NewSaleHandler(botCtx *user.BotContext, ProductId uint) {
	state := botCtx.GetUserState()
	botCtx.UpdateUserLevel(1)
	var msgValue string
	var PhotoID string
	if botCtx.Message != nil {
		msgValue = botCtx.Message.Text
		if botCtx.Message.Photo != nil {
			PhotoID = botCtx.Message.Photo[len(botCtx.Message.Photo)-1].FileID
		}
		delMsgCfg := tgbotapi.DeleteMessageConfig{
			ChatID:    botCtx.TelegramID,
			MessageID: botCtx.Message.MessageID,
		}
		botCtx.Ctx.BotAPI.Send(delMsgCfg)
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	if botCtx.User.UserNetwork == nil || !botCtx.User.UserNetwork.CanSell {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "back")))
		text := "🫥 Вы не можете свершать продажи!"

		msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.TelegramID, state.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
		msg.DisableWebPagePreview = true
		msg.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msg)

		return
	}
	_SaleData, exist := state.Data["SaleData"].(SaleData)
	if !exist {
		if ProductId == 0 {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "back")))
			text := "❌ Неизвестная ошибка"

			msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.TelegramID, state.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)

			return
		}
		Product, err := product.GetProductBtID(*db.DB, ProductId)
		if err != nil {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "back")))
			text := "❌ Неизвестная ошибка"
			msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.TelegramID, state.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)

			return
		}
		NewSale, err := Product.NewSale(db.DB, botCtx.User.UserNetwork)
		if err != nil {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "back")))
			text := "❌ Неизвестная ошибка №0021!"
			msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.TelegramID, state.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)

			return
		}
		_SaleData = SaleData{
			Sale: NewSale,
		}
	}
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		var firstArgsInt int64
		if len(data) == 2 {
			firstArgsInt, _ = strconv.ParseInt(data[1], 10, 0)
		}
		switch _SaleData.ActiveStep {
		case 0:
			switch data[0] {
			case "addPhoto":
				text := "Отправьте фото"
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "back")))
				_SaleData.ActiveStep = 1
				state.Data["SaleData"] = _SaleData
				msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				if state.MessageID != 0 {
					delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
					botCtx.Ctx.BotAPI.Send(delCFG)
				}
				botCtx.SendMessage(msg)
				return
			case "addInput":
				if int(firstArgsInt) < len(_SaleData.Sale.InputSales) {
					_SaleData.ActiveInputIndex = uint(firstArgsInt)
					text := "Отправьте Значение для: " + _SaleData.Sale.InputSales[_SaleData.ActiveInputIndex].Title
					rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "back")))
					_SaleData.ActiveStep = 3
					state.Data["SaleData"] = _SaleData
					msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
					msg.ParseMode = "HTML"
					msg.DisableWebPagePreview = true
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
					if state.MessageID != 0 {
						delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
						botCtx.Ctx.BotAPI.Send(delCFG)
					}
					botCtx.SendMessage(msg)
				}
				return
			case "photo":
				if int(firstArgsInt) < len(_SaleData.Sale.Photos) {
					_SaleData.ActivePhotoIndex = uint(firstArgsInt)
					_SaleData.ActiveStep = 5
				} else {
					return
				}
			case "input":
				if int(firstArgsInt) < len(_SaleData.Sale.InputSales) {
					_SaleData.ActiveInputIndex = uint(firstArgsInt)
					_SaleData.ActiveStep = 6
				} else {
					return
				}
			case "backALL":
				_SaleData.ActiveStep = 7
			case "save":
				if err := _SaleData.Sale.Confirm(db.DB); err != nil {
					callback := tgbotapi.NewCallbackWithAlert(botCtx.CallbackQuery.ID, err.Error())
					callback.ShowAlert = false
					botCtx.Ctx.BotAPI.Send(callback)
					return
				}
				callback := tgbotapi.NewCallbackWithAlert(botCtx.CallbackQuery.ID, "Продажа успешно сохранена!")
				callback.ShowAlert = false
				botCtx.Ctx.BotAPI.Send(callback)
				state.MessageID = 0
				GetAllProductsHandler(botCtx)
				delete(state.Data, "SaleData")
				return
			}
		case 2:
			switch data[0] {
			case "back":
				_SaleData.ActiveStep = 0
				_SaleData.ActivePhotoId = ""
			case "save":
				_SaleData.Sale.AddPhoto(db.DB, _SaleData.ActivePhotoId)
				_SaleData.ActiveStep = 0
				_SaleData.ActivePhotoId = ""
			}
		case 4:
			switch data[0] {
			case "back":
				_SaleData.ActiveStep = 0
				_SaleData.ActiveInputValue = ""
				_SaleData.ActiveInputIndex = 0
			case "save":
				if int(_SaleData.ActiveInputIndex) < len(_SaleData.Sale.InputSales) {
					err := _SaleData.Sale.AddInputValue(db.DB, _SaleData.ActiveInputIndex, _SaleData.ActiveInputValue)
					if err != nil {
						fmt.Print(err)
					}
					_SaleData.ActiveStep = 0
					_SaleData.ActiveInputValue = ""
					_SaleData.ActiveInputIndex = 0
				} else {
					return
				}
			}
		case 5:
			switch data[0] {
			case "back":
				_SaleData.ActiveStep = 0
				_SaleData.ActivePhotoIndex = 0
			case "delete":
				if int(_SaleData.ActivePhotoIndex) < len(_SaleData.Sale.Photos) {
					_SaleData.Sale.RemovePhoto(db.DB, _SaleData.Sale.Photos[_SaleData.ActivePhotoIndex].ID)
					_SaleData.ActiveStep = 0
					_SaleData.ActivePhotoIndex = 0
				} else {
					return
				}
			}
		case 6:
			switch data[0] {
			case "back":
				_SaleData.ActiveStep = 0
				_SaleData.ActiveInputIndex = 0
			case "delete":
				if int(_SaleData.ActiveInputIndex) < len(_SaleData.Sale.InputSales) {
					_SaleData.Sale.AddInputValue(db.DB, _SaleData.ActiveInputIndex, "")
					_SaleData.ActiveStep = 0
					_SaleData.ActiveInputValue = ""
					_SaleData.ActiveInputIndex = 0
				} else {
					return
				}
			}
		case 7:
			switch data[0] {
			case "back":
				_SaleData.ActiveStep = 0
			case "Cancel":
				state.MessageID = 0
				_SaleData.Sale.Cancel(db.DB)
				GetAllProductsHandler(botCtx)
				delete(state.Data, "SaleData")
				return
			}
		}
	}
	switch _SaleData.ActiveStep {
	case 0:
		_SaleData.ActiveStep = 0
		ShortLink, err := Utilities.ShortenURL(_SaleData.Sale.GetLink())
		Text := fmt.Sprintf("🎁 Оффер: <b>%s</b>\n\n", _SaleData.Sale.Product.Title)
		Text += fmt.Sprintf("📎 <i><b>Короткая ссылка:</b></i>\n<code>%s</code>\n\n", ShortLink)
		Text += fmt.Sprintf("%s\n\n", _SaleData.Sale.Product.Description)
		Text += "<b><i>⚠️ Внимание! Неподтвержденные продажи отменяются  автоматически по истечению 30 минут с момента создания!</i></b>\n\n"
		Text += fmt.Sprintf("<i>Загружено фотографий </i>: <b>%d/%d</b>\n\n", len(_SaleData.Sale.Photos), _SaleData.Sale.Product.PhotosCount)
		if len(_SaleData.Sale.Photos) != 0 {
			for i, _ := range _SaleData.Sale.Photos {
				Title := fmt.Sprintf("Фото %d", i+1)
				CallBack := fmt.Sprintf("photo_%d", i)
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(Title, CallBack)))
			}
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить фото 📷", "addPhoto")))
		if len(_SaleData.Sale.InputSales) != 0 {
			for i, Input := range _SaleData.Sale.InputSales {
				Title := Input.Value
				CallBack := fmt.Sprintf("input_%d", i)
				if Title == "" {
					Title = Input.Title + " ❓"
					if Input.Optional {
						Title = Input.Title + " ❔"
					}
					CallBack = fmt.Sprintf("addInput_%d", i)

				} else {
					Title += " ✅"
				}
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(Title, CallBack)))
			}
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Сохранить продажу 💾", "save")))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отменить продажу ❌", "backALL")))
		photo, err := Utilities.GenerateQRCode(ShortLink)
		if err != nil {
			return
		}
		msgPhotoCfg := tgbotapi.NewPhoto(botCtx.TelegramID, tgbotapi.FileBytes{Name: "QRCODE", Bytes: photo})
		msgPhotoCfg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		msgPhotoCfg.Caption = Text
		msgPhotoCfg.ParseMode = "HTML"
		if state.MessageID != 0 {
			delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
			botCtx.Ctx.BotAPI.Send(delCFG)
		}
		msgToDel, err := botCtx.Ctx.BotAPI.Send(msgPhotoCfg)
		if err != nil {
			fmt.Print(err)
		}
		state.MessageID = msgToDel.MessageID
	case 1:
		if botCtx.CallbackQuery != nil {
			_SaleData.ActiveStep = 0
			_SaleData.ActivePhotoId = ""
			ShortLink, err := Utilities.ShortenURL(_SaleData.Sale.GetLink())
			Text := fmt.Sprintf("🎁 Оффер: <b>%s</b>\n\n", _SaleData.Sale.Product.Title)
			Text += fmt.Sprintf("📎 <i><b>Короткая ссылка:</b></i>\n<code>%s</code>\n\n", ShortLink)
			Text += fmt.Sprintf("%s\n\n", _SaleData.Sale.Product.Description)
			Text += "<b><i>⚠️ Внимание! Неподтвержденные продажи отменяются  автоматически по истечению 30 минут с момента создания!</i></b>\n\n"
			Text += fmt.Sprintf("<i>Загружено фотографий </i>: <b>%d/%d</b>\n\n", len(_SaleData.Sale.Photos), _SaleData.Sale.Product.PhotosCount)
			if len(_SaleData.Sale.Photos) != 0 {
				for i, _ := range _SaleData.Sale.Photos {
					Title := fmt.Sprintf("Фото %d", i+1)
					CallBack := fmt.Sprintf("photo_%d", i)
					rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(Title, CallBack)))
				}
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить фото 📷", "addPhoto")))
			if len(_SaleData.Sale.InputSales) != 0 {
				for i, Input := range _SaleData.Sale.InputSales {
					Title := Input.Value
					CallBack := fmt.Sprintf("input_%d", i)
					if Title == "" {
						Title = Input.Title + " ❓"
						if Input.Optional {
							Title = Input.Title + " ❔"
						}
						CallBack = fmt.Sprintf("addInput_%d", i)

					} else {
						Title += " ✅"
					}
					rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(Title, CallBack)))
				}
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Сохранить продажу 💾", "save")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отменить продажу ❌", "backALL")))
			photo, err := Utilities.GenerateQRCode(ShortLink)
			if err != nil {
				return
			}
			msgPhotoCfg := tgbotapi.NewPhoto(botCtx.TelegramID, tgbotapi.FileBytes{Name: "QRCODE", Bytes: photo})
			msgPhotoCfg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			msgPhotoCfg.Caption = Text
			msgPhotoCfg.ParseMode = "HTML"
			if state.MessageID != 0 {
				delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
				botCtx.Ctx.BotAPI.Send(delCFG)
			}
			msgToDel, _ := botCtx.Ctx.BotAPI.Send(msgPhotoCfg)
			state.MessageID = msgToDel.MessageID
		} else if PhotoID != "" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Сохранить 💾", "save")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "back")))
			msgPhotoCfg := tgbotapi.NewPhoto(botCtx.TelegramID, tgbotapi.FileID(PhotoID))
			msgPhotoCfg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			_SaleData.ActivePhotoId = PhotoID
			_SaleData.ActiveStep++
			if state.MessageID != 0 {
				delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
				botCtx.Ctx.BotAPI.Send(delCFG)
			}
			msgToDel, err := botCtx.Ctx.BotAPI.Send(msgPhotoCfg)
			if err != nil {
				fmt.Print(err)
			}
			state.MessageID = msgToDel.MessageID
		}
	case 3:
		if botCtx.CallbackQuery != nil {
			_SaleData.ActiveStep = 0
			_SaleData.ActiveInputValue = ""
			_SaleData.ActiveInputIndex = 0
			ShortLink, err := Utilities.ShortenURL(_SaleData.Sale.GetLink())
			Text := fmt.Sprintf("🎁 Оффер: <b>%s</b>\n\n", _SaleData.Sale.Product.Title)
			Text += fmt.Sprintf("📎 <i><b>Короткая ссылка:</b></i>\n<code>%s</code>\n\n", ShortLink)
			Text += fmt.Sprintf("%s\n\n", _SaleData.Sale.Product.Description)
			Text += "<b><i>⚠️ Внимание! Неподтвержденные продажи отменяются  автоматически по истечению <u>30</u> минут с момента создания!</i></b>\n\n"
			Text += fmt.Sprintf("<i>Загружено фотографий </i>: <b>%d/%d</b>\n\n", len(_SaleData.Sale.Photos), _SaleData.Sale.Product.PhotosCount)
			if len(_SaleData.Sale.Photos) != 0 {
				for i, _ := range _SaleData.Sale.Photos {
					Title := fmt.Sprintf("Фото %d", i+1)
					CallBack := fmt.Sprintf("photo_%d", i)
					rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(Title, CallBack)))
				}
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Добавить фото 📷", "addPhoto")))
			if len(_SaleData.Sale.InputSales) != 0 {
				for i, Input := range _SaleData.Sale.InputSales {
					Title := Input.Value
					CallBack := fmt.Sprintf("input_%d", i)
					if Title == "" {
						Title = Input.Title + " ❓"
						if Input.Optional {
							Title = Input.Title + " ❔"
						}
						CallBack = fmt.Sprintf("addInput_%d", i)

					} else {
						Title += " ✅"
					}
					rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(Title, CallBack)))
				}
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Сохранить продажу 💾", "save")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отменить продажу ❌", "backALL")))
			photo, err := Utilities.GenerateQRCode(ShortLink)
			if err != nil {
				return
			}
			msgPhotoCfg := tgbotapi.NewPhoto(botCtx.TelegramID, tgbotapi.FileBytes{Name: "QRCODE", Bytes: photo})
			msgPhotoCfg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			msgPhotoCfg.Caption = Text
			msgPhotoCfg.ParseMode = "HTML"
			if state.MessageID != 0 {
				delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
				botCtx.Ctx.BotAPI.Send(delCFG)
			}
			msgToDel, _ := botCtx.Ctx.BotAPI.Send(msgPhotoCfg)
			state.MessageID = msgToDel.MessageID
		} else if msgValue != "" && int(_SaleData.ActiveInputIndex) < len(_SaleData.Sale.InputSales) {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Сохранить 💾", "save")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "back")))
			text := fmt.Sprintf("<i>%s :</i>\n %s", _SaleData.Sale.InputSales[_SaleData.ActiveInputIndex].Title, msgValue)
			_SaleData.ActiveInputValue = msgValue
			_SaleData.ActiveStep++
			msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			if state.MessageID != 0 {
				delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
				botCtx.Ctx.BotAPI.Send(delCFG)
			}
			botCtx.SendMessage(msg)

		}
	case 5:
		if int(_SaleData.ActivePhotoIndex) < len(_SaleData.Sale.Photos) {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Удалить 🗑️", "delete")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "back")))
			msgPhotoCfg := tgbotapi.NewPhoto(botCtx.TelegramID, tgbotapi.FileID(_SaleData.Sale.Photos[_SaleData.ActivePhotoIndex].File_ID))
			msgPhotoCfg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			if state.MessageID != 0 {
				delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
				botCtx.Ctx.BotAPI.Send(delCFG)
			}
			msgToDel, _ := botCtx.Ctx.BotAPI.Send(msgPhotoCfg)
			state.MessageID = msgToDel.MessageID
		} else {
			return
		}
	case 6:
		if int(_SaleData.ActiveInputIndex) < len(_SaleData.Sale.InputSales) {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Удалить 🗑️", "delete")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "back")))
			text := fmt.Sprintf("<b>%s:</b>\n<i>%s</i>", _SaleData.Sale.InputSales[_SaleData.ActiveInputIndex].Title, _SaleData.Sale.InputSales[_SaleData.ActiveInputIndex].Value)
			msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			if state.MessageID != 0 {
				delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
				botCtx.Ctx.BotAPI.Send(delCFG)
			}
			botCtx.SendMessage(msg)
		} else {
			return
		}
	case 7:
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отменить ❌", "Cancel")))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "back")))
		text := "🫥 Вы верены что хотите отменить продажу?"
		msg := tgbotapi.NewMessage(botCtx.TelegramID, text)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		if state.MessageID != 0 {
			delCFG := tgbotapi.DeleteMessageConfig{ChatID: botCtx.TelegramID, MessageID: state.MessageID}
			botCtx.Ctx.BotAPI.Send(delCFG)
		}
		botCtx.SendMessage(msg)

	}
	state.Data["SaleData"] = _SaleData
}
