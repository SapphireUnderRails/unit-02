package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Making a struct that will hold the user and the timestamp from when they used the message.
// type UserCooldown struct {
// 	userID        string
// 	unixTimeStamp int64
// }

// Making an array that will hold the list of users.
// var usersOnCooldown []UserCooldown

var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"test": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID
		current_timestamp := time.Now().Unix()

		// Check if the user exists in the database.
		query := fmt.Sprintf(`SELECT unix_timestamp FROM users_on_cooldown WHERE EXISTS(SELECT user_id FROM users_on_cooldown WHERE user_id = %v);`, authorID)
		row, err := db.Query(query)
		if err != nil {
			log.Println(err)
		}

		// Checking to see if any rows returned in query.
		if row.Next() {
			// There were, so advance through them, though there should only be one.
			for row.Next() {
				var unix_timestamp int64
				row.Scan(&unix_timestamp)

				// Is the user actually on cooldown or is it just an outdated entry? Let's check.
				if current_timestamp >= unix_timestamp+int64(60) {
					// It appears to be an outdated entry so, let's let the user claim their reward.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Here's your daily reward!",
						},
					})
				} else {
					// The user is on cool down. Fuck the user.
					// This is a rough implementation. Could probably change it up so that
					// it shows things like hours, minutes, and seconds remaining.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Fuck off, you're on cooldown.",
						},
					})

					return
				}
			}

			// Now we need to update the timestamp in the database so that the user
			// can't use the command again for a certain amount of time.
			query = fmt.Sprintf(`UPDATE users_on_cooldown SET unix_timestamp = %v WHERE user_id = %v;`, current_timestamp, authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Println("COULD NOT UPDATE UNIX TIMESTAMP: ", err)
			}
			log.Println(result)

		} else {
			// They weren't in the database, so fuck the user.
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintln("fuck you"),
				},
			})
		}
	},
}
