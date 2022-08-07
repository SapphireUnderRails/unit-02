package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// A map of command handlers for interactions.
var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"add_card": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Getting the image that was supposed to be uploaded.
		attachments := interaction.ApplicationCommandData().Resolved.Attachments
		var attachmentKey string

		for key := range attachments {
			attachmentKey = key
		}

		// Getting the name of the character on the card.
		characterName := interaction.ApplicationCommandData().Options[0].StringValue()

		// Getting the id of the card.
		cardID := interaction.ApplicationCommandData().Options[1].StringValue()

		// Getting the evolution of the card.
		evolution := interaction.ApplicationCommandData().Options[2].IntValue()

		// Getting whether or not the character on the card is an AU character.
		au := interaction.ApplicationCommandData().Options[4].BoolValue()
		auVar := boolToInt(au)

		// Getting whether or not the character on the card is crossover with a series.
		crossover := interaction.ApplicationCommandData().Options[4].BoolValue()
		crossoverVar := boolToInt(crossover)

		// Getting what series the character is crossed over with.
		crossoverSeries := interaction.ApplicationCommandData().Options[5].StringValue()

		// Getting the theme of the card.
		theme := interaction.ApplicationCommandData().Options[6].StringValue()

		// Createing a SQL query to register the card.
		query := fmt.Sprintf(`INSERT INTO cards(character_name, card_id, card_image, evolution, au, crossover, crossover_series, theme)
			VALUES("%v", "%v", "%v", "%v", "%v", "%v", "%v", "%v")`,
			characterName, cardID, attachments[attachmentKey].URL, evolution, auVar, crossoverVar, crossoverSeries, theme)

		// Executing that query.
		result, err := db.Exec(query)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT REGISTER CARD IN CARD DATABASE: %v", Red, Reset, err)
			return
		}
		log.Printf("%vSUCCESS%v - REGISTERED CARD INTO CARD DATABASE: %v", Green, Reset, result)

		// Letting the user know that the card is now registered.
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully registered %v %v", cardID, evolution),
			},
		})

	},
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
			query = fmt.Sprintf(`UPDATE users_currency SET credits = %v WHERE user_id = %v;`, credits+int64(7000), authorID)
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
	"single_pull": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Snagging the amount of credits so that they can be checked against.
		var credits int64
		var drawnCardID string
		var userCardID string
		authorID := interaction.Member.User.ID

		query := fmt.Sprintf(`SELECT credits FROM users_currency WHERE user_id = %v;`, authorID)
		err := db.QueryRow(query).Scan(&credits)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
			return
		}

		// Checking if the user has enough credits.
		if credits >= int64(100) {
			// Updating the amount of credits in the database for the user.
			query = fmt.Sprintf(`UPDATE users_currency SET credits = %v WHERE user_id = %v;`, credits-int64(100), authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

			// Performing a single row query to draw a card.
			query := `SELECT card_id FROM cards WHERE evolution = 1;`
			err = db.QueryRow(query).Scan(&drawnCardID)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET A CARD TO DRAW FROM DATABASE: %v", Red, Reset, err)
				return
			}

			// Performing a single row query to check if the user already has the card.
			query = fmt.Sprintf(`SELECT card_id FROM users_collection WHERE card_id = %v AND user_id = %v;`, drawnCardID, authorID)
			err = db.QueryRow(query).Scan(&userCardID)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("%vERROR%v - COULD QUERY USER COLLECTION IN DATABASE: %v", Red, Reset, err)
				return
			}

			// Does the drawn card exist in the users collection?
			if err == sql.ErrNoRows {
				// The user does not have the card in their collection.
				// Query to insert the card into the user's collection.

			}

		} else {
			// The user does not have enough credits...
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have enough credits to draw a card.",
				},
			})

			return
		}
	},
}
