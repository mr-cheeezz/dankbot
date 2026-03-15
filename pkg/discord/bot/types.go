package bot

type Config struct {
	BotToken  string
	ModRoleID string
}

type Message struct {
	ChannelID      string
	GuildID        string
	SenderID       string
	Sender         string
	DisplayName    string
	IsModerator    bool
	IsOwnerOrAdmin bool
	Content        string
}

type Handlers struct {
	OnMessage func(Message)
}
