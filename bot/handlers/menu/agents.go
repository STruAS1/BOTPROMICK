package menu

import (
	"BOTPROMICK/Utilities"
	"BOTPROMICK/db"
	"BOTPROMICK/db/models/user"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NetworkAgents(botCtx *user.BotContext) {
	botCtx.UpdateUserLevel(2)
	Agents(botCtx, true)
}
func NetworkAgentsWaitForComfirmed(botCtx *user.BotContext) {
	botCtx.UpdateUserLevel(3)
	Agents(botCtx, false)
}

type AgentsPages struct {
	CurrentPage       uint
	CountOfPage       uint
	TotalCounOfAgents int64
	Pages             [][]struct {
		FullName string
		UserID   uint
		Name     string
		IsOwner  bool
		tgID     int64
	}
}

func Agents(botCtx *user.BotContext, confirmed bool) {
	state := botCtx.GetUserState()
	if botCtx.User.UserNetwork != nil && botCtx.User.UserNetwork.CanInviteUser {
		network := botCtx.User.UserNetwork.Network(db.DB)
		if network == nil {
			return
		}
		if confirmed {

		} else {
			if botCtx.CallbackQuery != nil {
				data := strings.Split(botCtx.CallbackQuery.Data, "_")
				switch data[0] {
				case "сonfirm":
					if len(data) == 2 {
						userTGID, err := strconv.ParseInt(data[1], 10, 64)
						if err != nil {
							return
						}
						User, err := user.GetUser(db.DB, userTGID, "", "", "")
						if err != nil {
							return
						}
						if User.UserNetwork == nil || User.UserNetwork.NetworkID != network.ID || User.UserNetwork.Confirmed {
							return
						}
						User.UserNetwork.Confirmed = true
						db.DB.Save(User.UserNetwork)
						botCtx.Ctx.BotAPI.Send(tgbotapi.NewMessage(userTGID, "Вас приняли в сеть!"))
						delete(state.Data, "AgentsPages")
						NetworkAgents(botCtx)
						return
					}
				case "reject":
					if len(data) == 2 {
						userTGID, err := strconv.ParseInt(data[1], 10, 64)
						if err != nil {
							return
						}
						User, err := user.GetUser(db.DB, userTGID, "", "", "")
						if err != nil {
							return
						}
						if User.UserNetwork == nil || User.UserNetwork.Confirmed {
							return
						}
						err = network.RemoveUser(db.DB, User, botCtx.Ctx.BotAPI, "Ваша заявка на вступление в сеть отклонена!")
						if err != nil {
							fmt.Print(err)
						}
						delete(state.Data, "AgentsPages")
						NetworkAgents(botCtx)
						return

					}
				}

			}
		}
		const UsersPerPage = 8

		networkIDBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(networkIDBytes, uint32(network.ID+1_000_000_000))
		networkCode := hex.EncodeToString(networkIDBytes)

		_AgentsPages, exist := state.Data["AgentsPages"].(AgentsPages)
		if !exist {
			Agents, err := network.GetAllUsers(db.DB, confirmed)
			if err != nil {
				return
			}
			_AgentsPages.TotalCounOfAgents = int64(len(Agents))
			_AgentsPages.Pages = make([][]struct {
				FullName string
				UserID   uint
				Name     string
				IsOwner  bool
				tgID     int64
			}, 0, 10)

			for i, agent := range Agents {
				pageIndex := i / UsersPerPage
				if pageIndex >= len(_AgentsPages.Pages) {
					_AgentsPages.Pages = append(_AgentsPages.Pages, make([]struct {
						FullName string
						UserID   uint
						Name     string
						IsOwner  bool
						tgID     int64
					}, 0, UsersPerPage))
				}

				_AgentsPages.Pages[pageIndex] = append(_AgentsPages.Pages[pageIndex], struct {
					FullName string
					UserID   uint
					Name     string
					IsOwner  bool
					tgID     int64
				}{
					UserID:   agent.UserID,
					Name:     agent.Username,
					IsOwner:  agent.IsOwner,
					FullName: agent.FullName,
					tgID:     agent.TgID,
				})
			}

			_AgentsPages.CountOfPage = uint(len(_AgentsPages.Pages))
		}
		var rows [][]tgbotapi.InlineKeyboardButton
		if len(_AgentsPages.Pages) != 0 || int(_AgentsPages.CurrentPage) < len(_AgentsPages.Pages) {

			for i, agent := range _AgentsPages.Pages[_AgentsPages.CurrentPage] {
				label := agent.FullName
				if agent.Name != "" {
					label += fmt.Sprintf(" (@%s)", agent.Name)
				}
				if agent.IsOwner {
					label += " 👑"
				}
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("Agent_%d_%d", _AgentsPages.CurrentPage, i)),
				))
			}

			var showPrev, showNext bool
			if _AgentsPages.CountOfPage > 0 {
				showPrev = _AgentsPages.CurrentPage > 0
				showNext = _AgentsPages.CurrentPage+1 < _AgentsPages.CountOfPage
			}

			var row []tgbotapi.InlineKeyboardButton
			if showPrev {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️", "page_Prev"))
			}
			if showNext {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️", "page_next"))
			}
			if len(row) > 0 {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(row...))
			}
		}
		var msgText string

		if confirmed {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Заявки на вступление (%d)", network.CountOfUser-_AgentsPages.TotalCounOfAgents), "usersWaitForComfirmed")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🚀 Главное меню", "back")))
			msgText = fmt.Sprintf("🔑 Код на вступление в сеть: <i><code>%s</code></i>\n\n"+
				"📎 Ссылка на вступление: \n<i><code>https://t.me/%s?start=netId_%s</code></i>\n\n"+
				"👥 Список агентов сети:",
				networkCode, botCtx.Ctx.BotAPI.Self.UserName, networkCode)
		} else {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "back")))
			msgText = "👥 Список заявок:"
		}
		if state.MessageID == 0 {
			msg := tgbotapi.NewMessage(botCtx.TelegramID, msgText)
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			botCtx.SendMessage(msg)
		} else {
			msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.TelegramID, state.MessageID, msgText, tgbotapi.NewInlineKeyboardMarkup(rows...))
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)
		}

		state.Data["AgentsPages"] = _AgentsPages
	}
}

