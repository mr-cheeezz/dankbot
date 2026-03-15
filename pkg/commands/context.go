package commands

type Context struct {
	Platform      string
	Channel       string
	SenderID      string
	Sender        string
	DisplayName   string
	IsModerator   bool
	IsBroadcaster bool
	Message       string
	Command       string
	Args          []string
}

type Handler func(Context) (Result, error)
