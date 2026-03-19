package modules

import "context"

type Module interface {
	Name() string
	RegisterCommands() map[string]CommandDefinition
	Start(context.Context) error
}

type MessageHandler interface {
	HandleMessage(CommandContext) (MessageResult, error)
}

type MessageResult struct {
	Reply          string
	StopProcessing bool
}

type CommandHandler func(CommandContext) (string, error)

type CommandDefinition struct {
	Handler        CommandHandler
	Description    string
	Usage          string
	Example        string
	CanDisable     bool
	Configurable   bool
	DefaultEnabled bool
}

type CommandContext struct {
	Platform      string
	Channel       string
	SenderID      string
	Sender        string
	DisplayName   string
	IsModerator   bool
	IsBroadcaster bool
	CommandPrefix string
	FirstMessage  bool
	MessageID     string
	Message       string
	Command       string
	Args          []string
}
