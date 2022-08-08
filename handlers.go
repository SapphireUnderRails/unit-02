package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// A map of command handlers for interactions.
var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"add_cards": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Getting all the files in the directory.
		filesList, err := os.ReadDir("./Card Art")
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT LIST CARDS: %v", Red, Reset, err)
			return
		}

		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Now registering %d cards...", len(filesList)),
			},
		})

		for _, file := range filesList {
			// Grabbing the image file path.
			filePath := fmt.Sprintf("./Card Art/%v", file.Name())

			// Reading the file into memory.
			imageBytes, err := os.Open(filePath)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT READ IMAGE: %v", Red, Reset, err)
				return
			}

			// Uploading that image to discord for saving.
			msg, err := session.ChannelFileSend(interaction.ChannelID, file.Name(), imageBytes)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPLOAD IMAGE: %v", Red, Reset, err)
				return
			}

			// Getting all the variables for the cards.
			name := strings.ReplaceAll(file.Name(), ".png", "")
			nameParts := strings.Split(name, " ")
			log.Println(nameParts)

			var character string
			switch nameParts[0] {
			case "SG01":
				character = "Hibiki"
			case "SG02":
				character = "Tsubasa"
			case "SG03":
				character = "Chris"
			case "SG04":
				character = "Maria"
			case "SG05":
				character = "Shirabe"
			case "SG06":
				character = "Kirika"
			case "SG07":
				character = "Kanade"
			case "SG08":
				character = "Miku"
			case "SG09":
				character = "Serena"
				// case "SG10":
				// 	character = "Fine"
				// case "SG11":
				// 	character = "Dr.Ver"
				// case "SG12":
				// 	character = "Genjuro"
				// case "SG13":
				// 	character = "Ogawa"
				// case "SG14":
				// 	character = "St. Germain"
				// case "SG15":
				// 	character = "Cagliostro"
				// case "SG16":
				// 	character = "Prelati"
				// case "SG18":
				// 	character = "Adam"
				// case "SG19":
				// 	character = "Carol"
				// case "SG21":
				// 	character = "Phara"
				// case "SG22":
				// 	character = "Leiur"
				// case "SG23":
				// 	character = "Garie"
				// case "SG24":
				// 	character = "Micha"
				// case "SG25":
				// 	character = "Andou"
				// case "SG26":
				// 	character = "Shiori"
				// case "SG27":
				// 	character = "Yumi"
				// case "SG28":
				// 	character = "Vanessa"
				// case "SG29":
				// 	character = "Millaarc"
				// case "SG30":
				// 	character = "Elsa"
				// case "SG31":
				// 	character = "Shem-Ha"
			}

			cardID := fmt.Sprintf("%v_%v", nameParts[0], nameParts[1])
			evolution := nameParts[2]
			cardImage := msg.Attachments[0].URL

			// Craetiing a query to inser the cards into the card database.
			query := fmt.Sprintf(`INSERT INTO cards(character_name, card_id, evolution, card_image) VALUES("%v", "%v", %v, "%v");`,
				character, cardID, evolution, cardImage)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT REGISTER CARD IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - REGISTERED CARD IN CARD DATABASE: %v", Green, Reset, result)

			time.Sleep(time.Millisecond * 10)
		}
	},
	"register": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID

		// Creating a query to insert the user into the database with a phony unix timestamp and no credits.
		query := fmt.Sprintf(`INSERT INTO users_registration(user_id, unix_timestamp, credits) VALUES("%s", 0, 10000);`,
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
				Content: "Welcome to testing. You are now registered to play. Here's 10,000 credits to get you started!",
			},
		})
	},
	"daily": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID
		current_timestamp := time.Now().Unix()
		var id int64
		var database_timestamp int64
		var credits int64

		// Perform a single row query to make sure the user is registered.
		query := fmt.Sprintf(`SELECT id FROM users_registration WHERE user_id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER FROM REGISTRATION DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		if err == sql.ErrNoRows {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			return
		}

		// Perform a single row query in the database to retrieve the timestamp.
		query = fmt.Sprintf(`SELECT unix_timestamp FROM users_registration WHERE user_id = %s;`, authorID)
		err = db.QueryRow(query).Scan(&database_timestamp)
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
			query = fmt.Sprintf(`UPDATE users_registration SET credits = %v WHERE user_id = %v;`, credits+int64(150), authorID)
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
					Content: "Here's your daily reward of 150 credits!",
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
	"credits": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var id int64
		var credits int64
		authorID := interaction.Member.User.ID

		// Perform a single row query to make sure the user is registered.
		query := fmt.Sprintf(`SELECT id FROM users_registration WHERE user_id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER FROM REGISTRATION DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		if err == sql.ErrNoRows {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			return
		}

		// Perform a single row query to get the amount of credits a user has.
		query = fmt.Sprintf(`SELECT credits FROM users_registration WHERE user_id = %s;`, authorID)
		err = db.QueryRow(query).Scan(&credits)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE CREDITs FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("You currently have %d credits!", credits),
			},
		})
	},
	"characters": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var character string
		var characters []string
		// Creating a query to get distinct character names from the cards table.
		query := `SELECT DISTINCT character_name FROM cards;`
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE CHARACTERS FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		for rows.Next() {
			err := rows.Scan(&character)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT RETRIEVE CHARACTER FROM ROW:\n\t%v", Red, Reset, err)
				return
			}

			characters = append(characters, character)
		}

		// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: strings.Join(characters, ", "),
			},
		})
	},
	"single_pull": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var credits int64
		var drawnCardID string
		var evolution int8
		var customName string
		var cardImage string

		var id int64
		authorID := interaction.Member.User.ID

		// Perform a single row query to make sure the user is registered.
		query := fmt.Sprintf(`SELECT id FROM users_registration WHERE user_id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER FROM REGISTRATION DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		if err == sql.ErrNoRows {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			return
		}

		// Snagging the amount of credits so that they can be checked against.
		query = fmt.Sprintf(`SELECT credits FROM users_registration WHERE user_id = %v;`, authorID)
		err = db.QueryRow(query).Scan(&credits)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
			return
		}

		// Checking if the user has enough credits.
		if credits >= int64(200) {
			// Updating the amount of credits in the database for the user.
			query = fmt.Sprintf(`UPDATE users_registration SET credits = %v WHERE user_id = %v;`, credits-int64(200), authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

			// Performing a single row query to draw a card with optional character.
			var query string
			if len(interaction.ApplicationCommandData().Options) == 0 {
				query = `SELECT card_id, card_image FROM cards WHERE evolution = 1 ORDER BY RAND() LIMIT 1;`
			} else {
				query = fmt.Sprintf(`SELECT card_id, card_image FROM cards WHERE evolution = 1 AND character_name = "%v" ORDER BY RAND() LIMIT 1;`,
					strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()))
			}
			err = db.QueryRow(query).Scan(&drawnCardID, &cardImage)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("%vERROR%v - COULD NOT GET A CARD TO DRAW FROM DATABASE: %v", Red, Reset, err)
				return
			}

			if err == sql.ErrNoRows {
				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
						Content: "I couldn't find that character pool. Are you sure you spelled that character's name right?",
					},
				})

				return
			}

			// Performing a single row query to check if the user already has the card in their collection.
			query = fmt.Sprintf(`SELECT card_id, evolution FROM users_collection WHERE card_id = "%v" AND user_id = %v;`, drawnCardID, authorID)
			err = db.QueryRow(query).Scan(&drawnCardID, &evolution)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("%vERROR%v - COULD NOT QUERY USER COLLECTION IN DATABASE: %v", Red, Reset, err)
				return
			}

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
							Content: fmt.Sprintf("Whoah there! You've already maxed out your %v, I've refunded half your draw. Go ahead and try again!", customName),
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
							Content: fmt.Sprintf("Check it out! You've evolved your %v!", customName),
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
	"ten_pull": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var credits int64
		var drawnCardID string
		var evolution int8
		var customName string
		var cardImage string

		var id int64
		authorID := interaction.Member.User.ID

		// Perform a single row query to make sure the user is registered.
		query := fmt.Sprintf(`SELECT id FROM users_registration WHERE user_id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER FROM REGISTRATION DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		if err == sql.ErrNoRows {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			return
		}

		// Snagging the amount of credits so that they can be checked against.
		query = fmt.Sprintf(`SELECT credits FROM users_registration WHERE user_id = %v;`, authorID)
		err = db.QueryRow(query).Scan(&credits)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
			return
		}

		if credits >= int64(1500) {
			// Updating the amount of credits in the database for the user.
			query = fmt.Sprintf(`UPDATE users_registration SET credits = %v WHERE user_id = %v;`, credits-int64(1500), authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Now drawing ten cards for %v#%v...",
						interaction.Member.User.Username, interaction.Member.User.Discriminator),
				},
			})

			for i := 0; i < 10; i++ {
				// Performing a single row query to draw a card with optional character.
				var query string
				if len(interaction.ApplicationCommandData().Options) == 0 {
					query = `SELECT card_id, card_image FROM cards WHERE evolution = 1 ORDER BY RAND() LIMIT 1;`
				} else {
					query = fmt.Sprintf(`SELECT card_id, card_image FROM cards WHERE evolution = 1 AND character_name = "%v" ORDER BY RAND() LIMIT 1;`,
						strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()))
				}
				err = db.QueryRow(query).Scan(&drawnCardID, &cardImage)
				if err != nil && err != sql.ErrNoRows {
					log.Printf("%vERROR%v - COULD NOT GET A CARD TO DRAW FROM DATABASE: %v", Red, Reset, err)
					return
				}

				if err == sql.ErrNoRows {
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: "I couldn't find that character pool. Are you sure you spelled that character's name right?",
						},
					})

					return
				}

				// Performing a single row query to check if the user already has the card in their collection.
				query = fmt.Sprintf(`SELECT card_id, evolution FROM users_collection WHERE card_id = "%v" AND user_id = %v;`, drawnCardID, authorID)
				err = db.QueryRow(query).Scan(&drawnCardID, &evolution)
				if err != nil && err != sql.ErrNoRows {
					log.Printf("%vERROR%v - COULD NOT QUERY USER COLLECTION IN DATABASE: %v", Red, Reset, err)
					return
				}

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

					webhook := discordgo.WebhookParams{
						Embeds:  embeds,
						Content: fmt.Sprintf("(%d/10) You've drawn %v! I've added it to your collection.", i+1, drawnCardID),
					}

					// Informing the user that they have maxxed out the level on the card.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					session.FollowupMessageCreate(interaction.Interaction, true, &webhook)
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

						webhook := discordgo.WebhookParams{
							Embeds:  embeds,
							Content: fmt.Sprintf("(%d/10) Whoah there! You've already maxxed out your %v! I've given you 100 credits back.", i+1, customName),
						}

						// Informing the user that they have maxxed out the level on the card.
						// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
						session.FollowupMessageCreate(interaction.Interaction, true, &webhook)
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

						webhook := discordgo.WebhookParams{
							Embeds:  embeds,
							Content: fmt.Sprintf("(%d/10) Check it out! You've evolved your %v!", i+1, customName),
						}

						// Informing the user that they have maxxed out the level on the card.
						// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
						session.FollowupMessageCreate(interaction.Interaction, true, &webhook)
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

						webhook := discordgo.WebhookParams{
							Embeds:  embeds,
							Content: fmt.Sprintf("(%d/10) Check it out! You've evolved your %v!", i+1, customName),
						}

						// Informing the user that they have maxxed out the level on the card.
						// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
						session.FollowupMessageCreate(interaction.Interaction, true, &webhook)
					}
				}
				// Sleeping for a couple seconds to let the user see the card.
				time.Sleep(time.Second)
				time.Sleep(time.Second)
			}
		} else {
			// The user does not have enough credits...
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have enough credits to draw ten cards.",
				},
			})

			return
		}
	},
	"rename_card": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		oldName := interaction.ApplicationCommandData().Options[0].StringValue()
		newName := interaction.ApplicationCommandData().Options[1].StringValue()

		authorID := interaction.Member.User.ID
		var id int64

		// Perform a single row query to make sure the user is registered.
		query := fmt.Sprintf(`SELECT id FROM users_registration WHERE user_id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER FROM REGISTRATION DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		if err == sql.ErrNoRows {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			return
		}

		// Creating a query to rename the card in the user's collection.
		query = fmt.Sprintf(`UPDATE users_collection SET custom_name = "%v" WHERE custom_name = "%v";`, newName, oldName)

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
}
