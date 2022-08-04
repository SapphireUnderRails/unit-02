package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// A map of command handlers for interactions.
var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"test": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID
		current_timestamp := time.Now().Unix()
		var database_timestamp int64

		// Perform a single row query in the database to retrieve the timestamp.
		query := fmt.Sprintf(`SELECT unix_timestamp FROM users_on_cooldown WHERE EXISTS(SELECT user_id FROM users_on_cooldown WHERE user_id = %s);`, authorID)
		err := db.QueryRow(query).Scan(&database_timestamp)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER'S TIMESTAMP FROM DATABASE:\n\t%v", Red, Reset, err)

			// If this is true then it means that the user was not in the database and we need to place them in there.
			if err.Error() == "sql: no rows in result set" {
				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintln("Here's your daily reward!"),
					},
				})

				// Creating a query to inser the user into the database and place them on cooldown.
				query = fmt.Sprintf(`INSERT INTO users_on_cooldown(user_id, unix_timestamp) VALUES("%s", "%d");`,
					authorID, current_timestamp)

				// Executing that query.
				result, err := db.Exec(query)
				if err != nil {
					log.Printf("%vERROR%v - COULD NOT PLACE USER IN DATABASE: %v", Red, Reset, err)
					return
				}
				log.Printf("%vSUCCESS%v - PLACED USER INTO DATABASE AND ON COOLDOWN: %v", Green, Reset, result)
			}
		} else {
			// Checking to see if the user is on cooldown or if it is just an outdated entry.
			if current_timestamp >= database_timestamp+int64(180) {
				// It was an outdated entry, so we should give the user their reward and place them on cooldown again.
				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Here's your daily reward!",
					},
				})
				// Now we need to update the timestamp in the database so that the user
				// can't use the command again for a certain amount of time.
				query = fmt.Sprintf(`UPDATE users_on_cooldown SET unix_timestamp = %v WHERE user_id = %v;`, current_timestamp, authorID)
				result, err := db.Exec(query)
				if err != nil {
					log.Printf("%vERROR%v - COULD NOT UPDATE UNIX TIMESTAMP IN DATABASE: %v", Red, Reset, err)
				}
				log.Printf("%vSUCCESS%v - UPDATED USER COOLDOWN: %v", Green, Reset, result)
			} else {
				// The user is actually on cooldown so we should let them know to comeback later.
				// This is a rough implementation so we should probably display the time and date the user should come back at.
				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Come back on <t:%v:D> at <t:%v:T> to claim your daily reward!",
							database_timestamp+int64(86400), database_timestamp+int64(86400)),
					},
				})
			}
		}
	},
}
