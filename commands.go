package main

import (
	"github.com/bwmarrin/discordgo"
)

// Permissions for commands.
var manageServerPermission int64 = discordgo.PermissionManageServer

// The list of commands for the bot.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "add_cards",
		Description:              "Loops through './Card Art' folder and registers all the cards in there.",
		DefaultMemberPermissions: &manageServerPermission,
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
		Name:        "credits",
		Description: "This command tells you how many credits you have.",
	},
	{
		Name:        "characters",
		Description: "This command lists the available gacha pools to pull from.",
	},
	{
		Name:        "single_pull",
		Description: "This command pulls one random card from the gacha pool.",

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "The name of the character you wish to draw for.",
				Required:    false,
			},
		},
	},
	{
		Name:        "ten_pull",
		Description: "This command pulls ten random cards from the gacha pool.",

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "The name of the character you wish to draw for.",
				Required:    false,
			},
		},
	},
	{
		Name:        "list",
		Description: "This command lists the cards in your collection.",

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "The name of the character you wish to list.",
				Required:    false,
			},
		},
	},
	{
		Name:        "rename_card",
		Description: "This command renames a card.",

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "old_name",
				Description: "The old name of the card that you wish to rename.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "new_name",
				Description: "The new name of the card that you wish to rename.",
				Required:    true,
			},
		},
	},
}
