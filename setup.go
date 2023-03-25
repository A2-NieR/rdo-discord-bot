package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

var (
	roleEmojis = make(map[string]string)
)

func (b *Bot) prepareServer(s *discordgo.Session, m *discordgo.Ready) {
	b.getChannelIDs()
	b.setupRoles()
	b.setupCommands()
	b.updateChangelog()
	log.Println("Initial setup complete. Bot is now ready and waiting...")
	fmt.Println("================================================================================")
}

func (b *Bot) getChannelIDs() {
	log.Println("Reading channels...")
	channels, err := b.Session.GuildChannels(b.GuildID)
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	for _, c := range channels {
		switch c.Name {
		case "general":
			b.generalChannelID = c.ID
		case "roles":
			b.rolesChannelID = c.ID
		case "commands":
			b.commandsChannelID = c.ID
		case "bulletin":
			b.bulletinChannelID = c.ID
		case "pc":
			b.pcChannelID = c.ID
		case "ps4":
			b.playstationChannelID = c.ID
		case "xbox-one":
			b.xboxChannelID = c.ID
		}
	}
}

func (b *Bot) setupRoles() {
	//Emojis need to be manually assigned for free servers
	roleEmojis["Bountyhunter"] = "⛓"
	roleEmojis["Trader"] = "🤝"
	roleEmojis["Collector"] = "🔮"
	roleEmojis["Moonshiner"] = "🥃"
	roleEmojis["Naturalist"] = "🌿"
	roleEmojis["PC"] = "💻"
	roleEmojis["PS4"] = "🅿"
	roleEmojis["XBOX"] = "❎"

	roleSelfAssignDescription := "React to this message to assign your roles:\n\n⛓ Bountyhunter \n\n🤝 Trader\n\n🔮 Collector\n\n🥃 Moonshiner\n\n🌿 Naturalist\n\n💻 PC\n\n🅿 Playstation\n\n❎ Xbox"

	log.Println("Reading server roles...")
	roles, err := b.Session.GuildRoles(b.GuildID)
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	for _, r := range roles {
		// Skipping @everyone & bot/application role
		if r.Name == "@everyone" || r.Name == b.BotRole {
			continue
		}

		// Store roles in map for self assignment
		guildRoles[r.Name] = &serverRole{ID: r.ID, Name: r.Name, Emoji: roleEmojis[r.Name]}
	}

	log.Println("Reading roles channel messages...")
	rolesChannelMessages, err := b.Session.ChannelMessages(b.rolesChannelID, 10, "", "", "")
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	roleMessageEmbed := &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       "Server Roles",
		Description: roleSelfAssignDescription,
		Color:       colorWhite,
	}

	if len(rolesChannelMessages) == 0 {
		log.Println("Writing role selfassignment message...")
		roleMessage, err := b.Session.ChannelMessageSendEmbed(b.rolesChannelID, roleMessageEmbed)
		if err != nil {
			b.ErrorReport.Notify(err, nil)
			log.Println(err)
		}

		// Assign bot to all roles to provide emoji-reactions
		for _, role := range guildRoles {
			err = b.Session.MessageReactionAdd(b.rolesChannelID, roleMessage.ID, role.Emoji)
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}

			log.Printf("Assign bot to role: %s", role.Name)
			err = b.Session.GuildMemberRoleAdd(b.GuildID, b.Session.State.User.ID, role.ID)
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		}
		b.roleSelfAssignMessageID = roleMessage.ID
	} else if rolesChannelMessages[0].Embeds[0].Description != roleSelfAssignDescription {
		log.Println("Updating role selfassignment message...")
		roleMessage, err := b.Session.ChannelMessageEditEmbed(b.rolesChannelID, rolesChannelMessages[0].ID, roleMessageEmbed)
		if err != nil {
			b.ErrorReport.Notify(err, nil)
			log.Println(err)
		}
		b.roleSelfAssignMessageID = roleMessage.ID
	} else {
		b.roleSelfAssignMessageID = rolesChannelMessages[0].ID
	}
}

