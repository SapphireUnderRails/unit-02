package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Creating a struct to hold the Discord token.
type Token struct {
	DiscordToken string
}

// Creating a variable to hold the Token struct.
var token Token

// Main functions.
func main() {

	// Retrieve the tokens from the tokens.json file.
	tokensFile, err := os.ReadFile("token.json")
	if err != nil {
		log.Fatal("COULD NOT READ 'token.json' FILE: ", err)
	}

	// Unmarshal the tokens from tokensFile.
	json.Unmarshal(tokensFile, &token)

	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + token.DiscordToken)
	if err != nil {
		log.Fatal("ERROR CREATING DISCORD SESSION: ", err)
	}

	// Identify that we want all intents.
	session.Identify.Intents = discordgo.IntentsAll

	// Now we open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatal("ERROR OPENING WEBSOCKET: ", err)
	}

	// Making a map of registered commands.
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	// Looping through the commands array and registering them.
	// https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.ApplicationCommandCreate
	for i, command := range commands {
		registered_command, err := session.ApplicationCommandCreate(session.State.User.ID, "1001077854936760352", command)
		if err != nil {
			log.Printf("CANNOT CREATE '%v' COMMAND: %v", command.Name, err)
		}
		registeredCommands[i] = registered_command
	}

	// Looping through the array of interaction handlers and adding them to the session.
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
			handler(session, interaction)
		}
	})

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Lopping through the registeredCommands array and deleting all the commands.
	for _, v := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "1001077854936760352", v.ID)
		if err != nil {
			log.Printf("CANNOT DELETE '%v' COMMAND: %v", v.Name, err)
		}
	}

	// Cleanly close down the Discord session.
	session.Close()
}
