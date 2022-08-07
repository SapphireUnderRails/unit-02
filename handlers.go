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

		// Creating a query to insert the user into the database with a phony unix timestamp and no credits.
		query := fmt.Sprintf(`INSERT INTO users_registration(user_id, unix_timestamp, credits) VALUES("%s", 0, 0);`,
			authorID)

		// Executing that query.
		result, err := db.Exec(query)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT PLACE USER IN REGISTRATION DATABASE: %v", Red, Reset, err)
			return
		}
		log.Printf("%vSUCCESS%v - PLACED USER INTO REGISTRATION DATABASE: %v", Green, Reset, result)

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
		query := fmt.Sprintf(`SELECT unix_timestamp FROM users_registration WHERE user_id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&database_timestamp)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER'S TIMESTAMP FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		// Checking to see if the user is on cooldown or if it is just an outdated entry.
		if current_timestamp >= database_timestamp+int64(86400) {
			// It was an outdated entry, so we should give the user their reward and place them on cooldown again.

			// Updating the timestamp in the database so that the user can't use the command again for a certain amount of time.
			query = fmt.Sprintf(`UPDATE users_registration SET unix_timestamp = %v WHERE user_id = %v;`, current_timestamp, authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE UNIX TIMESTAMP IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER COOLDOWN: %v", Green, Reset, result)

			// Snagging the amount of credits so that they can be updated.
			query := fmt.Sprintf(`SELECT credits FROM users_registration WHERE user_id = %v;`, authorID)
			err = db.QueryRow(query).Scan(&credits)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
				return
			}

			// Updating the amount of credits in the database for the user.
			query = fmt.Sprintf(`UPDATE users_registration SET credits = %v WHERE user_id = %v;`, credits+int64(7000), authorID)
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
					Content: "Here's your daily reward!",
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
	"rename_card": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		oldName := interaction.ApplicationCommandData().Options[0].StringValue()
		newName := interaction.ApplicationCommandData().Options[1].StringValue()

		// Creating a query to rename the card in the user's collection.
		query := fmt.Sprintf(`UPDATE users_collection SET custom_name = "%v" WHERE custom_name = "%v";`, newName, oldName)

		// Executing that query.
		result, err := db.Exec(query)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT UPDATE USER'S CUSTOM NAME IN DATABASE: %v", Red, Reset, err)
			return
		}
		log.Printf("%vSUCCESS%v - UPDATED USER'S CUSTOM NAME IN DATABASE: %v", Green, Reset, result)

		// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully renamed your '%v' to '%v'.", oldName, newName),
			},
		})

	},
	"single_pull": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Snagging the amount of credits so that they can be checked against.
		var credits int64
		var drawnCardID string
		var evolution int8
		var customName string
		var cardImage string
		authorID := interaction.Member.User.ID

		query := fmt.Sprintf(`SELECT credits FROM users_registration WHERE user_id = %v;`, authorID)
		err := db.QueryRow(query).Scan(&credits)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
			return
		}

		// Checking if the user has enough credits.
		if credits >= int64(100) {
			// Updating the amount of credits in the database for the user.
			query = fmt.Sprintf(`UPDATE users_registration SET credits = %v WHERE user_id = %v;`, credits-int64(100), authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

			// Performing a single row query to draw a card.
			query := `SELECT card_id, card_image FROM cards WHERE evolution = 1 ORDER BY RAND() LIMIT 1;`
			err = db.QueryRow(query).Scan(&drawnCardID, &cardImage)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET A CARD TO DRAW FROM DATABASE: %v", Red, Reset, err)
				return
			}

			// Performing a single row query to check if the user already has the card.
			query = fmt.Sprintf(`SELECT card_id, evolution FROM users_collection WHERE card_id = "%v" AND user_id = %v;`, drawnCardID, authorID)
			err = db.QueryRow(query).Scan(&drawnCardID, &evolution)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("%vERROR%v - COULD QUERY USER COLLECTION IN DATABASE: %v", Red, Reset, err)
				return
			}

			// Does the drawn card exist in the users collection?
			if err == sql.ErrNoRows {
				// The user does not have the card in their collection.
				// Query to insert the card into the user's collection.
				query = fmt.Sprintf(`INSERT INTO users_collection(user_id, card_id, evolution, custom_name) VALUES("%v", "%v", "%v", "%v");`,
					authorID, drawnCardID, 1, drawnCardID)

				// Executing that query.
				result, err := db.Exec(query)
				if err != nil {
					log.Printf("%vERROR%v - COULD NOT PLACE CARD IN USER COLLECTION DATABASE: %v", Red, Reset, err)
					return
				}
				log.Printf("%vSUCCESS%v - PLACED CARD INTO USER COLLECTION DATABASE: %v", Green, Reset, result)

				// Constructing an embed to hold the card image.
				image := discordgo.MessageEmbedImage{
					URL: cardImage,
				}
				embeds := []*discordgo.MessageEmbed{
					{
						Image: &image,
					},
				}
				// Informing the user that they have collected the card.
				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
						Content: fmt.Sprintf("Successfully added %v to your collection.", drawnCardID),
						Embeds:  embeds,
					},
				})
			} else {
				// The user does have this card, the only question is what level do they have?
				if evolution == 3 {
					// If the evolution level is the max level, then we need to refund the user for this draw.
					// Updating the amount of credits in the database for the user.
					query = fmt.Sprintf(`UPDATE users_registration SET credits = %v WHERE user_id = %v;`, credits+int64(100), authorID)
					result, err := db.Exec(query)
					if err != nil {
						log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
						return
					}
					log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

					// Creating a query to retrieve the image of the maximum level card.
					query = fmt.Sprintf(`SELECT card_image FROM cards where card_id = "%v" AND evolution = %v`, drawnCardID, evolution)
					err = db.QueryRow(query).Scan(&cardImage)
					if err != nil {
						log.Printf("%vERROR%v - COULD QUERY CARDS IN DATABASE: %v", Red, Reset, err)
						return
					}

					// Creating a query to retrieve the user's custom name of the card.
					query = fmt.Sprintf(`SELECT custom_name FROM users_collection WHERE user_id = %v AND card_id = "%v"`, authorID, drawnCardID)
					err = db.QueryRow(query).Scan(&customName)
					if err != nil {
						log.Printf("%vERROR%v - COULD NOT GET CUSTOM NAME FROM DATABASE: %v", Red, Reset, err)
						return
					}

					// Constructing an embed to hold the card image.
					image := discordgo.MessageEmbedImage{
						URL: cardImage,
					}
					embeds := []*discordgo.MessageEmbed{
						{
							Image: &image,
						},
					}

					// Informing the user that they have no more levels to gain on the card.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: fmt.Sprintf("Whoah there! You've already maxed out your %v, I've refunded your draw. Go ahead and try again!", customName),
							Embeds:  embeds,
						},
					})
				} else if evolution == 2 {
					// If the evolution level is 2, then we need to evolve the card to level 3.

					// Updating the evolution level in the database of the user's card.
					query = fmt.Sprintf(`UPDATE users_collection SET evolution = 3 WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
					result, err := db.Exec(query)
					if err != nil {
						log.Printf("%vERROR%v - COULD NOT UPDATE USER EVOLUTION IN DATABASE: %v", Red, Reset, err)
						return
					}
					log.Printf("%vSUCCESS%v - UPDATED USER EVOLUTION: %v", Green, Reset, result)

					// Creating a query to retrieve the image of the maximum level card.
					query = fmt.Sprintf(`SELECT card_image FROM cards WHERE card_id = "%v" AND evolution = 3;`, drawnCardID)
					err = db.QueryRow(query).Scan(&cardImage)
					if err != nil {
						log.Printf("%vERROR%v - COULD QUERY CARDS IN DATABASE: %v", Red, Reset, err)
						return
					}

					// Creating a query to retrieve the user's custom name of the card.
					query = fmt.Sprintf(`SELECT custom_name FROM users_collection WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
					err = db.QueryRow(query).Scan(&customName)
					if err != nil {
						log.Printf("%vERROR%v - COULD NOT GET CUSTOM NAME FROM DATABASE: %v", Red, Reset, err)
						return
					}

					// Constructing an embed to hold the card image.
					image := discordgo.MessageEmbedImage{
						URL: cardImage,
					}
					embeds := []*discordgo.MessageEmbed{
						{
							Image: &image,
						},
					}

					// Informing the user that they have maxxed out the level on the card.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: fmt.Sprintf("Check it out! You've leveled up your %v!", customName),
							Embeds:  embeds,
						},
					})
				} else if evolution == 1 {
					// If the evolution level is 1, then we need to evolve the card to level 2.

					// Updating the evolution level in the database of the user's card.
					query = fmt.Sprintf(`UPDATE users_collection SET evolution = 2 WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
					result, err := db.Exec(query)
					if err != nil {
						log.Printf("%vERROR%v - COULD NOT UPDATE USER EVOLUTION IN DATABASE: %v", Red, Reset, err)
						return
					}
					log.Printf("%vSUCCESS%v - UPDATED USER EVOLUTION: %v", Green, Reset, result)

					// Creating a query to retrieve the image of the maximum level card.
					query = fmt.Sprintf(`SELECT card_image FROM cards where card_id = "%v" AND evolution = 2;`, drawnCardID)
					err = db.QueryRow(query).Scan(&cardImage)
					if err != nil {
						log.Printf("%vERROR%v - COULD QUERY CARDS IN DATABASE: %v", Red, Reset, err)
						return
					}

					// Creating a query to retrieve the user's custom name of the card.
					query = fmt.Sprintf(`SELECT custom_name FROM users_collection WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
					err = db.QueryRow(query).Scan(&customName)
					if err != nil {
						log.Printf("%vERROR%v - COULD NOT GET CUSTOM NAME FROM DATABASE: %v", Red, Reset, err)
						return
					}

					// Constructing an embed to hold the card image.
					image := discordgo.MessageEmbedImage{
						URL: cardImage,
					}
					embeds := []*discordgo.MessageEmbed{
						{
							Image: &image,
						},
					}

					// Informing the user that they have maxxed out the level on the card.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: fmt.Sprintf("Check it out! You've evolved your %v!", customName),
							Embeds:  embeds,
						},
					})
				}
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
