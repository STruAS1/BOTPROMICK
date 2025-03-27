package sales

import "BOTPROMICK/db/models/user"

func Handle(botCtx *user.BotContext) {
	state := botCtx.GetUserState()
	switch state.Level {
	case 1:
		NewSaleHandler(botCtx, 0)
	}

}
