package menu

import (
	"BOTPROMICK/db/models/user"
	"strconv"
	"strings"
)

func Handle(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	switch state.Level {
	case 1:
		handleLvl1(botCtx)
	case 2:
		handleLvl2(botCtx)
	case 3:
		handleLvl3(botCtx)
	case 4:
		handleLvl4(botCtx)
	case 5:
		handleLvl5(botCtx)
	case 6:
		handleLvl6(botCtx)
	}

}

func handleLvl1(botCtx *user.BotContext) {
	if botCtx.Message != nil {
		HandleRegister(botCtx)
	}
}

func handleLvl2(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "back":
			delete(state.Data, "AgentsPages")
			HandleStartCommand(botCtx)
		case "usersWaitForComfirmed":
			delete(state.Data, "AgentsPages")
			NetworkAgentsWaitForComfirmed(botCtx)
		case "Agent":
			if len(data) == 3 {
				UserId, _ := strconv.ParseInt(data[2], 10, 8)
				pageId, _ := strconv.ParseInt(data[1], 10, 8)
				EditUser(botCtx, int(pageId), int(UserId))
			}
		default:
			NetworkAgents(botCtx)
		}
	}
}

func handleLvl3(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "back":
			delete(state.Data, "AgentsPages")
			NetworkAgents(botCtx)
		case "Agent":
			if len(data) == 3 {
				UserId, _ := strconv.ParseInt(data[2], 10, 0)
				pageId, _ := strconv.ParseInt(data[1], 10, 0)
				ConFirmUser(botCtx, int(pageId), int(UserId))
			}
		default:
			NetworkAgentsWaitForComfirmed(botCtx)
		}
	}
}

func handleLvl4(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "back":
			delete(state.Data, "JoinNetwork")
			HandleStartCommand(botCtx)
		}
	} else {
		HandleJoinNetwork(botCtx)
	}
}

func handleLvl5(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "back":
			delete(state.Data, "EditNetwork")
			EditNetworkNameHandler(botCtx)
		default:
			EditNetworkNameHandler(botCtx)
		}
	} else {
		EditNetworkNameHandler(botCtx)
	}
}

func handleLvl6(botCtx *user.BotContext) {
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		if len(data) != 3 {
			NetworkAgents(botCtx)
			return
		}
		UserId, _ := strconv.ParseInt(data[2], 10, 0)
		pageId, _ := strconv.ParseInt(data[1], 10, 0)
		EditUser(botCtx, int(pageId), int(UserId))
	}
}
