package main

import (
	"context"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chrislgardner/go-discord-bot/hnydiscordgo"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
)

//MessageRespond is the handler for which message respond function should be called
func MessageRespond(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Username == "Adilio" {
		adilioMessage(s, m)
	}
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!") {
		return
	}

	ctx := context.Background()
	var span *trace.Span
	me := hnydiscordgo.MessageEvent{Message: m.Message, Context: ctx}

	ctx, span = hnydiscordgo.StartSpanOrTraceFromMessage(&me, s)

	m.Content = strings.Replace(m.Content, "!", "", 1)
	span.AddField("name", "MessageRespond")

	roles, err := getMemberRoles(ctx, s, m.Message)
	if err != nil {
		span.AddField("member.role.error", err)
	}

	span.AddField("member.roles", roles)

	if strings.HasPrefix(m.Content, "help") {
		span.AddField("command", "help")
		help := `Commands available:
		ping - returns pong if bot is running
		catfact - returns a random cat fact
		relationship - returns a random relationship objective or synergy
		mc - runs various minecraft commands if enabled for the user
		mtg - returns a scryfall search link based on user criteria, see mtg help for more details.
		`
		sendResponse(ctx, s, m.ChannelID, help)
	} else if strings.HasPrefix(m.Content, "ping") {
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
	} else if strings.HasPrefix(m.Content, "emoji") {
		span.AddField("command", "emoji-test")

		sendResponse(ctx, s, m.ChannelID, "<:emotest:788860836009345024>")
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

		enabled := false

		if getFeatureFlagState(ctx, m.Author.ID, roles, "relationship-command") {
			span.AddField("flags.relationship", true)
			enabled = true
		}

		if enabled {
			rel, err := getRelationship(ctx)

			if err != nil {
				span.AddField("error", err)
				sendResponse(ctx, s, m.ChannelID, "error getting cat fact")
			}
			if strings.Contains(m.Content, "synergy") {
				span.AddField("relationship.output.synergy", true)
				sendResponse(ctx, s, m.ChannelID, rel.synergy)
			} else {
				span.AddField("relationship.output.objective", true)
				sendResponse(ctx, s, m.ChannelID, rel.objective)
			}
		} else {
			span.AddField("flags.relationship", false)
			sendResponse(ctx, s, m.ChannelID, "Command not allowed")
		}
	} else if strings.HasPrefix(m.Content, "mc") {
		span.AddField("command", "minecraft")

		enabled := false

		if strings.Contains(m.Content, " whitelist ") {
			if getFeatureFlagState(ctx, m.Author.ID, roles, "mc-commands") {
				span.AddField("flags.minecraft", true)
				enabled = true
			}

			if enabled {
				resp, err := sendMinecraftCommand(ctx, m.Content)
				if err != nil {
					span.AddField("error", err)
					sendResponse(ctx, s, m.ChannelID, err.Error())
				}

				sendResponse(ctx, s, m.ChannelID, resp)
			}
		} else if getFeatureFlagState(ctx, m.Author.ID, roles, "mc-admin") {
			span.AddField("flags.minecraft", true)
			span.AddField("flags.minecraft-admin", true)

			resp, err := sendMinecraftCommand(ctx, m.Content)
			if err != nil {
				span.AddField("error", err)
				sendResponse(ctx, s, m.ChannelID, err.Error())
			}

			sendResponse(ctx, s, m.ChannelID, resp)

		} else {
			span.AddField("flags.minecraft", false)
			sendResponse(ctx, s, m.ChannelID, "Command not allowed")
		}
	} else if strings.HasPrefix(m.Content, "mtg") {
		span.AddField("command", "magic")

		str := strings.Replace(m.Content, "mtg", "", 1)

		resp, err := mtgCommand(ctx, strings.TrimSpace(str))
		if err != nil {
			span.AddField("error", err)
			sendResponse(ctx, s, m.ChannelID, err.Error())
		} else {
			sendResponse(ctx, s, m.ChannelID, resp)
		}
	}

	span.Send()

}

func getMemberRoles(ctx context.Context, s *discordgo.Session, m *discordgo.Message) ([]string, error) {
	ctx, span := beeline.StartSpan(ctx, "get_discord_role")
	defer span.Send()

	member, err := s.GuildMember(m.GuildID, m.Author.ID)

	if err != nil {
		beeline.AddField(ctx, "error", err)
		return nil, err
	}

	guildRoles, err := s.GuildRoles(m.GuildID)

	if err != nil {
		beeline.AddField(ctx, "error", err)
		return nil, err
	}

	var roles []string

	for _, role := range member.Roles {
		for _, guildRole := range guildRoles {
			if guildRole.ID == role {
				roles = append(roles, guildRole.Name)
			}
		}
	}

	return roles, nil
}
