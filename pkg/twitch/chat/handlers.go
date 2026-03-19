package chat

import twitch "github.com/gempir/go-twitch-irc/v4"

type Handlers struct {
	OnConnect        func()
	OnPrivateMessage func(Message)
	OnNotice         func(NoticeMessage)
	OnReconnect      func()
}

func (c *Client) RegisterHandlers(handlers Handlers) {
	c.handlers = handlers
}

func (c *Client) registerIRCHandlers() {
	c.irc.OnConnect(func() {
		if c.handlers.OnConnect == nil {
			return
		}

		c.handlers.OnConnect()
	})

	c.irc.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if c.handlers.OnPrivateMessage == nil {
			return
		}

		c.handlers.OnPrivateMessage(Message{
			Channel:        message.Channel,
			SenderID:       message.User.ID,
			Sender:         message.User.Name,
			DisplayName:    message.User.DisplayName,
			IsModerator:    hasBadge(message.User.Badges, "moderator"),
			IsBroadcaster:  hasBadge(message.User.Badges, "broadcaster"),
			FirstMessage:   message.FirstMessage,
			Text:           message.Message,
			ReplyTo:        message.ID,
			CustomRewardID: message.CustomRewardID,
		})
	})

	c.irc.OnNoticeMessage(func(message twitch.NoticeMessage) {
		if c.handlers.OnNotice == nil {
			return
		}

		c.handlers.OnNotice(NoticeMessage{
			Channel: message.Channel,
			Text:    message.Message,
		})
	})

	c.irc.OnReconnectMessage(func(message twitch.ReconnectMessage) {
		_ = message
		if c.handlers.OnReconnect == nil {
			return
		}

		c.handlers.OnReconnect()
	})
}

func hasBadge(badges map[string]int, name string) bool {
	if len(badges) == 0 {
		return false
	}

	_, ok := badges[name]
	return ok
}
