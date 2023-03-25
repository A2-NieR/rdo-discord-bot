package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type serverRole struct {
	ID    string
	Name  string
	Emoji string
}

var (
	guildRoles = make(map[string]*serverRole)
)

func (b *Bot) userWelcome(s *discordgo.Session, u *discordgo.GuildMemberAdd) {
	if len(u.Roles) == 0 {
		_, err := b.Session.ChannelMessageSend(b.generalChannelID, "Howdy <@"+u.User.ID+">, welcome to the server!\nTo get you started please select your roles in <#"+b.rolesChannelID+"> and have a look inside <#"+b.commandsChannelID+">.")
		if err != nil {
			b.ErrorReport.Notify(err, nil)
			log.Println(err)
		}
	}
}

func (b *Bot) assignRole(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if b.roleSelfAssignMessageID != "" && r.MessageID == b.roleSelfAssignMessageID && r.UserID != s.State.User.ID {
		for _, role := range guildRoles {
			if r.Emoji.Name == role.Emoji {
				err := b.Session.GuildMemberRoleAdd(b.GuildID, r.UserID, role.ID)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}
		}
	}
}

func (b *Bot) unassignRole(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if b.roleSelfAssignMessageID != "" && r.MessageID == b.roleSelfAssignMessageID && r.UserID != s.State.User.ID {
		for _, role := range guildRoles {
			if r.Emoji.Name == role.Emoji {
				err := b.Session.GuildMemberRoleRemove(b.GuildID, r.UserID, role.ID)
				if err != nil {
					b.ErrorReport.Notify(err, nil)
					log.Println(err)
				}
			}
		}
	}
}
