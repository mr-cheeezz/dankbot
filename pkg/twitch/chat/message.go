package chat

type Message struct {
	Channel        string
	SenderID       string
	Sender         string
	DisplayName    string
	IsModerator    bool
	IsBroadcaster  bool
	IsVIP          bool
	IsSubscriber   bool
	FirstMessage   bool
	Text           string
	ReplyTo        string
	CustomRewardID string
}

type NoticeMessage struct {
	Channel string
	Text    string
}
