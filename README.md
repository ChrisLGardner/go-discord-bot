# go-discord-bot
Discord Bot written in Go for doing stuff and learning Go.

## Feature flag usage

Many features of the bot are kept behind feature flags to limit access to them, primarily for testing but also as a form of RBAC. When adding new commands to the bot it is suggested to wrap it in a feature flag, such as the code used in the [remindme command](https://github.com/ChrisLGardner/go-discord-bot/blob/main/routing.go#L231-L237), and adding it to the flags.json file. This will ensure the flag is created in Optimizely when the code is deployed and new commands and features can be tested in a controlled way without impacting other users/servers.
