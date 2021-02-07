package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/honeycombio/beeline-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type reminder struct {
	due             time.Time
	message         string
	server          string
	creator         string
	channel         string
	sourceMessage   string
	sourceTimestamp time.Time
}

func sendReminders(session *discordgo.Session) {
	storage := os.Getenv("COSMOSDB_URI")
	interval, err := strconv.Atoi(os.Getenv("REMINDER_INTERVAL"))
	if err != nil {
		interval = 5
	}
	for {
		time.Sleep(time.Duration(interval) * time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ctx, span := beeline.StartSpan(ctx, "sendReminders")

		db, err := connectDb(ctx, storage)
		if err != nil {
			span.AddField("sendReminders.connect.error", err)
			span.Send()
			continue
		}

		reminders, err := findReminders(ctx, db, interval)
		if err != nil {
			span.AddField("sendReminders.find.error", err)
			span.Send()
			continue
		}

		for _, r := range reminders {
			ctx, childSpan := beeline.StartSpan(ctx, "sendReminderIndividual")
			childSpan.AddField("sendReminderIndividual.due", r.due)
			childSpan.AddField("sendReminderIndividual.message", r.message)
			childSpan.AddField("sendReminderIndividual.server", r.server)
			childSpan.AddField("sendReminderIndividual.creator", r.creator)
			childSpan.AddField("sendReminderIndividual.channel", r.channel)
			childSpan.AddField("sendReminderIndividual.sourceMessage", r.sourceMessage)
			childSpan.AddField("sendReminderIndividual.sourceTimestamp", r.sourceTimestamp)

			message := fmt.Sprintf("Hey <@%s>, remember %s", r.creator, r.message)

			sendResponse(ctx, session, r.channel, message)

			childSpan.Send()
		}

		span.Send()
	}
}

func findReminders(ctx context.Context, db *mongo.Client, interval int) ([]reminder, error) {

	return nil, nil
}

func createReminder(ctx context.Context, message *discordgo.Message) (string, error) {

	ctx, span := beeline.StartSpan(ctx, "createReminder")
	defer span.Send()

	r, err := parseReminder(ctx, message)
	if err != nil {
		span.AddField("createReminder.error", err)
		return "", err
	}

	err = storeReminder(ctx, r)
	if err != nil {
		span.AddField("createReminder.error", err)
		return "", err
	}

	return "Reminder added.", nil
}

func parseReminder(ctx context.Context, message *discordgo.Message) (reminder, error) {

	ctx, span := beeline.StartSpan(ctx, "parseReminder")
	defer span.Send()

	// expect message content to be like:
	// "do the thing 5h"
	// "5d wrankle the sprockets"
	// support minute, hour, day, Month

	timePattern := regexp.MustCompile("(?P<count>\\d+)(?P<interval>m|[hH]|[dD]|M)")
	names := timePattern.SubexpNames()
	elements := map[string]string{}
	matchingStrings := timePattern.FindAllStringSubmatch(message.Content, -1)
	matches := []string{}

	span.AddField("parseReminder.matcheslength", len(matchingStrings))
	span.AddField("parseReminder.matches", matchingStrings)

	if len(matchingStrings) == 0 {
		span.AddField("parseReminder.error", "No interval specified")
		return reminder{}, fmt.Errorf("No interval specified")
	}

	matches = matchingStrings[0]

	for i, match := range matches {
		elements[names[i]] = match
	}

	span.AddField("parseReminder.count", elements["count"])
	span.AddField("parseReminder.interval", elements["interval"])

	reminderText := strings.Replace(message.Content, fmt.Sprintf("%s%s", elements["count"], elements["interval"]), "", 1)

	sourceDate, err := message.Timestamp.Parse()
	if err != nil {
		span.AddField("parseReminder.error", err)
		return reminder{}, err
	}

	var dueDate time.Time
	timeCount, _ := strconv.Atoi(elements["count"])

	switch elements["interval"] {
	case "m":
		dueDate = sourceDate.Add(time.Duration(timeCount) * time.Minute)
	case "h", "H":
		dueDate = sourceDate.Add(time.Duration(timeCount) * time.Hour)
	case "d", "D":
		dueDate = sourceDate.AddDate(0, 0, timeCount)
	case "M":
		dueDate = sourceDate.AddDate(0, timeCount, 0)
	}

	r := reminder{
		due:             dueDate,
		message:         reminderText,
		server:          message.GuildID,
		creator:         message.Author.ID,
		channel:         message.ChannelID,
		sourceMessage:   message.ID,
		sourceTimestamp: sourceDate,
	}

	span.AddField("parseReminder.due", r.due)
	span.AddField("parseReminder.message", r.message)
	span.AddField("parseReminder.server", r.server)
	span.AddField("parseReminder.creator", r.creator)
	span.AddField("parseReminder.channel", r.channel)
	span.AddField("parseReminder.sourceMessage", r.sourceMessage)
	span.AddField("parseReminder.sourceTimestamp", r.sourceTimestamp)

	return r, nil
}

func storeReminder(ctx context.Context, r reminder) error {

	ctx, span := beeline.StartSpan(ctx, "storeReminder")
	defer span.Send()

	db, err := connectDb(ctx, os.Getenv("COSMOSDB_URI"))
	if err != nil {
		span.AddField("storeReminder.error", err)
		return err
	}

	err = writeDbObject(ctx, db, r)
	if err != nil {
		span.AddField("storeReminder.error", err)
		return err
	}

	if err = db.Disconnect(ctx); err != nil {
		span.AddField("storeReminder.error", err)
		return err
	}

	return nil
}
