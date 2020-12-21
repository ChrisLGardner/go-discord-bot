package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

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
	synergy       string
	objective     string
	relationships int
	credit        string
}

func sendResponse(ctx context.Context, s *discordgo.Session, cid string, m string) {

	ctx, span := beeline.StartSpan(ctx, "send_response")
	defer span.Send()
	beeline.AddField(ctx, "response", m)
	beeline.AddField(ctx, "chennel", cid)

	s.ChannelMessageSend(cid, m)

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

	beeline.AddField(ctx, "relationship.synergy", rel.synergy)
	beeline.AddField(ctx, "relationship.objective", rel.objective)
	beeline.AddField(ctx, "relationship.credit", rel.credit)

	return rel, nil
}
