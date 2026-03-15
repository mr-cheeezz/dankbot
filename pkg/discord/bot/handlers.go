package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (c *Client) registerHandlers() {
	c.session.AddHandler(func(s *discordgo.Session, message *discordgo.MessageCreate) {
		if message == nil || message.Author == nil {
			return
		}
		if message.Author.Bot {
			return
		}
		if c.handlers.OnMessage == nil {
			return
		}

		displayName := strings.TrimSpace(message.Author.GlobalName)
		if displayName == "" {
			displayName = strings.TrimSpace(message.Author.Username)
		}

		c.handlers.OnMessage(Message{
			ChannelID:      message.ChannelID,
			GuildID:        message.GuildID,
			SenderID:       message.Author.ID,
			Sender:         message.Author.Username,
			DisplayName:    displayName,
			IsModerator:    hasRole(message.Member, strings.TrimSpace(c.config.ModRoleID)),
			IsOwnerOrAdmin: memberHasAdmin(message.Member),
			Content:        message.Content,
		})
	})
}

func hasRole(member *discordgo.Member, roleID string) bool {
	if member == nil || roleID == "" {
		return false
	}

	for _, candidate := range member.Roles {
		if candidate == roleID {
			return true
		}
	}

	return false
}

func memberHasAdmin(member *discordgo.Member) bool {
	if member == nil {
		return false
	}

	return member.Permissions&discordgo.PermissionAdministrator != 0
}
