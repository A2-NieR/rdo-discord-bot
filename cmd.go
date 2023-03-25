package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Player struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `bson:"name"`
	DiscordId  string             `bson:"discord_id"`
	RockstarId string             `bson:"rockstar_id"`
	Bounty     string             `bson:"bounty"`
	Camp       string             `bson:"camp"`
	Footer     string             `bson:"footer"`
	Online     bool               `bson:"online"`
	Platform   string             `bson:"platform"`
	Time       time.Time          `bson:"time"`
	Expires    time.Time          `bson:"expires"`
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "setup",
			Description: "Initial Red Dead Online profile setup.",
		},
		{
			Name:        "me",
			Description: "Show and edit your current profile info.",
		},
		{
			Name:        "online",
			Description: "Flag yourself as online in this channel.",
		},
		{
			Name:        "offline",
			Description: "Flag yourself as offline in this channel.",
		},
		{
			Name:        "show",
			Description: "See who is currently online.",
		},
	}

	onlineControlButtons = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Set Bounty",
					Style:    discordgo.PrimaryButton,
					CustomID: "set_bounty",
				},
				discordgo.Button{
					Label:    "Set Camp",
					Style:    discordgo.PrimaryButton,
					CustomID: "set_camp",
				},
				discordgo.Button{
					Label:    "Set Footer",
					Style:    discordgo.PrimaryButton,
					CustomID: "set_footer",
				},
				discordgo.Button{
					Label:    "Show Players",
					Style:    discordgo.PrimaryButton,
					CustomID: "show_players",
				},
				discordgo.Button{
					Label:    "Go Offline",
					Style:    discordgo.DangerButton,
					CustomID: "go_offline",
				},
			},
		},
	}
	profileControlButtons = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Set Bounty",
					Style:    discordgo.SecondaryButton,
					CustomID: "set_bounty",
				},
				discordgo.Button{
					Label:    "Set Camp",
					Style:    discordgo.SecondaryButton,
					CustomID: "set_camp",
				},
				discordgo.Button{
					Label:    "Set Footer",
					Style:    discordgo.SecondaryButton,
					CustomID: "set_footer",
				},
				discordgo.Button{
					Label:    "Set R* ID",
					Style:    discordgo.SecondaryButton,
					CustomID: "set_rid",
				},
			},
		},
	}

	commandHandlers = map[string]func(b *Bot, i *discordgo.InteractionCreate){
		"setup": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used /setup in channel " + i.ChannelID)
			err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					Title:   "Profile Setup",
					Content: "Enter your current data to get you started. \nYou only have to do this once or after the bot was offline.",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "rid_input_" + i.Member.User.ID,
									Label:       "R* ID:",
									Style:       discordgo.TextInputShort,
									Placeholder: "123456789",
									Required:    false,
									MinLength:   9,
									MaxLength:   9,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "bounty_input_" + i.Member.User.ID,
									Label:       "Bounty (0-100):",
									Style:       discordgo.TextInputShort,
									Placeholder: "19.99",
									Required:    false,
									MinLength:   1,
									MaxLength:   5,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "foot_input_" + i.Member.User.ID,
									Label:       "Footer Message:",
									Style:       discordgo.TextInputShort,
									Placeholder: "What are you up to?",
									Required:    false,
									MaxLength:   42,
								},
							},
						},
					},
					Flags:    discordgo.MessageFlagsEphemeral,
					CustomID: "setup_" + i.Member.User.ID,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"me": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used /me in channel " + i.ChannelID)
			var result Player
			avatarURL := rdoAvatarUnknownURL
			rockstarIdStatus := "R* ID is not set"
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}

			err := b.Collection.FindOne(context.TODO(), filter).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You have not set up your profile. \nPlease use </setup:" + b.setupCommandID + "> to start. 🤠",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						b.ErrorReport.Notify(err, nil)
						log.Println(err)
					}
				}
			}
			if result.RockstarId != "" {
				rockstarIdStatus = "R* ID is set"
				avatarURL = strings.Join([]string{rdoAvatarURLPrefix, result.RockstarId, rdoAvatarURLSuffix}, "")
			}
			err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Type:        discordgo.EmbedTypeRich,
							Title:       "Your current profile data:",
							Description: rockstarIdStatus + "\n Camp: " + result.Camp + "\n Bounty: $" + result.Bounty + "\n Footer: " + result.Footer,
							Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: avatarURL},
						},
					},
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}

			_, err = b.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: "Update your profile data below:\n\nTo find your R* ID, visit your Social Club profile here: <https://socialclub.rockstargames.com/games/rdr2/overview>.\nOn the tiny avatar of your character do a right-click and click on *Open image in new tab*. In the browser address bar you will notice a 9-digit number (just before */pedshot_0.jpg*) which is your Rockstar ID.\n", Components: profileControlButtons, Flags: discordgo.MessageFlagsEphemeral})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"online": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used /online in channel " + i.ChannelID)
			if i.ChannelID == b.pcChannelID || i.ChannelID == b.playstationChannelID || i.ChannelID == b.xboxChannelID {
				var result Player
				avatarURL := rdoAvatarUnknownURL
				platform := "-"
				filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}

				err := b.Collection.FindOne(context.TODO(), filter).Decode(&result)
				if err != nil {
					if err == mongo.ErrNoDocuments {
						err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{
								Content: "You have not set up your profile. \nPlease use </setup:" + b.setupCommandID + "> to start. 🤠",
								Flags:   discordgo.MessageFlagsEphemeral,
							},
						})
						if err != nil {
							b.ErrorReport.Notify(err, nil)
							log.Println(err)
						}
					}
				}

				switch i.ChannelID {
				case b.pcChannelID:
					platform = "PC"
				case b.playstationChannelID:
					platform = "PS4"
				case b.xboxChannelID:
					platform = "XBOX"
				}

				playerOnline := bson.M{
					"$set": bson.D{
						{Key: "online", Value: true},
						{Key: "platform", Value: platform},
						{Key: "time", Value: time.Now().Format(time.RFC3339)},
						{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)},
					},
				}

				err = b.Collection.FindOneAndUpdate(context.TODO(), filter, playerOnline).Decode(&result)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}

				onlineData := []*discordgo.MessageEmbedField{
					{
						Name:   "Bounty:",
						Value:  "$" + result.Bounty,
						Inline: true,
					},
					{
						Name:   "Camp:",
						Value:  result.Camp,
						Inline: true,
					},
					{
						Name:   "Platform:",
						Value:  platform,
						Inline: true,
					},
				}

				if result.RockstarId != "" {
					avatarURL = strings.Join([]string{rdoAvatarURLPrefix, result.RockstarId, rdoAvatarURLSuffix}, "")
				}

				err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							{
								Type:      discordgo.EmbedTypeRich,
								Color:     colorGreen,
								Title:     result.Name + " is now online.",
								Thumbnail: &discordgo.MessageEmbedThumbnail{URL: avatarURL},
								Fields:    onlineData,
								Footer: &discordgo.MessageEmbedFooter{
									Text: result.Footer,
								},
							},
						},
					},
				})
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}

				_, err = b.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: "Quick Controls for your online session:", Components: onlineControlButtons, Flags: discordgo.MessageFlagsEphemeral})
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			} else {
				err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Please use the `/online` command only in:\n<#" + b.pcChannelID + ">\n<#" + b.playstationChannelID + ">\n<#" + b.xboxChannelID + ">",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}
		},
		"offline": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used /offline in channel " + i.ChannelID)
			var result Player
			avatarURL := rdoAvatarUnknownURL
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}
			playerOffline := bson.M{
				"$set": bson.D{
					{Key: "online", Value: false},
					{Key: "time", Value: time.Now().Format(time.RFC3339)},
					{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)},
				},
			}

			err := b.Collection.FindOneAndUpdate(context.TODO(), filter, playerOffline).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You have not set up your profile. \nPlease use </setup:" + b.setupCommandID + "> to start. 🤠",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						b.ErrorReport.Notify(err, nil)
						log.Println(err)
					}
				}
			}

			if result.RockstarId != "" {
				avatarURL = strings.Join([]string{rdoAvatarURLPrefix, result.RockstarId, rdoAvatarURLSuffix}, "")
			}

			err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Type:      discordgo.EmbedTypeRich,
							Color:     colorRed,
							Title:     result.Name + " is now offline.",
							Thumbnail: &discordgo.MessageEmbedThumbnail{URL: avatarURL},
						},
					},
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"show": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used /show in channel " + i.ChannelID)
			avatarURL := rdoAvatarUnknownURL
			listPlayers := ""
			bounty := "-"
			camp := "-"
			playTime := "-"
			footer := ""
			var cursor *mongo.Cursor
			var err error
			var results []Player
			playerList := []*discordgo.MessageEmbed{}

			opts := options.Find().SetSort(bson.D{{Key: "time", Value: 1}})

			switch i.ChannelID {
			case b.pcChannelID:
				cursor, err = b.Collection.Find(context.TODO(), bson.D{{Key: "platform", Value: "PC"}, {Key: "online", Value: true}}, opts)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			case b.playstationChannelID:
				cursor, err = b.Collection.Find(context.TODO(), bson.D{{Key: "platform", Value: "PS4"}, {Key: "online", Value: true}}, opts)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			case b.xboxChannelID:
				cursor, err = b.Collection.Find(context.TODO(), bson.D{{Key: "platform", Value: "XBOX"}, {Key: "online", Value: true}}, opts)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}

			if err = cursor.All(context.TODO(), &results); err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}

			if len(results) == 0 {
				playerList = []*discordgo.MessageEmbed{{
					Type:        discordgo.EmbedTypeRich,
					Color:       colorDark,
					Description: "There are no players online at the moment.",
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: rdoAvatarUnknownURL,
					},
				},
				}
			} else {
				for _, player := range results {
					if player.RockstarId != "" {
						avatarURL = strings.Join([]string{rdoAvatarURLPrefix, player.RockstarId, rdoAvatarURLSuffix}, "")
					}
					if player.Camp != "" {
						camp = player.Camp
					}
					if player.Bounty != "" {
						bounty = player.Bounty
					}
					if player.Footer != "" {
						footer = player.Footer
					}
					playTime = time.Since(player.Time).Truncate(time.Second).String()

					playerList = append(playerList, &discordgo.MessageEmbed{
						Type:  discordgo.EmbedTypeRich,
						Color: colorGrey,
						Title: player.Name,
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: avatarURL,
						},
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Bounty:",
								Value:  "$" + bounty,
								Inline: true,
							},
							{
								Name:   "Camp:",
								Value:  camp,
								Inline: true,
							},
							{
								Name:   "Online:",
								Value:  playTime,
								Inline: true,
							},
						},
						Footer: &discordgo.MessageEmbedFooter{Text: footer},
					})
				}
			}

			err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Title:   "Online Players",
					Content: listPlayers,
					Flags:   discordgo.MessageFlagsEphemeral,
					Embeds:  playerList,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
	}

	buttonHandlers = map[string]func(b *Bot, i *discordgo.InteractionCreate){
		"set_bounty": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used button set_bounty in channel " + i.ChannelID)
			err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					Title: "Set Bounty",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "bounty_input_" + i.Member.User.ID,
									Label:       "Set your current bounty (0-100):",
									Style:       discordgo.TextInputShort,
									Placeholder: "10.01",
									Required:    true,
									MinLength:   1,
									MaxLength:   5,
								},
							},
						},
					},
					Flags:    discordgo.MessageFlagsEphemeral,
					CustomID: "set_bounty_" + i.Member.User.ID,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"set_camp": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used button set_camp in channel " + i.ChannelID)
			selectMinVal := 1

			err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Set your current camp location.\nYour profile will be updated as soon as you select an option.",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									MenuType:    3,
									Placeholder: "Choose Location",
									MinValues:   &selectMinVal,
									MaxValues:   1,
									CustomID:    "camp_selection",
									Options: []discordgo.SelectMenuOption{
										{
											Label: "Bayou Nwa",
											Value: "Bayou Nwa",
										},
										{
											Label: "Big Valley",
											Value: "Big Valley",
										},
										{
											Label: "Cholla Springs",
											Value: "Cholla Springs",
										},
										{
											Label: "Cumberland Forest",
											Value: "Cumberland Forest",
										},
										{
											Label: "Gaptooth Ridge",
											Value: "Gaptooth Ridge",
										},
										{
											Label: "Great Plains",
											Value: "Great Plains",
										},
										{
											Label: "Grizzlies",
											Value: "Grizzlies",
										},
										{
											Label: "Heartlands",
											Value: "Heartlands",
										},
										{
											Label: "Hennigan's Stead",
											Value: "Hennigan's Stead",
										},
										{
											Label: "Rio Bravo",
											Value: "Rio Bravo",
										},
										{
											Label: "Roanoke Ridge",
											Value: "Roanoke Ridge",
										},
										{
											Label: "Scarlett Meadows",
											Value: "Scarlett Meadows",
										},
										{
											Label: "Tall Trees",
											Value: "Tall Trees",
										},
									},
								},
							},
						},
					},
					Flags:    discordgo.MessageFlagsEphemeral,
					CustomID: "select_camp_" + i.Member.User.ID,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"set_footer": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used button set_footer in channel " + i.ChannelID)
			err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					Title:   "Footer Message",
					Content: "Enter a message that appears in the footer of your online notification.",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "footer_input_" + i.Member.User.ID,
									Label:       "Set your footer message",
									Style:       discordgo.TextInputShort,
									Placeholder: "What are you up to?",
									Required:    false,
									MaxLength:   42,
								},
							},
						},
					},
					Flags:    discordgo.MessageFlagsEphemeral,
					CustomID: "set_footer_" + i.Member.User.ID,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"show_players": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used button show_players in channel " + i.ChannelID)
			avatarURL := rdoAvatarUnknownURL
			listPlayers := ""
			bounty := "-"
			camp := "-"
			playTime := "-"
			footer := ""
			var cursor *mongo.Cursor
			var err error
			var results []Player
			playerList := []*discordgo.MessageEmbed{}

			opts := options.Find().SetSort(bson.D{{Key: "time", Value: 1}})

			switch i.ChannelID {
			case b.pcChannelID:
				cursor, err = b.Collection.Find(context.TODO(), bson.D{{Key: "platform", Value: "PC"}, {Key: "online", Value: true}}, opts)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			case b.playstationChannelID:
				cursor, err = b.Collection.Find(context.TODO(), bson.D{{Key: "platform", Value: "PS4"}, {Key: "online", Value: true}}, opts)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			case b.xboxChannelID:
				cursor, err = b.Collection.Find(context.TODO(), bson.D{{Key: "platform", Value: "XBOX"}, {Key: "online", Value: true}}, opts)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}

			if err = cursor.All(context.TODO(), &results); err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}

			if len(results) == 0 {
				playerList = []*discordgo.MessageEmbed{{
					Type:        discordgo.EmbedTypeRich,
					Color:       colorDark,
					Description: "There are no players online at the moment.",
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: rdoAvatarUnknownURL,
					},
				},
				}
			} else {
				for _, player := range results {
					if player.RockstarId != "" {
						avatarURL = strings.Join([]string{rdoAvatarURLPrefix, player.RockstarId, rdoAvatarURLSuffix}, "")
					}
					if player.Camp != "" {
						camp = player.Camp
					}
					if player.Bounty != "" {
						bounty = player.Bounty
					}
					if player.Footer != "" {
						footer = player.Footer
					}
					playTime = time.Since(player.Time).Truncate(time.Second).String()

					playerList = append(playerList, &discordgo.MessageEmbed{
						Type:  discordgo.EmbedTypeRich,
						Color: colorGrey,
						Title: player.Name,
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: avatarURL,
						},
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Bounty:",
								Value:  "$" + bounty,
								Inline: true,
							},
							{
								Name:   "Camp:",
								Value:  camp,
								Inline: true,
							},
							{
								Name:   "Online:",
								Value:  playTime,
								Inline: true,
							},
						},
						Footer: &discordgo.MessageEmbedFooter{Text: footer},
					})
				}
			}

			err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Title:   "Online Players",
					Content: listPlayers,
					Flags:   discordgo.MessageFlagsEphemeral,
					Embeds:  playerList,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"go_offline": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used button go_offline in channel " + i.ChannelID)
			var result Player
			avatarURL := rdoAvatarUnknownURL
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}
			playerOffline := bson.M{
				"$set": bson.D{
					{Key: "online", Value: false},
					{Key: "time", Value: time.Now().Format(time.RFC3339)},
					{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)},
				},
			}

			err := b.Collection.FindOneAndUpdate(context.TODO(), filter, playerOffline).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You have not set up your profile. \nPlease use </setup:" + b.setupCommandID + "> to start. 🤠",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						b.ErrorReport.Notify(err, nil)
						log.Println(err)
					}
				}
			}

			if result.RockstarId != "" {
				avatarURL = strings.Join([]string{rdoAvatarURLPrefix, result.RockstarId, rdoAvatarURLSuffix}, "")
			}

			err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Type:      discordgo.EmbedTypeRich,
							Color:     colorRed,
							Title:     result.Name + " is now offline.",
							Thumbnail: &discordgo.MessageEmbedThumbnail{URL: avatarURL},
						},
					},
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
		"set_rid": func(b *Bot, i *discordgo.InteractionCreate) {
			log.Println(i.Member.User.Username + " used button set_rid in channel " + i.ChannelID)
			err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					Title: "Set Rockstar ID",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "rid_input_" + i.Member.User.ID,
									Label:       "Copy & Paste your R* ID:",
									Style:       discordgo.TextInputShort,
									Placeholder: "123456789",
									Required:    false,
									MaxLength:   9,
								},
							},
						},
					},
					Flags:    discordgo.MessageFlagsEphemeral,
					CustomID: "set_rid_" + i.Member.User.ID,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		},
	}
)

