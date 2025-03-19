package main

import (
	"BOTPROMICK/bot"
	"BOTPROMICK/config"
	"BOTPROMICK/db"
)

func main() {
	cfg := config.LoadConfig()
	db.Connect(cfg)
	bot.StartBot(cfg)
}
