package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chrislgardner/go-discord-bot/hnydiscordgo"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
	rcon "github.com/katnegermis/pocketmine-rcon"
)

type catFact struct {
	Fact   string
	length int
}

type relationship struct {
	Synergy       string
	Objective     string
	Relationships int
	Credit        string
}

func sendResponse(ctx context.Context, s *discordgo.Session, cid string, m string) {

	ctx, span := beeline.StartSpan(ctx, "send_response")
	defer span.Send()
	beeline.AddField(ctx, "response", m)
	beeline.AddField(ctx, "chennel", cid)

	s.ChannelMessageSend(cid, m)

}

func sendReply(ctx context.Context, s *discordgo.Session, m string, om *discordgo.MessageReference) {

	ctx, span := beeline.StartSpan(ctx, "sendReply")
	defer span.Send()

	span.AddField("sendReply.response", m)
	span.AddField("sendReply.originalMessage.id", om.MessageID)
	span.AddField("sendReply.originalMessage.guildID", om.GuildID)
	span.AddField("sendReply.originalMessage.channelID", om.ChannelID)

	s.ChannelMessageSendReply(om.ChannelID, m, om)

}

func chooseRandom(opt []string) (string, int) {
	randomIndex := rand.Intn(len(opt))
	choice := opt[randomIndex]

	return choice, randomIndex
}

func getCatFact(ctx context.Context) (catFact, error) {

	ctx, span := beeline.StartSpan(ctx, "getCatFact")

	defer span.Send()
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

func sendMinecraftCommand(ctx context.Context, comm string) (string, error) {
	ctx, span := beeline.StartSpan(ctx, "minecraft_command")
	defer span.Send()

	conn, err := connectMinecraft(ctx)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return "", err
	}

	comm = strings.TrimPrefix(comm, "mc ")
	r, err := conn.SendCommand(comm)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return "", err
	}

	return r, nil

}

func connectMinecraft(ctx context.Context) (*rcon.Connection, error) {
	ctx, span := beeline.StartSpan(ctx, "connect_minecraft")
	defer span.Send()

	addr := os.Getenv("MCSERVERADDR")
	pass := os.Getenv("MCSERVERPASS")

	beeline.AddField(ctx, "mc.server.address", addr)
	conn, err := rcon.NewConnection(addr, pass)

	if err != nil {
		beeline.AddField(ctx, "error", err)
		return nil, err
	}
	return conn, nil
}

func adilioMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.Contains(strings.ToLower(m.Message.Content), "lol") {
		ctx := context.Background()
		var span *trace.Span
		me := hnydiscordgo.MessageEvent{Message: m.Message, Context: ctx}

		ctx, span = hnydiscordgo.StartSpanOrTraceFromMessage(&me, s)
		span.AddField("command", "AdilioLol")

		sendResponse(ctx, s, m.ChannelID, "<:adilio:788826086628261889> <:adilol:769263097772245032>")

		span.Send()
	}
	if strings.Contains(strings.ToLower(m.Message.Content), "idea") {
		ctx := context.Background()
		var span *trace.Span
		me := hnydiscordgo.MessageEvent{Message: m.Message, Context: ctx}

		ctx, span = hnydiscordgo.StartSpanOrTraceFromMessage(&me, s)
		span.AddField("command", "AdilioIdea")

		sendResponse(ctx, s, m.ChannelID, "<:steviecoaster:767894596687888444> <:steviefok:774365852698804224>")

		span.Send()
	}

}

func quipMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	var span *trace.Span
	me := hnydiscordgo.MessageEvent{Message: m.Message, Context: ctx}

	ctx, span = hnydiscordgo.StartSpanOrTraceFromMessage(&me, s)

	if strings.Contains(strings.ToLower(m.Message.Content), "bezos") {
		span.AddField("command", "QuipBezos")
		quip := "Do you mean the ex-husband of billionaire philanthropist Mackenzie Scott?"
		sendResponse(ctx, s, m.ChannelID, quip)
		span.Send()
	}
}

