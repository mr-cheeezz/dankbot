package tabs

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

type Module struct {
	tabsStore     *postgres.UserTabStore
	settingsStore *postgres.TabsModuleSettingsStore
	allowedIDs    map[string]struct{}
}

func New(tabsStore *postgres.UserTabStore, settingsStore *postgres.TabsModuleSettingsStore, allowedIDs ...string) *Module {
	allowed := make(map[string]struct{}, len(allowedIDs))
	for _, id := range allowedIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = struct{}{}
		}
	}

	return &Module{
		tabsStore:     tabsStore,
		settingsStore: settingsStore,
		allowedIDs:    allowed,
	}
}

func (m *Module) Name() string {
	return "tabs"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"tab": {
			Handler:     m.tab,
			Description: "Shows a viewer's tab balance.",
			Usage:       "!tab <user>",
			Example:     "!tab mr_cheeezz",
		},
		"tab add": {
			Handler:     m.tabAdd,
			Description: "Adds an amount to a viewer's tab.",
			Usage:       "!tab add <user> <amount>",
			Example:     "!tab add mr_cheeezz 20",
		},
		"tab set": {
			Handler:     m.tabSet,
			Description: "Sets a viewer's tab balance.",
			Usage:       "!tab set <user> <amount>",
			Example:     "!tab set mr_cheeezz 15.50",
		},
		"tab paid": {
			Handler:     m.tabPaid,
			Description: "Clears a viewer's tab after payment.",
			Usage:       "!tab paid <user>",
			Example:     "!tab paid mr_cheeezz",
		},
		"tab give": {
			Handler:     m.tabGive,
			Description: "Opens a tab for a viewer if one does not exist yet.",
			Usage:       "!tab give <user>",
			Example:     "!tab give mr_cheeezz",
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	if m.settingsStore == nil {
		return nil
	}
	return m.settingsStore.EnsureDefault(ctx)
}

func (m *Module) tab(ctx modules.CommandContext) (string, error) {
	var login string
	if len(ctx.Args) < 1 {
		login = normalizeLogin(ctx.Sender)
	} else {
		login = normalizeLogin(ctx.Args[0])
	}
	if login == "" {
		return "usage: " + commandPrefix(ctx) + "tab <user>", nil
	}
	settings, err := m.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", nil
	}
	interestStartDelayDays := postgres.ResolveTabsInterestStartDelayDays(
		settings.InterestStartDelayMode,
		settings.InterestStartDelayValue,
		settings.InterestStartDelayUnit,
	)

	entry, interestApplied, err := m.tabsStore.GetWithInterest(
		context.Background(),
		login,
		settings.InterestRatePct,
		settings.InterestEveryDays,
		interestStartDelayDays,
	)
	if err != nil {
		return "", err
	}
	if entry == nil || entry.BalanceCents <= 0 {
		return fmt.Sprintf("%s has no open tab.", login), nil
	}

	if interestApplied > 0 {
		return fmt.Sprintf(
			"tab for %s: %s (interest applied: +%s).",
			displayName(entry),
			formatMoney(entry.BalanceCents),
			formatMoney(interestApplied),
		), nil
	}
	return fmt.Sprintf("tab for %s: %s.", displayName(entry), formatMoney(entry.BalanceCents)), nil
}

func (m *Module) tabAdd(ctx modules.CommandContext) (string, error) {
	if !m.canManage(ctx) {
		return "", nil
	}
	if len(ctx.Args) < 2 {
		return "usage: " + commandPrefix(ctx) + "tab add <user> <amount>", nil
	}
	settings, err := m.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", nil
	}
	interestStartDelayDays := postgres.ResolveTabsInterestStartDelayDays(
		settings.InterestStartDelayMode,
		settings.InterestStartDelayValue,
		settings.InterestStartDelayUnit,
	)

	login := normalizeLogin(ctx.Args[0])
	if login == "" {
		return "usage: " + commandPrefix(ctx) + "tab add <user> <amount>", nil
	}
	amountCents, err := parseMoneyToCents(ctx.Args[1])
	if err != nil {
		return "amount must be a valid number (for example: 10 or 10.50).", nil
	}
	if amountCents == 0 {
		return "amount must be non-zero.", nil
	}

	entry, interestApplied, err := m.tabsStore.Add(
		context.Background(),
		login,
		login,
		amountCents,
		settings.InterestRatePct,
		settings.InterestEveryDays,
		interestStartDelayDays,
	)
	if err != nil {
		return "", err
	}

	if interestApplied > 0 {
		return fmt.Sprintf(
			"added %s to %s's tab. new total: %s (interest +%s).",
			formatSignedMoney(amountCents),
			displayName(entry),
			formatMoney(entry.BalanceCents),
			formatMoney(interestApplied),
		), nil
	}
	return fmt.Sprintf(
		"added %s to %s's tab. new total: %s.",
		formatSignedMoney(amountCents),
		displayName(entry),
		formatMoney(entry.BalanceCents),
	), nil
}

