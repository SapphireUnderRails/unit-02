package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// A map of command handlers for interactions.
var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"register": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID

		// Creating a query to insert the user into the database with a phony unix timestamp.
		query := fmt.Sprintf(`INSERT INTO users_on_cooldown(user_id, unix_timestamp) VALUES("%s", "%d");`,
			authorID, 0)

		// Executing that query.
		result, err := db.Exec(query)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT PLACE USER IN COOLDOWN DATABASE: %v", Red, Reset, err)
			return
		}
		log.Printf("%vSUCCESS%v - PLACED USER INTO COOLDOWN DATABASE: %v", Green, Reset, result)

		// Creating a query to insert the user into the database and give them an inital credit amount of 0.
		query = fmt.Sprintf(`INSERT INTO users_currency(user_id, credits) VALUES("%s", "0");`, authorID)

		// Executing that query.
		result, err = db.Exec(query)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT PLACE USER IN WALLET DATABASE: %v", Red, Reset, err)
			return
		}
		log.Printf("%vSUCCESS%v - PLACED USER INTO WALLET DATABASE: %v", Green, Reset, result)

		// Notify the user that they are now registerd.
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You are now registered to play!",
			},
		})
	},
	"daily": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID
		current_timestamp := time.Now().Unix()
		var database_timestamp int64
		var credits int64

		// Perform a single row query in the database to retrieve the timestamp.
		query := fmt.Sprintf(`SELECT unix_timestamp FROM users_on_cooldown WHERE user_id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&database_timestamp)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER'S TIMESTAMP FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		// Checking to see if the user is on cooldown or if it is just an outdated entry.
		if current_timestamp >= database_timestamp+int64(86400) {
			// It was an outdated entry, so we should give the user their reward and place them on cooldown again.

			// Updating the timestamp in the database so that the user can't use the command again for a certain amount of time.
			query = fmt.Sprintf(`UPDATE users_on_cooldown SET unix_timestamp = %v WHERE user_id = %v;`, current_timestamp, authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE UNIX TIMESTAMP IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER COOLDOWN: %v", Green, Reset, result)

			// Snagging the amount of credits so that they can be updated.
			query := fmt.Sprintf(`SELECT credits FROM users_currency WHERE user_id = %v;`, authorID)
			err = db.QueryRow(query).Scan(&credits)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
				return
			}

			// Updating the amount of credits in the database for the user.
			query = fmt.Sprintf(`UPDATE users_currency SET credits = %v WHERE user_id = %v;`, credits+int64(100), authorID)
			result, err = db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Here's your daily reward of 100 credits!",
				},
			})
		} else {
			// The user is actually on cooldown so we should let them know to comeback later.
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Come back on <t:%v:D> at <t:%v:T> to claim your daily reward!",
						database_timestamp+int64(86400), database_timestamp+int64(86400)),
				},
			})
		}
	},
}