func ConFirmUser(botCtx *user.BotContext, page, Index int) {
	state := botCtx.GetUserState()
	_AgentsPages, exist := state.Data["AgentsPages"].(AgentsPages)
	if !exist {
		NetworkAgentsWaitForComfirmed(botCtx)
	}
	Agent := _AgentsPages.Pages[page][Index]
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Принять ✅", fmt.Sprintf("сonfirm_%d", Agent.tgID)),
		tgbotapi.NewInlineKeyboardButtonData("Отклонить ❌", fmt.Sprintf("reject_%d", Agent.tgID)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "hui")))
	msgText := fmt.Sprintf("<b><i><a href='https://t.me/%s'>👤 %s</a></i></b>", Agent.Name, Agent.FullName)
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		botCtx.TelegramID,
		state.MessageID,
		msgText,
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)
	msg.ParseMode = "HTML"
	if _, err := botCtx.Ctx.BotAPI.Send(msg); err != nil {
		fmt.Println(err)
	}
}

func EditUser(botCtx *user.BotContext, page, Index int) {
	state := botCtx.GetUserState()
	botCtx.UpdateUserLevel(6)
	_AgentsPages, exist := state.Data["AgentsPages"].(AgentsPages)
	if !exist {
		NetworkAgents(botCtx)
	}
	Agent := _AgentsPages.Pages[page][Index]
	User, err := user.GetUser(db.DB, Agent.tgID, "", "", "")
	if err != nil {
		return
	}
	msgText := fmt.Sprintf("<b><i><a href='https://t.me/%s'>👤 %s</a></i></b>", User.Username, User.FullName)
	msgText += fmt.Sprintf("💰 <b>Баланс: <code>%s</code></b>\n\n", Utilities.ConvertToFancyStringFloat(fmt.Sprintf("%f", float64(User.Balance/100))))
	msgText += "<b>📊Продаж:</b>\n"
	msgText += fmt.Sprintf("⭐️ <b>Сегодня:</b> <code>%s</code>\n", Utilities.ConvertToFancyString(1))
	msgText += fmt.Sprintf("👀 <b>За всё время:</b> <code>%s</code>\n", Utilities.ConvertToFancyString(1))
	var rows [][]tgbotapi.InlineKeyboardButton
	sufixs := [4]string{"❌", "❌", "❌", "❌"}
	for i, sufix := range [...]bool{
		User.UserNetwork.CanSell,
		User.UserNetwork.CanInviteUser,
		User.UserNetwork.CanViewAllSales,
		User.UserNetwork.CanEditNetwork,
	} {
		if sufix {
			sufixs[i] = "✅"
		}
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Делать продажи "+sufixs[0], "hui")))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Просмотр/Приглашение агентов "+sufixs[1], "hui")))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Просмотр всех продаж сети "+sufixs[2], "hui")))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Редактирование сети "+sufixs[3], "hui")))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("« назад", "hui")))

}
