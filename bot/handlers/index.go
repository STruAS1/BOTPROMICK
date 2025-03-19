package handlers

import (
	"BOTPROMICK/bot/handlers/menu"
	"BOTPROMICK/db/models/user"
	"strings"
)

var nameHandlers = map[string]func(*user.BotContext){
	"menu": menu.Handle,
	// "JokeMenu": MenuJokes.Handle,
}

func HandleUpdate(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	if botCtx.Message != nil {
		switch botCtx.Message.Command() {
		case "start":
			state.MessageID = 0
			botCtx.ClearAllUserData()
			menu.HandleStartCommand(botCtx)
			return
		}
	}
	if state.Level != 0 {
		if handler, exists := nameHandlers[state.Name]; exists {
			handler(botCtx)
		} else {
			menu.Handle(botCtx)
		}
	} else {
		if botCtx.CallbackQuery != nil {
			switch strings.Split(botCtx.CallbackQuery.Data, "_")[0] {
			case "NetworkAgents":
				botCtx.UpdateUserName("menu")
				menu.NetworkAgents(botCtx)
			case "JoinNetwork":
				botCtx.UpdateUserName("menu")
				menu.HandleJoinNetwork(botCtx)
			case "Cancel":
				botCtx.UpdateUserName("menu")
				menu.CancelToJoinNetwork(botCtx)
			case "NetworkSettingsName":
				botCtx.UpdateUserName("menu")
				menu.EditNetworkNameHandler(botCtx)
			// case "Docs":
			// 	context.UpdateUserName(botCtx, "start")
			// 	start.HandleDocs(botCtx)
			// case "ConnectWallet":
			// 	context.UpdateUserName(botCtx, "start")
			// 	start.HandleTonConnect(botCtx)
			// case "Settings":
			// 	context.UpdateUserName(botCtx, "start")
			// 	start.HandleSettings(botCtx)
			// case "Withdraw":
			// 	context.UpdateUserName(botCtx, "start")
			// 	start.HandleWithdraw(botCtx)
			// case "DisconnectWallet":
			// 	context.UpdateUserName(botCtx, "start")
			// 	TonConnectCallback.Disconnect(botCtx.UserID)
			// 	start.HandleSettings(botCtx)
			// case "SetAuthor":
			// 	context.UpdateUserName(botCtx, "start")
			// 	start.HandleSetAuthor(botCtx)
			// case "SetAnonymsMode":
			// 	var user models.User
			// 	db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)
			// 	user.SetAnonymsMode(db.DB)
			// 	context.UpdateUserName(botCtx, "start")
			// 	start.HandleSettings(botCtx)
			// case "NewJoke":
			// 	// time, IsCooldow := Utilities.GetRemainingCooldown(uint(botCtx.UserID))
			// 	// if IsCooldow {
			// 	// 	alert := tgbotapi.NewCallbackWithAlert(botCtx.CallbackQuery.ID, "Вы сможете шуткануть через "+time)
			// 	// 	alert.ShowAlert = false
			// 	// 	botCtx.Ctx.BotAPI.Request(alert)
			// 	// 	return
			// 	// }
			// 	context.UpdateUserName(botCtx, "JokeMenu")
			// 	MenuJokes.NewJokeHandle(botCtx)
			// case "ViewJokes":
			// 	context.UpdateUserName(botCtx, "JokeMenu")
			// 	MenuJokes.HandleJokeViewer(botCtx)
			// case "MyJokes":
			// 	context.UpdateUserName(botCtx, "JokeMenu")
			// 	MenuJokes.HandleMyJokes(botCtx)

			default:
				state.MessageID = 0
				botCtx.ClearAllUserData()
				menu.HandleStartCommand(botCtx)
				return
			}
		}
		if botCtx.Message != nil {
			state.MessageID = 0
			botCtx.ClearAllUserData()
			menu.HandleStartCommand(botCtx)
			return
		}
	}

}
