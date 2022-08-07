package main

import (
	"github.com/bwmarrin/discordgo"
)

// Permissions for commands.
var manageServerPermission int64 = discordgo.PermissionManageServer

// The list of commands for the bot.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "add_card",
		Description:              "This command adds a card to the database.",
		DefaultMemberPermissions: &manageServerPermission,

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character_name",
				Description: "The name of the character on the card you wish to upload to the database.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "card_id",
				Description: "The ID of the card you wish to upload to the database.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "evolution",
				Description: "The evolution of the card you wish to upload to the database.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "au",
				Description: "Whether or not the card you wish to upload to the database is of an AU character.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "crossover",
				Description: "Whether or not the card you wish to upload to the database is a crossover unit.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "crossover_series",
				Description: "The series that the unit that you wish to upload to the database is crossed over with.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "theme",
				Description: "The theme of the unit that you wish to upload to the database is crossed over with.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "image",
				Description: "The image of the card you wish to upload to the database.",
				Required:    true,
			},
		},
	},
	{
		Name:        "register",
		Description: "This command registers you to play!",
	},
	{
		Name:        "daily",
		Description: "This command claims your daily credits!",
	},
	{
		Name:        "single_pull",
		Description: "This command pulls one random card from the gacha pool.",
	},
}
