package modules

import (
	"context"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/commands"
)

type Runner struct {
	dispatcher *commands.Dispatcher
	modules    []Module
}

func NewRunner(dispatcher *commands.Dispatcher) *Runner {
	return &Runner{dispatcher: dispatcher}
}

func (r *Runner) Register(module Module) {
	if module == nil {
		return
	}

	if r.dispatcher != nil {
		for name, definition := range module.RegisterCommands() {
			moduleDefinition := definition
			moduleHandler := definition.Handler
			r.dispatcher.RegisterWithDefinition(name, func(ctx commands.Context) (commands.Result, error) {
				reply, err := moduleHandler(CommandContext{
					Platform:      ctx.Platform,
					Channel:       ctx.Channel,
					SenderID:      ctx.SenderID,
					Sender:        ctx.Sender,
					DisplayName:   ctx.DisplayName,
					IsModerator:   ctx.IsModerator,
					IsBroadcaster: ctx.IsBroadcaster,
					CommandPrefix: ctx.CommandPrefix,
					FirstMessage:  false,
					MessageID:     "",
					Message:       ctx.Message,
					Command:       ctx.Command,
					Args:          append([]string(nil), ctx.Args...),
				})
				if err != nil {
					return commands.Result{}, err
				}
				return commands.Result{
					Handled: reply != "",
					Reply:   reply,
				}, nil
			}, commands.Definition{
				Name:           name,
				Module:         module.Name(),
				Description:    moduleDefinition.Description,
				Usage:          moduleDefinition.Usage,
				Example:        moduleDefinition.Example,
				CanDisable:     moduleDefinition.CanDisable,
				Configurable:   moduleDefinition.Configurable,
				DefaultEnabled: moduleDefinition.DefaultEnabled,
			})
		}
	}

	r.modules = append(r.modules, module)
}

func (r *Runner) Start(ctx context.Context) error {
	for _, module := range r.modules {
		if err := module.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) HandleMessage(ctx CommandContext) (MessageResult, error) {
	for _, module := range r.modules {
		handler, ok := module.(MessageHandler)
		if !ok {
			continue
		}

		result, err := handler.HandleMessage(ctx)
		if err != nil {
			return MessageResult{}, err
		}
		if result.StopProcessing || strings.TrimSpace(result.Reply) != "" {
			return result, nil
		}
	}

	return MessageResult{}, nil
}
