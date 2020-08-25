package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
)

//MessageEvent Wrapper type to track context with message body
type MessageEvent struct {
	Message *discordgo.Message
	Context context.Context
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

	m.Content = strings.Replace(m.Content, "!", "", 1)

	ctx := context.Background()
	var span *trace.Span
	me := MessageEvent{m.Message, ctx}

	ctx, span = StartSpanOrTraceFromMessage(&me)

	for k, v := range getSessionProps(s) {
		span.AddField(k, v)
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	} else if m.Content == "test" {
		time.Sleep(3 * time.Second)
		s.ChannelMessageSend(m.ChannelID, "test success")
	}

	beeline.Flush(ctx)

}

// StartSpanOrTraceFromMessage creates and returns a span for the provided MessageEvent. If
// there is an existing span in the Context, this function will create the new span as a
// child span and return it. If not, it will create a new trace object and return the root
// span.
func StartSpanOrTraceFromMessage(me *MessageEvent) (context.Context, *trace.Span) {
	ctx := me.Context
	span := trace.GetSpanFromContext(ctx)
	if span == nil {
		// there is no trace yet. We should make one! and use the root span.
		var tr *trace.Trace
		ctx, tr = trace.NewTrace(ctx, "")

		span = tr.GetRootSpan()
	} else {
		// we had a parent! let's make a new child for this handler
		ctx, span = span.CreateChild(ctx)
	}
	// go get any common HTTP headers and attributes to add to the span
	for k, v := range getMessageProps(me) {
		span.AddField(k, v)
	}
	return ctx, span
}

func getMessageProps(me *MessageEvent) map[string]interface{} {

	messageProps := make(map[string]interface{})

	messageProps["message.ID"] = me.Message.ID
	messageProps["message.ChannelID"] = me.Message.ChannelID
	messageProps["message.GuildID"] = me.Message.GuildID
	messageProps["message.AuthorID"] = me.Message.Author.ID
	messageProps["message.AuthorUsername"] = me.Message.Author.Username
	messageProps["message.MessageType"] = me.Message.Type
	messageProps["message.RawContent"] = me.Message.Content
	messageProps["message.MentionEveryone"] = me.Message.MentionEveryone
	messageProps["message.MentionRoles"] = me.Message.MentionRoles

	channels := []string{""}
	for _, mc := range me.Message.MentionChannels {
		channels = append(channels, mc.ID)
	}

	messageProps["message.MentionChannels"] = channels

	mentions := []string{""}
	for _, m := range me.Message.Mentions {
		mentions = append(mentions, m.ID)
	}

	messageProps["message.Mentions"] = mentions

	return messageProps
}

func getSessionProps(s *discordgo.Session) map[string]interface{} {
	sessionProps := make(map[string]interface{})

	sessionProps["session.ShardID"] = s.ShardID

	return sessionProps
}
