package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chrislgardner/go-discord-bot/hnydiscordgo"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
)

type catFact struct {
	Fact   string
	length int
}

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

	session.AddHandler(messageRespond)
}

func messageRespond(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!") {
		return
	}

	ctx := context.Background()
	var span *trace.Span
	me := hnydiscordgo.MessageEvent{Message: m.Message, Context: ctx}

	ctx, span = hnydiscordgo.StartSpanOrTraceFromMessage(&me, s)

	m.Content = strings.Replace(m.Content, "!", "", 1)

	if strings.HasPrefix(m.Content, "ping") {
		span.AddField("command", "ping")
		s.ChannelMessageSend(m.ChannelID, "pong")
	} else if strings.HasPrefix(m.Content, "test") {
		span.AddField("command", "test")
		time.Sleep(3 * time.Second)
		s.ChannelMessageSend(m.ChannelID, "test success")
	} else if strings.HasPrefix(m.Content, "split") {
		span.AddField("command", "split")
		str := strings.Split(m.Content, " ")
		s.ChannelMessageSend(m.ChannelID, strings.Join(str[1:], "-"))
	} else if strings.HasPrefix(m.Content, "catfact") {
		span.AddField("command", "catfact")

		beeline.StartSpan(ctx, "catfact")

		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			beeline.AddField(ctx, "error", err)
			span.AddField("error", err)
		}

		defer resp.Body.Close()

		var fact catFact

		err = json.NewDecoder(resp.Body).Decode(&fact)
		if err != nil {
			beeline.AddField(ctx, "error", err)
			span.AddField("error", err)
		}

		beeline.AddField(ctx, "response", fact.Fact)

		s.ChannelMessageSend(m.ChannelID, fact.Fact)

	} else if strings.HasPrefix(m.Content, "relationships") {
		span.AddField("command", "relationships")

		s.ChannelMessageSend(m.ChannelID, "some interesting otter things")
	}

	span.Send()

}
