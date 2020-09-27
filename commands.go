package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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

//MessageRespond is the handler for which message respond function should be called
func MessageRespond(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!") {
		return
	}

	ctx := context.Background()
	var span *trace.Span
	me := hnydiscordgo.MessageEvent{Message: m.Message, Context: ctx}

	ctx, span = hnydiscordgo.StartSpanOrTraceFromMessage(&me, s)

	m.Content = strings.Replace(m.Content, "!", "", 1)
	span.AddField("name", "MessageRespond")

	roles, err := GetMemberRoles(ctx, s, m.Message)
	if err != nil {
		span.AddField("member.role.error", err)
	}

	span.AddField("member.roles", roles)

	if strings.HasPrefix(m.Content, "ping") {
		span.AddField("command", "ping")
		sendResponse(ctx, s, m.ChannelID, "pong")
	} else if strings.HasPrefix(m.Content, "test") {
		span.AddField("command", "test")
		time.Sleep(3 * time.Second)
		sendResponse(ctx, s, m.ChannelID, "test success")
	} else if strings.HasPrefix(m.Content, "split") {
		span.AddField("command", "split")
		str := strings.Split(m.Content, " ")
		sendResponse(ctx, s, m.ChannelID, strings.Join(str[1:], "-"))
	} else if strings.HasPrefix(m.Content, "catfact") {
		span.AddField("command", "catfact")

		enabled := false

		if getFeatureFlagState(ctx, m.Author.ID, roles, "catfact-command") {
			span.AddField("flags.catfact", true)
			enabled = true
		}

		if enabled {
			fact, err := getCatFact(ctx)

			if err != nil {
				span.AddField("error", err)
				sendResponse(ctx, s, m.ChannelID, "error getting cat fact")
			}
			sendResponse(ctx, s, m.ChannelID, fact.Fact)
		} else {
			span.AddField("flags.catfact", false)
			sendResponse(ctx, s, m.ChannelID, "Command not allowed")
		}

	} else if strings.HasPrefix(m.Content, "relationships") {
		span.AddField("command", "relationships")

		s.ChannelMessageSend(m.ChannelID, "some interesting otter things")
	}

	span.Send()

}

func sendResponse(ctx context.Context, s *discordgo.Session, cid string, m string) {

	beeline.StartSpan(ctx, "send_response")
	beeline.AddField(ctx, "response", m)
	beeline.AddField(ctx, "chennel", cid)

	s.ChannelMessageSend(cid, m)

}

func getCatFact(ctx context.Context) (catFact, error) {

	beeline.StartSpan(ctx, "getCatFact")

	resp, err := http.Get("https://catfact.ninja/fact")
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return catFact{}, err
	}

	defer resp.Body.Close()

	var fact catFact

	err = json.NewDecoder(resp.Body).Decode(&fact)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return catFact{}, err
	}

	beeline.AddField(ctx, "response", fact.Fact)

	return fact, nil
}

func GetMemberRoles(ctx context.Context, s *discordgo.Session, m *discordgo.Message) ([]string, error) {
	ctx, span := beeline.StartSpan(ctx, "get_discord_role")
	defer span.Send()

	member, err := s.GuildMember(m.GuildID, m.Author.ID)

	if err != nil {
		beeline.AddField(ctx, "error", err)
		return nil, err
	}

	return member.Roles, nil
}
