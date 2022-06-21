package hnydiscordgo

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/honeycombio/beeline-go/propagation"
	"github.com/honeycombio/beeline-go/trace"
)

//MessageEvent Wrapper type to track context with message body
type MessageEvent struct {
	Message *discordgo.Message
	Context context.Context
}

// StartSpanOrTraceFromMessage creates and returns a span for the provided MessageEvent. If
// there is an existing span in the Context, this function will create the new span as a
// child span and return it. If not, it will create a new trace object and return the root
// span.
func StartSpanOrTraceFromMessage(me *MessageEvent, s *discordgo.Session) (context.Context, *trace.Span) {
	ctx := me.Context
	span := trace.GetSpanFromContext(ctx)

	if span == nil {
		// there is no trace yet. We should make one! and use the root span.
		var tr *trace.Trace
		ctx, tr = trace.NewTrace(ctx, &propagation.PropagationContext{})

		span = tr.GetRootSpan()
	} else {
		// we had a parent! let's make a new child for this handler
		ctx, span = span.CreateChild(ctx)
	}
	// go get any common HTTP headers and attributes to add to the span
	for k, v := range getMessageProps(me) {
		span.AddField(k, v)
	}

	for k, v := range getSessionProps(s) {
		span.AddField(k, v)
	}

	return ctx, span
}

func StartTraceFromThreadJoin(c *discordgo.Channel, s *discordgo.Session) (context.Context, *trace.Span) {
	ctx := context.Background()
	var tr *trace.Trace
	ctx, tr = trace.NewTrace(ctx, &propagation.PropagationContext{})

	span := tr.GetRootSpan()

	span.AddField("name", "JoinThread")
	for k, v := range getChannelProps(c) {
		span.AddField(k, v)
	}

	for k, v := range getSessionProps(s) {
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

func getChannelProps(c *discordgo.Channel) map[string]interface{} {
	channelProps := make(map[string]interface{})

	channelProps["channel.ID"] = c.ID
	channelProps["channel.GuildID"] = c.GuildID
	channelProps["channel.Name"] = c.Name
	channelProps["channel.Type"] = c.Type
	channelProps["channel.NSFW"] = c.NSFW
	channelProps["channel.ParentID"] = c.ParentID

	if c.IsThread() {
		channelProps["channel.OwnerID"] = c.OwnerID
	}

	return channelProps
}