func featureRequestResponse(ctx context.Context, author string)) string {
	ctx, span := beeline.StartSpan(ctx, "featureRequestResponse")
	defer span.Send()

	fqResponses := []string{"File your own damned issue <@%s>: https://github.com/ChrisLGardner/go-discord-bot/issues",
		"Hey <@%s>, I'll keep an eye out for your PR: https://github.com/ChrisLGardner/go-discord-bot/pulls"}
	span.AddField("featureRequestResponse.possibleChoices", fqResponses)

	fqResponse, randNum := chooseRandom(fqResponses)
	message := fmt.Sprintf(fqResponse, author)

	span.AddField("featureRequestResponse.randomNumber", randNum)

	return message
}

func languageResponse(ctx context.Context) string {
	ctx, span := beeline.StartSpan(ctx, "languageResponse")
	defer span.Send()

	languageGifs := []string{"https://tenor.com/view/captain-america-marvel-avengers-gif-18378867",
		"https://tenor.com/view/marvel-tony-stark-iron-man-gif-18079972",
		"https://tenor.com/view/captain-america-marvel-avengers-gif-14328153"}
	span.AddField("languageResponse.possibleChoices", languageGifs)

	pickGif, randNum := chooseRandom(languageGifs)
	span.AddField("languageResponse.randomNumber", randNum)

	return pickGif
}

func toBeFairResponse(ctx context.Context) string {
	ctx, span := beeline.StartSpan(ctx, "toBeFairResponse")
	defer span.Send()

	toBeFairGifs := []string{"https://tenor.com/view/letter-kenny-wayne-to-be-fair-gif-14458907",
		"https://tenor.com/view/letterkenny-to-be-fair-serious-lets-be-fair-gif-16087355",
		"https://tenor.com/view/letterkenny-to-be-tobefair-gif-14136631"}
	span.AddField("toBeFairResponse.possibleChoices", toBeFairGifs)

	pickGif, randNum := chooseRandom(toBeFairGifs)
	span.AddField("toBeFairResponse.randomNumber", randNum)

	return pickGif
}

func kevinResponse(ctx context.Context) string {
	ctx, span := beeline.StartSpan(ctx, "kevinResponse")
	defer span.Send()

	kevins := []string{
		"https://gph.is/g/4zVyePw",
		"https://tenor.com/view/home-alone-kevin-gif-15171451",
	}

	span.AddField("languageResponse.possibleChoices", kevins)

	pickGif, randNum := chooseRandom(kevins)
	span.AddField("languageResponse.randomNumber", randNum)

	return pickGif
}

func getRelationship(ctx context.Context) (relationship, error) {

	ctx, span := beeline.StartSpan(ctx, "getRelationship")

	defer span.Send()
	resp, err := http.Get("https://buildingrelationships.dev")
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return relationship{}, err
	}

	defer resp.Body.Close()

	var rel relationship

	err = json.NewDecoder(resp.Body).Decode(&rel)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return relationship{}, err
	}

	beeline.AddField(ctx, "relationship.synergy", rel.Synergy)
	beeline.AddField(ctx, "relationship.objective", rel.Objective)
	beeline.AddField(ctx, "relationship.credit", rel.Credit)

	return rel, nil
}

func getTime(ctx context.Context, t time.Time, s string) (string, error) {
	ctx, span := beeline.StartSpan(ctx, "getTime")
	defer span.Send()

	span.AddField("timezone.member", s)
	span.AddField("timezone.timeNow", t.UTC())

	if s == "" {
		span.AddField("timezone.error", "no user specified")
		return "no user specified", nil
	}

	memberTimes := make(map[string]string)

	err := json.Unmarshal([]byte(os.Getenv("MEMBER_TIMEZONES")), &memberTimes)
	if err != nil {
		span.AddField("timezone.error", err)
		return "", err
	}

	if memberTimes[s] == "" {
		err := fmt.Errorf("User not found")
		span.AddField("timezone.error", err)
		return "", err
	}

	location, err := time.LoadLocation(memberTimes[s])
	if err != nil {
		span.AddField("timezone.error", err)
		return "", err
	}
	span.AddField("timezone.location.raw", memberTimes[s])
	span.AddField("timezone.location.time", location)

	raw := t.In(location)
	result := fmt.Sprintf("%s : %02d:%02d, %d %s %d, (%s)", s, raw.Hour(), raw.Minute(), raw.Day(), raw.Month(), raw.Year(), raw.Location())

	span.AddField("timezone.result", result)

	return result, nil
}
