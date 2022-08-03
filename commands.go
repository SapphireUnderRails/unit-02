package main

import "github.com/bwmarrin/discordgo"

// The list of commands for the bot.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "test",
		Description: "This is just a test command!",
	},
}
