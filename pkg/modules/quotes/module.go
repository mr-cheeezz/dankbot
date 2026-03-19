package quotes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

type Module struct {
	store      *postgres.QuoteStore
	auditStore *postgres.AuditLogStore
	settings   *postgres.QuoteModuleSettingsStore
	allowedIDs map[string]struct{}
}

func New(store *postgres.QuoteStore, auditStore *postgres.AuditLogStore, allowedIDs ...string) *Module {
	allowed := make(map[string]struct{})
	for _, id := range allowedIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = struct{}{}
		}
	}

	return &Module{
		store:      store,
		auditStore: auditStore,
		allowedIDs: allowed,
	}
}

func (m *Module) Name() string {
	return "quotes"
}

func (m *Module) SetSettingsStore(store *postgres.QuoteModuleSettingsStore) {
	m.settings = store
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"quote": {
			Handler:     m.quote,
			Description: "Shows a random quote or a specific quote by number.",
			Usage:       "!quote [num]",
			Example:     "!quote 12",
		},
		"add quote": {
			Handler:     m.addQuote,
			Description: "Adds a new quote.",
			Usage:       "!add quote <message>",
			Example:     "!add quote this stream is cooked",
		},
		"create quote": {
			Handler:     m.addQuote,
			Description: "Adds a new quote.",
			Usage:       "!create quote <message>",
			Example:     "!create quote this stream is cooked",
		},
		"del quote": {
			Handler:     m.deleteQuote,
			Description: "Deletes a quote by number.",
			Usage:       "!del quote <num>",
			Example:     "!del quote 12",
		},
		"rm quote": {
			Handler:     m.deleteQuote,
			Description: "Deletes a quote by number.",
			Usage:       "!rm quote <num>",
			Example:     "!rm quote 12",
		},
		"edit quote": {
			Handler:     m.editQuote,
			Description: "Edits an existing quote.",
			Usage:       "!edit quote <num> <new message>",
			Example:     "!edit quote 12 this stream is saved",
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	if m.settings != nil {
		if err := m.settings.EnsureDefault(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) quote(ctx modules.CommandContext) (string, error) {
	enabled, err := m.moduleEnabled(context.Background())
	if err != nil || !enabled {
		return "", err
	}
	if m.store == nil {
		return "", fmt.Errorf("quote store is not configured")
	}

	if len(ctx.Args) == 0 {
		quote, err := m.store.Random(context.Background())
		if err != nil {
			return "", err
		}
		if quote == nil {
			return "there are no quotes yet.", nil
		}

		return formatQuote(*quote), nil
	}

	id, err := parseQuoteID(ctx.Args[0])
	if err != nil {
		return "usage: " + commandPrefix(ctx) + "quote [num]", nil
	}

	quote, err := m.store.Get(context.Background(), id)
	if err != nil {
		return "", err
	}
	if quote == nil {
		return fmt.Sprintf("quote #%d does not exist.", id), nil
	}

	return formatQuote(*quote), nil
}

func (m *Module) addQuote(ctx modules.CommandContext) (string, error) {
	enabled, err := m.moduleEnabled(context.Background())
	if err != nil || !enabled {
		return "", err
	}
	if !m.canManageQuotes(ctx) {
		return "", nil
	}
	if m.store == nil {
		return "", fmt.Errorf("quote store is not configured")
	}
	if len(ctx.Args) == 0 {
		return "usage: " + commandPrefix(ctx) + ctx.Command + " <message>", nil
	}

	quote, err := m.store.Create(context.Background(), strings.Join(ctx.Args, " "), quoteActor(ctx))
	if err != nil {
		return "", err
	}
	m.logAction(ctx, commandPrefix(ctx)+ctx.Command, fmt.Sprintf("added quote #%d", quote.ID))

	return fmt.Sprintf("added quote #%d.", quote.ID), nil
}

func (m *Module) deleteQuote(ctx modules.CommandContext) (string, error) {
	enabled, err := m.moduleEnabled(context.Background())
	if err != nil || !enabled {
		return "", err
	}
	if !m.canManageQuotes(ctx) {
		return "", nil
	}
	if m.store == nil {
		return "", fmt.Errorf("quote store is not configured")
	}
	if len(ctx.Args) == 0 {
		return "usage: " + commandPrefix(ctx) + ctx.Command + " <num>", nil
	}

	id, err := parseQuoteID(ctx.Args[0])
	if err != nil {
		return "usage: " + commandPrefix(ctx) + ctx.Command + " <num>", nil
	}

	deleted, err := m.store.Delete(context.Background(), id)
	if err != nil {
		return "", err
	}
	if !deleted {
		return fmt.Sprintf("quote #%d does not exist.", id), nil
	}
	m.logAction(ctx, commandPrefix(ctx)+ctx.Command, fmt.Sprintf("deleted quote #%d", id))

	return fmt.Sprintf("deleted quote #%d.", id), nil
}

func (m *Module) editQuote(ctx modules.CommandContext) (string, error) {
	enabled, err := m.moduleEnabled(context.Background())
	if err != nil || !enabled {
		return "", err
	}
	if !m.canManageQuotes(ctx) {
		return "", nil
	}
	if m.store == nil {
		return "", fmt.Errorf("quote store is not configured")
	}
	if len(ctx.Args) < 2 {
		return "usage: " + commandPrefix(ctx) + "edit quote <num> <new message>", nil
	}

	id, err := parseQuoteID(ctx.Args[0])
	if err != nil {
		return "usage: " + commandPrefix(ctx) + "edit quote <num> <new message>", nil
	}

	quote, err := m.store.Update(context.Background(), id, strings.Join(ctx.Args[1:], " "), quoteActor(ctx))
	if err != nil {
		return "", err
	}
	if quote == nil {
		return fmt.Sprintf("quote #%d does not exist.", id), nil
	}
	m.logAction(ctx, commandPrefix(ctx)+ctx.Command, fmt.Sprintf("edited quote #%d", quote.ID))

	return fmt.Sprintf("edited quote #%d.", quote.ID), nil
}

func (m *Module) canManageQuotes(ctx modules.CommandContext) bool {
	if ctx.IsBroadcaster || ctx.IsModerator {
		return true
	}

	_, ok := m.allowedIDs[strings.TrimSpace(ctx.SenderID)]
	return ok
}

func (m *Module) moduleEnabled(ctx context.Context) (bool, error) {
	if m.settings == nil {
		return true, nil
	}

	settings, err := m.settings.Get(ctx)
	if err != nil {
		return false, err
	}
	if settings == nil {
		return true, nil
	}

	return settings.Enabled, nil
}

func formatQuote(quote postgres.Quote) string {
	return fmt.Sprintf("quote #%d: %s", quote.ID, strings.TrimSpace(quote.Message))
}

func parseQuoteID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid quote id")
	}

	return id, nil
}

func commandPrefix(ctx modules.CommandContext) string {
	prefix := strings.TrimSpace(ctx.CommandPrefix)
	if prefix == "" {
		return "!"
	}

	return prefix
}

func quoteActor(ctx modules.CommandContext) string {
	if name := strings.TrimSpace(ctx.DisplayName); name != "" {
		return name
	}
	if name := strings.TrimSpace(ctx.Sender); name != "" {
		return name
	}

	return strings.TrimSpace(ctx.SenderID)
}

func (m *Module) logAction(ctx modules.CommandContext, command, detail string) {
	if m.auditStore == nil || strings.TrimSpace(command) == "" || strings.TrimSpace(detail) == "" {
		return
	}

	if _, err := m.auditStore.Create(context.Background(), postgres.AuditLog{
		Platform:  strings.TrimSpace(ctx.Platform),
		ActorID:   strings.TrimSpace(ctx.SenderID),
		ActorName: quoteActor(ctx),
		Command:   strings.TrimSpace(command),
		Detail:    strings.TrimSpace(detail),
	}); err != nil {
		fmt.Printf("audit log error: %v\n", err)
	}
}