func (b *Bot) setupCommands() {
	log.Println("Updating server commands...")
	registeredCommands, err := b.Session.ApplicationCommandBulkOverwrite(b.Session.State.User.ID, b.GuildID, commands)
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}
	for _, cmd := range registeredCommands {
		switch cmd.Name {
		case "setup":
			b.setupCommandID = cmd.ID
		case "me":
			b.meCommandID = cmd.ID
		case "online":
			b.onlineCommandID = cmd.ID
		case "offline":
			b.offlineCommandID = cmd.ID
		case "show":
			b.showPlayersCommandID = cmd.ID
		}
	}

	log.Println("Reading commands channel messages...")
	commandsChannelMessages, err := b.Session.ChannelMessages(b.commandsChannelID, 10, "", "", "")
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	commandMessageContent := "</setup:" + b.setupCommandID + "> : Set up your RDO profile for the server. Here you can set your R* ID for the Avatar, your camp location, bounty and a message that displays in the footer region in your online notification.\nTo find your R* ID, visit your Social Club profile here: <https://socialclub.rockstargames.com/games/rdr2/overview>.\nOn the tiny avatar of your character do a right-click and click on *Open image in new tab*. In the browser address bar you will notice a 9-digit number (just before */pedshot_0.jpg*). This is your R* ID which you can enter during setup to have your avatar displayed in online notifications.\n`/setup` is a convenient way to provide all info at once.\n\n</me:" + b.meCommandID + "> : This command displays your current profile information along with buttons for editing. It is a quick way to check and update your info.\n\n</online:" + b.onlineCommandID + "> : Flag yourself as online to let others know you are ingame.\nThe bot will respond with a message providing you with a couple of buttons for quickly editing your information during your gameplay.\nUse it in the channel of your platform (or lobby).\n\n</offline:" + b.offlineCommandID + "> : Flag yourself as offline to let others know you are not ingame anymore.\nUse it in the same channel where you flagged yourself as online.\n\n</show:" + b.showPlayersCommandID + "> : Show players that are online with their current data."

	if len(commandsChannelMessages) == 0 {
		log.Println("Adding command instructions...")
		_, err := b.Session.ChannelMessageSend(b.commandsChannelID, commandMessageContent)
		if err != nil {
			b.ErrorReport.Notify(err, nil)
			log.Println(err)
		}
	} else if commandsChannelMessages[0].Content != commandMessageContent {
		log.Println("Updating command instructions...")
		_, err := b.Session.ChannelMessageEdit(b.commandsChannelID, commandsChannelMessages[0].ID, commandMessageContent)
		if err != nil {
			b.ErrorReport.Notify(err, nil)
			log.Println(err)
		}
	}
}

func (b *Bot) updateChangelog() {
	var changelogFile *os.File
	var parsedChangelog bytes.Buffer

	log.Println("Reading changelog messages...")
	changelogMessages, err := b.Session.ChannelMessages(b.bulletinChannelID, 100, "", "", "")
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	if len(changelogMessages) > 0 {
		for _, m := range changelogMessages {
			err := b.Session.ChannelMessageDelete(b.bulletinChannelID, m.ID)
			if err != nil {
				b.ErrorReport.Notify(err, nil)
				log.Println(err)
			}
		}
	}

	log.Println("Getting current changelogs...")
	res, err := http.Get(b.ChangelogURL)
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}
	defer res.Body.Close()

	if _, err := os.Stat("updates.md"); err == nil {
		err := os.Remove("updates.md")
		if err != nil {
			b.ErrorReport.Notify(err, nil)
			log.Println(err)
		}
	}

	changelogFile, err = os.Create("updates.md")
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}
	defer changelogFile.Close()

	_, err = io.Copy(changelogFile, res.Body)
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	source, err := os.ReadFile("updates.md")
	if err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	if err := goldmark.New(goldmark.WithParserOptions(parser.WithBlockParsers())).Convert(source, &parsedChangelog); err != nil {
		b.ErrorReport.Notify(err, nil)
		log.Println(err)
	}

	changelogContent := strings.Trim(strings.ReplaceAll(parsedChangelog.String(), "<h1>Change Log</h1>\n", ""), " ")
	changelogs := strings.Split(changelogContent, "<p>.</p>")

	p := bluemonday.StripTagsPolicy()

	log.Println("Writing changelogs...")
	for _, c := range changelogs {
		htmlContent := strings.TrimSpace(c)
		extractTitle := strings.Split(htmlContent, "</h2>")
		extractChanges := strings.Split(extractTitle[1], "<h3>")
		extractAdded := strings.TrimSpace(extractChanges[1])
		extractChanged := strings.TrimSpace(extractChanges[2])
		extractFixed := strings.TrimSpace(extractChanges[3])

		sanitizedTitle := p.Sanitize(strings.TrimSpace(extractTitle[0]))
		sanitizedAdded := p.Sanitize(strings.ReplaceAll(extractAdded, "<li>", "- "))
		sanitizedChanged := p.Sanitize(strings.ReplaceAll(extractChanged, "<li>", "- "))
		sanitizedFixed := p.Sanitize(strings.ReplaceAll(extractFixed, "<li>", "- "))

		changelogString := "**" + sanitizedTitle + "**\n"
		changelogString += "```\n" + sanitizedAdded + "```"
		changelogString += "```\n" + sanitizedChanged + "```"
		changelogString += "```\n" + sanitizedFixed + "```\n"

		_, err = b.Session.ChannelMessageSend(b.bulletinChannelID, changelogString)
		if err != nil {
			b.ErrorReport.Notify(err, nil)
			log.Println(err)
		}
	}
}
