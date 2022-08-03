package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var onCooldown = map[string]int64{}

var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"test": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID
		timestamp := time.Now().Unix()

		fmt.Println(stringInKeys(authorID, onCooldown))

		// Checking if the user is in the cool down list.
		if stringInKeys(authorID, onCooldown) {
			// The user is in the cooldown list, but are they actually on cooldown? Let's check.
			if timestamp > timestamp+int64(time.Minute) {
				// The user actually is not on cool down, so we can give them their daily reward.
				// Rough implementation at the moment could spice it up to use hours and minutes in the exact response.

				//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Here's your daily reward!",
					},
				})
			} else if timestamp < timestamp+int64(time.Minute) {
				// The user is not on cooldown at the moment, give them their daily reward.
				//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Sorry, you're on cool down at the moment! Come back later!",
					},
				})

				// Now we need to place the user on cooldown.
				onCooldown[authorID] = timestamp
			}
		} else {
			// The user is not in the cooldown list, so we can go ahead and give them their daily reward.
			//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Here's your daily reward!",
				},
			})

			// Now we need to put the user on the cooldown list.
			onCooldown[authorID] = timestamp
		}

		fmt.Println(onCooldown)
	},
}
