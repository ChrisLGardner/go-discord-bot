package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/honeycombio/beeline-go"
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/entities"
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

	go sendReminders(session)

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	session.AddHandler(MessageRespond)
	session.AddHandler(MessageReact)
	session.AddHandler(JoinThread)
}

func getFeatureFlagState(ctx context.Context, id string, roles []string, flag string) bool {

	ctx, span := beeline.StartSpan(ctx, "get_feature_flag_main")
	defer span.Send()

	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: os.Getenv("OPTIMIZELY_KEY"),
	}

	optlyClient, err := optimizelyFactory.Client()
	if err != nil {
		panic(err)
	}

	defer optlyClient.Close()

	beeline.AddField(ctx, "feature_flag_name", flag)
	beeline.AddField(ctx, "feature_flag_role", roles)

	enabled := false

	for _, role := range roles {
		ctx, span := beeline.StartSpan(ctx, "get_feature_flag")
		defer span.Send()

		beeline.AddField(ctx, "feature_flag_name", flag)
		beeline.AddField(ctx, "feature_flag_role", role)

		attributes := map[string]interface{}{
			"role": role,
		}

		user := entities.UserContext{
			ID:         id,
			Attributes: attributes,
		}

		enabled, _ := optlyClient.IsFeatureEnabled(flag, user)
		beeline.AddField(ctx, "feature_flag_enabled", enabled)

		if enabled {
			return enabled
		}
	}

	return enabled
}

func JoinThread(s *discordgo.Session, t *discordgo.ThreadCreate) {
	// ctx, span := beeline.StartSpan(ctx, "get_feature_flag_main")
	// defer span.Send()
	if t.IsThread() {
		s.ThreadJoin(t.Channel.ID)
	}
}
