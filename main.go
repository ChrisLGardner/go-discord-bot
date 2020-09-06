package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/chrislgardner/go-discord-bot/commands"
	"github.com/honeycombio/beeline-go"
)

func main() {

	beeline.Init(beeline.Config{
		WriteKey: os.Getenv("HONEYCOMB_KEY"),
		Dataset:  os.Getenv("HONEYCOMB_DATASET"),
	})

	defer beeline.Close()

	// Open a simple Discord session
	token := os.Getenv("DISCORD_TOKEN")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	err = session.Open()
	if err != nil {
		panic(err)
	}

	// Wait for the user to cancel the process
	defer func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
	}()

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	session.AddHandler(commands.MessageRespond)
}
