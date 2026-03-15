package chat

func (c *Client) Join(channel string) error {
	if channel == "" {
		return nil
	}

	current := c.sync.OwnedChannels()
	for _, owned := range current {
		if owned == channel {
			return nil
		}
	}

	c.sync.SetOwnedChannels(append(current, channel))
	c.irc.Join(channel)
	return nil
}

func (c *Client) Depart(channel string) error {
	if channel == "" {
		return nil
	}

	current := c.sync.OwnedChannels()
	next := make([]string, 0, len(current))
	for _, owned := range current {
		if owned != channel {
			next = append(next, owned)
		}
	}

	c.sync.SetOwnedChannels(next)
	c.irc.Depart(channel)
	return nil
}

func (c *Client) Say(channel, message string) error {
	if channel == "" || message == "" {
		return nil
	}

	c.irc.Say(channel, message)
	return nil
}

func (c *Client) Reply(channel, replyTo, message string) error {
	if channel == "" || replyTo == "" || message == "" {
		return nil
	}

	c.irc.Reply(channel, replyTo, message)
	return nil
}
