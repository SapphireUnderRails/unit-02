package main

import "github.com/bwmarrin/discordgo"

// The list of commands for the bot.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "register",
		Description: "This command registers you to play!",
	},
	{
		Name:        "daily",
		Description: "This command claims your daily credits!",
	},
}