func (m *Module) tabSet(ctx modules.CommandContext) (string, error) {
	if !m.canManage(ctx) {
		return "", nil
	}
	if len(ctx.Args) < 2 {
		return "usage: " + commandPrefix(ctx) + "tab set <user> <amount>", nil
	}
	settings, err := m.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", nil
	}
	interestStartDelayDays := postgres.ResolveTabsInterestStartDelayDays(
		settings.InterestStartDelayMode,
		settings.InterestStartDelayValue,
		settings.InterestStartDelayUnit,
	)

	login := normalizeLogin(ctx.Args[0])
	if login == "" {
		return "usage: " + commandPrefix(ctx) + "tab set <user> <amount>", nil
	}
	amountCents, err := parseMoneyToCents(ctx.Args[1])
	if err != nil {
		return "amount must be a valid number (for example: 10 or 10.50).", nil
	}
	if amountCents < 0 {
		amountCents = 0
	}

	entry, interestApplied, err := m.tabsStore.Set(
		context.Background(),
		login,
		login,
		amountCents,
		settings.InterestRatePct,
		settings.InterestEveryDays,
		interestStartDelayDays,
	)
	if err != nil {
		return "", err
	}

	if interestApplied > 0 {
		return fmt.Sprintf(
			"set %s's tab to %s (interest +%s applied before update).",
			displayName(entry),
			formatMoney(entry.BalanceCents),
			formatMoney(interestApplied),
		), nil
	}
	return fmt.Sprintf("set %s's tab to %s.", displayName(entry), formatMoney(entry.BalanceCents)), nil
}

func (m *Module) tabPaid(ctx modules.CommandContext) (string, error) {
	if !m.canManage(ctx) {
		return "", nil
	}
	if len(ctx.Args) < 1 {
		return "usage: " + commandPrefix(ctx) + "tab paid <user>", nil
	}
	settings, err := m.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", nil
	}
	interestStartDelayDays := postgres.ResolveTabsInterestStartDelayDays(
		settings.InterestStartDelayMode,
		settings.InterestStartDelayValue,
		settings.InterestStartDelayUnit,
	)

	login := normalizeLogin(ctx.Args[0])
	if login == "" {
		return "usage: " + commandPrefix(ctx) + "tab paid <user>", nil
	}

	entry, _, err := m.tabsStore.MarkPaid(
		context.Background(),
		login,
		login,
		settings.InterestRatePct,
		settings.InterestEveryDays,
		interestStartDelayDays,
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s's tab is now paid off.", displayName(entry)), nil
}

func (m *Module) tabGive(ctx modules.CommandContext) (string, error) {
	if !m.canManage(ctx) {
		return "", nil
	}
	if len(ctx.Args) < 1 {
		return "usage: " + commandPrefix(ctx) + "tab give <user>", nil
	}
	settings, err := m.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", nil
	}
	interestStartDelayDays := postgres.ResolveTabsInterestStartDelayDays(
		settings.InterestStartDelayMode,
		settings.InterestStartDelayValue,
		settings.InterestStartDelayUnit,
	)

	login := normalizeLogin(ctx.Args[0])
	if login == "" {
		return "usage: " + commandPrefix(ctx) + "tab give <user>", nil
	}

	entry, _, err := m.tabsStore.Ensure(
		context.Background(),
		login,
		login,
		settings.InterestRatePct,
		settings.InterestEveryDays,
		interestStartDelayDays,
	)
	if err != nil {
		return "", err
	}
	if entry == nil {
		return "could not open a tab for that user.", nil
	}

	return fmt.Sprintf("opened tab for %s. current balance: %s.", displayName(entry), formatMoney(entry.BalanceCents)), nil
}

func (m *Module) settings() (postgres.TabsModuleSettings, error) {
	if m.settingsStore == nil {
		return postgres.DefaultTabsModuleSettings(), nil
	}

	settings, err := m.settingsStore.Get(context.Background())
	if err != nil {
		return postgres.TabsModuleSettings{}, err
	}
	if settings == nil {
		defaults := postgres.DefaultTabsModuleSettings()
		return defaults, nil
	}
	return *settings, nil
}

func (m *Module) canManage(ctx modules.CommandContext) bool {
	if ctx.IsBroadcaster || ctx.IsModerator {
		return true
	}
	_, ok := m.allowedIDs[strings.TrimSpace(ctx.SenderID)]
	return ok
}

func normalizeLogin(value string) string {
	return strings.TrimSpace(strings.ToLower(strings.TrimPrefix(value, "@")))
}

func displayName(entry *postgres.UserTab) string {
	if entry == nil {
		return "user"
	}
	if name := strings.TrimSpace(entry.DisplayName); name != "" {
		return name
	}
	if login := strings.TrimSpace(entry.Login); login != "" {
		return login
	}
	return "user"
}

func parseMoneyToCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty amount")
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(parsed * 100)), nil
}

func formatMoney(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	dollars := cents / 100
	remainder := cents % 100
	return fmt.Sprintf("%s$%d.%02d", sign, dollars, remainder)
}

func formatSignedMoney(cents int64) string {
	if cents >= 0 {
		return "+" + formatMoney(cents)
	}
	return formatMoney(cents)
}

func commandPrefix(ctx modules.CommandContext) string {
	prefix := strings.TrimSpace(ctx.CommandPrefix)
	if prefix == "" {
		return "!"
	}
	return prefix
}