func (b *Bot) registerCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(b, i)
		}
	case discordgo.InteractionMessageComponent:
		if strings.HasPrefix(i.MessageComponentData().CustomID, "camp_selection") {
			camp := i.MessageComponentData().Values[0]

			var result Player
			player := bson.M{
				"$set": bson.D{
					{Key: "camp", Value: strings.Trim(camp, " ")},
					{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)},
				},
			}
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}

			err := b.Collection.FindOneAndUpdate(context.TODO(), filter, player).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}

			err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your camp location is now set to **" + camp + "**",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		}

		if h, ok := buttonHandlers[i.MessageComponentData().CustomID]; ok {
			h(b, i)
		}
	case discordgo.InteractionModalSubmit:
		modalData := i.ModalSubmitData()
		if strings.HasPrefix(modalData.CustomID, "setup") {
			rockstarId := modalData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			bounty := modalData.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			footer := modalData.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			var result Player
			var player bson.D
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}

			player = append(player, bson.E{Key: "discord_id", Value: i.Member.User.ID})
			if i.Member.Nick != "" {
				player = append(player, bson.E{Key: "name", Value: i.Member.Nick})
			} else {
				player = append(player, bson.E{Key: "name", Value: i.Member.User.Username})
			}
			player = append(player, bson.E{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)})

			if rockstarId != "" {
				player = append(player, bson.E{Key: "rockstar_id", Value: strings.Trim(rockstarId, " ")})
			}
			if bounty != "" {
				player = append(player, bson.E{Key: "bounty", Value: strings.Trim(bounty, " ")})
			}
			if footer != "" {
				player = append(player, bson.E{Key: "footer", Value: strings.Trim(footer, " ")})
			}

			err := b.Collection.FindOne(context.TODO(), filter).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					_, err = b.Collection.InsertOne(context.TODO(), player)
					if err != nil {
						b.ErrorReport.Notify(err, nil)
						log.Println(err)
					}

					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Success! Your initial profile info is now set. You can now go online, offline and show other online players.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						b.ErrorReport.Notify(err, nil)
						log.Println(err)
					}
				}
			} else {
				playerUpdate := bson.M{
					"$set": player,
				}

				_, err = b.Collection.UpdateOne(context.TODO(), filter, playerUpdate)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Success! Your profile has been updated.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}
		} else if strings.HasPrefix(modalData.CustomID, "set_footer") {
			footer := modalData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			var result Player
			player := bson.M{
				"$set": bson.D{
					{Key: "footer", Value: strings.Trim(footer, " ")},
					{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)},
				},
			}
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}

			err := b.Collection.FindOneAndUpdate(context.TODO(), filter, player).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your footer message is set. Feel free to change it anytime.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		} else if strings.HasPrefix(modalData.CustomID, "set_bounty") {
			bounty := modalData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			var result Player
			player := bson.M{
				"$set": bson.D{
					{Key: "bounty", Value: strings.Trim(bounty, " ")},
					{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)},
				},
			}
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}

			err := b.Collection.FindOneAndUpdate(context.TODO(), filter, player).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your bounty is now set to **$" + bounty + "**",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		} else if strings.HasPrefix(modalData.CustomID, "set_rid") {
			rockstarId := modalData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			var result Player
			player := bson.M{
				"$set": bson.D{
					{Key: "rockstar_id", Value: rockstarId},
					{Key: "expires", Value: time.Now().Add(time.Hour * 24 * 365)},
				},
			}
			filter := bson.D{{Key: "discord_id", Value: i.Member.User.ID}}

			err := b.Collection.FindOneAndUpdate(context.TODO(), filter, player).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Successfully updated your Rockstar ID.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		}
	}
}
