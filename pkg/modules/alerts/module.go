package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/eventsub"
)

const predictionProgressAlertThreshold = 50000

type Module struct {
	redis      *redispkg.Client
	stateStore *postgres.BotStateStore

	mu      sync.RWMutex
	channel string
	say     func(channel, message string) error
}

func New(redisClient *redispkg.Client, stateStore *postgres.BotStateStore) *Module {
	return &Module{
		redis:      redisClient,
		stateStore: stateStore,
	}
}

func (m *Module) Name() string {
	return "alerts"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	if m.redis == nil {
		return nil
	}

	subscription, err := m.redis.Subscribe(ctx, eventsub.AlertsChannel)
	if err != nil {
		return err
	}

	go m.run(ctx, subscription)
	return nil
}

func (m *Module) SetChatOutput(channel string, say func(channel, message string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.channel = strings.TrimSpace(channel)
	m.say = say
}

func (m *Module) run(ctx context.Context, subscription *redispkg.Subscription) {
	defer func() {
		if err := subscription.Close(); err != nil {
			fmt.Printf("alerts subscription close error: %v\n", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-subscription.Messages():
			if !ok {
				return
			}

			if err := m.handlePublishedEvent(ctx, message.Payload); err != nil {
				fmt.Printf("alerts event handling error: %v\n", err)
			}
		}
	}
}

func (m *Module) handlePublishedEvent(ctx context.Context, payload string) error {
	channel, say := m.output()
	if channel == "" || say == nil {
		return nil
	}

	if m.stateStore != nil {
		state, err := m.stateStore.Get(ctx)
		if err != nil {
			return err
		}
		if state != nil && state.KillswitchEnabled {
			return nil
		}
	}

	var event eventsub.PublishedEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return fmt.Errorf("decode published alert event: %w", err)
	}
	if event.Source != eventsub.SourceTwitchEventSub {
		return nil
	}

	message, err := m.renderTwitchAlert(event.Type, event.Event)
	if err != nil {
		return err
	}
	if strings.TrimSpace(message) == "" {
		return nil
	}

	return say(channel, message)
}

func (m *Module) renderTwitchAlert(eventType string, raw json.RawMessage) (string, error) {
	switch eventType {
	case "stream.online":
		var event eventsub.StreamStatusEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode stream online event: %w", err)
		}
		return fmt.Sprintf("%s just went live!", displayName(event.BroadcasterUserName, event.BroadcasterUserLogin)), nil
	case "stream.offline":
		var event eventsub.StreamStatusEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode stream offline event: %w", err)
		}
		return fmt.Sprintf("%s just went offline.", displayName(event.BroadcasterUserName, event.BroadcasterUserLogin)), nil
	case "channel.ad_break.begin":
		var event eventsub.AdBreakBeginEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode ad break event: %w", err)
		}
		return fmt.Sprintf("Heads up, an ad break just started for %d seconds.", event.DurationSeconds), nil
	case "channel.subscribe":
		var event eventsub.SubscriptionEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode subscription event: %w", err)
		}
		return fmt.Sprintf("Thank you %s for subscribing at %s!", displayName(event.UserName, event.UserLogin), humanTier(event.Tier)), nil
	case "channel.subscription.gift":
		var event eventsub.SubscriptionGiftEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode subscription gift event: %w", err)
		}
		if event.IsAnonymous {
			return fmt.Sprintf("An anonymous gifter just gifted %d subs at %s!", max(event.Total, 1), humanTier(event.Tier)), nil
		}
		return fmt.Sprintf("%s just gifted %d subs at %s!", displayName(event.UserName, event.UserLogin), max(event.Total, 1), humanTier(event.Tier)), nil
	case "channel.cheer":
		var event eventsub.CheerEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode cheer event: %w", err)
		}
		if event.IsAnonymous {
			return fmt.Sprintf("Someone just cheered %d bits!", event.Bits), nil
		}
		return fmt.Sprintf("Thank you %s for cheering %d bits!", displayName(event.UserName, event.UserLogin), event.Bits), nil
	case "channel.channel_points_custom_reward_redemption.add":
		var event eventsub.ChannelPointRedemptionEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode redemption event: %w", err)
		}
		return fmt.Sprintf("%s redeemed %s!", displayName(event.UserName, event.UserLogin), strings.TrimSpace(event.Reward.Title)), nil
	case "channel.poll.begin":
		var event eventsub.PollEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode poll begin event: %w", err)
		}
		return renderPollStart(event), nil
	case "channel.poll.progress":
		return "", nil
	case "channel.poll.end":
		var event eventsub.PollEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode poll end event: %w", err)
		}
		return renderPollEnd(event), nil
	case "channel.prediction.begin":
		var event eventsub.PredictionEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode prediction begin event: %w", err)
		}
		return renderPredictionStart(event), nil
	case "channel.prediction.progress":
		var event eventsub.PredictionEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode prediction progress event: %w", err)
		}
		return renderPredictionProgress(event), nil
	case "channel.prediction.lock":
		var event eventsub.PredictionEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode prediction lock event: %w", err)
		}
		return renderPredictionLocked(event), nil
	case "channel.prediction.end":
		var event eventsub.PredictionEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode prediction end event: %w", err)
		}
		if strings.EqualFold(strings.TrimSpace(event.Status), "canceled") {
			return fmt.Sprintf("The prediction was just canceled by %s.", displayName(event.BroadcasterUserName, event.BroadcasterUserLogin)), nil
		}
		return renderPredictionEnd(event), nil
	case "channel.hype_train.begin":
		var event eventsub.HypeTrainEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode hype train begin event: %w", err)
		}
		return fmt.Sprintf("The hype train just started at level %d!", levelOrOne(event.Level)), nil
	case "channel.hype_train.progress":
		var event eventsub.HypeTrainEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode hype train progress event: %w", err)
		}
		if event.Goal > 0 {
			return fmt.Sprintf("Hype train update: level %d with %d/%d progress!", levelOrOne(event.Level), event.Progress, event.Goal), nil
		}
		return fmt.Sprintf("Hype train update: level %d!", levelOrOne(event.Level)), nil
	case "channel.hype_train.end":
		var event eventsub.HypeTrainEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return "", fmt.Errorf("decode hype train end event: %w", err)
		}
		return fmt.Sprintf("The hype train ended at level %d.", levelOrOne(event.Level)), nil
	default:
		return "", nil
	}
}

func (m *Module) output() (string, func(channel, message string) error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.channel, m.say
}

func displayName(primary string, fallbacks ...string) string {
	if value := strings.TrimSpace(primary); value != "" {
		return value
	}
	for _, fallback := range fallbacks {
		if value := strings.TrimSpace(fallback); value != "" {
			return value
		}
	}

	return "someone"
}

func humanTier(tier string) string {
	switch strings.TrimSpace(strings.ToLower(tier)) {
	case "1000":
		return "tier 1"
	case "2000":
		return "tier 2"
	case "3000":
		return "tier 3"
	case "prime":
		return "prime"
	case "":
		return "a subscription tier"
	default:
		return tier
	}
}

func predictionOutcomeTitle(event eventsub.PredictionEvent) string {
	winningID := strings.TrimSpace(event.WinningOutcomeID)
	if winningID == "" {
		return ""
	}

	for _, outcome := range event.Outcomes {
		if strings.TrimSpace(outcome.ID) == winningID {
			return strings.TrimSpace(outcome.Title)
		}
	}

	return ""
}

func levelOrOne(level int) int {
	if level <= 0 {
		return 1
	}

	return level
}

func max(value, fallback int) int {
	if value <= 0 {
		return fallback
	}

	return value
}

func renderPollStart(event eventsub.PollEvent) string {
	pollMaker := displayName(event.BroadcasterUserName, event.BroadcasterUserLogin)
	title := strings.TrimSpace(event.Title)
	options := formatPollOptions(event.Choices)

	if event.ChannelPointsVoting.IsEnabled && event.ChannelPointsVoting.AmountPerVote > 0 {
		return fmt.Sprintf("%s has created a poll with extra votes for %d channel points. %s | %s", pollMaker, event.ChannelPointsVoting.AmountPerVote, title, options)
	}

	return fmt.Sprintf("%s has created a poll. %s | %s", pollMaker, title, options)
}

func renderPollEnd(event eventsub.PollEvent) string {
	winner := pollWinningChoice(event.Choices)
	breakdown := formatPollChannelPointBreakdown(event)
	if winner == "" {
		if breakdown != "" {
			return fmt.Sprintf("The poll has ended. Channel points spent: %s", breakdown)
		}
		return "The poll has ended."
	}

	if breakdown != "" {
		return fmt.Sprintf("The poll has ended. %s was the winner. Channel points spent: %s", winner, breakdown)
	}

	return fmt.Sprintf("The poll has ended. %s was the winner.", winner)
}

func renderPredictionStart(event eventsub.PredictionEvent) string {
	creator := displayName(event.BroadcasterUserName, event.BroadcasterUserLogin)
	title := strings.TrimSpace(event.Title)
	if title == "" {
		title = "a prediction"
	}

	voteWindow := formatPredictionVoteWindow(event)
	if len(event.Outcomes) == 2 {
		left := strings.TrimSpace(event.Outcomes[0].Title)
		right := strings.TrimSpace(event.Outcomes[1].Title)
		if left == "" {
			left = "option 1"
		}
		if right == "" {
			right = "option 2"
		}

		return fmt.Sprintf("%s has started a prediction! %s | %s or %s%s", creator, title, left, right, voteWindow)
	}

	return fmt.Sprintf("%s has started a prediction! %s Click on the prediction to see all options PogU%s", creator, title, voteWindow)
}

func formatPredictionVoteWindow(event eventsub.PredictionEvent) string {
	if event.LocksAt == nil || event.LocksAt.IsZero() {
		return ""
	}

	start := event.StartedAt
	if start.IsZero() {
		return ""
	}

	window := event.LocksAt.Sub(start)
	if window <= 0 {
		return ""
	}

	return fmt.Sprintf(" you have %s to vote.", humanDuration(window))
}

func renderPredictionLocked(event eventsub.PredictionEvent) string {
	leader, runnerUp := topPredictionOutcomes(event.Outcomes)
	if leader == nil {
		return "Prediction voting is now closed."
	}

	message := fmt.Sprintf("Prediction voting is now closed. %s has the most points (%d).", strings.TrimSpace(leader.Title), leader.ChannelPoints)
	if predictor := topPredictionUser(*leader); predictor != nil && predictor.ChannelPointsUsed > 0 {
		message += fmt.Sprintf(" %s put %d in.", displayName(predictor.UserName, predictor.UserLogin), predictor.ChannelPointsUsed)
	}

	if runnerUp != nil && strings.TrimSpace(runnerUp.Title) != "" {
		message += fmt.Sprintf(" %s has %d.", strings.TrimSpace(runnerUp.Title), runnerUp.ChannelPoints)
		if predictor := topPredictionUser(*runnerUp); predictor != nil && predictor.ChannelPointsUsed > 0 {
			message += fmt.Sprintf(" %s put %d in.", displayName(predictor.UserName, predictor.UserLogin), predictor.ChannelPointsUsed)
		}
	}

	return message
}

func renderPredictionEnd(event eventsub.PredictionEvent) string {
	winner := winningPredictionOutcome(event)
	title := strings.TrimSpace(event.Title)
	if title == "" {
		title = "this prediction"
	}
	if winner == nil {
		return fmt.Sprintf("The prediction ended. The outcome of %s was unavailable.", title)
	}

	message := fmt.Sprintf("The prediction ended the outcome of %s was %s.", title, strings.TrimSpace(winner.Title))
	if winner.ChannelPoints > 0 {
		recipients := topPredictionUsersSummary(*winner, 3)
		if recipients != "" {
			message += fmt.Sprintf(" %d go to %s.", winner.ChannelPoints, recipients)
		}
	}

	return message
}

func renderPredictionProgress(event eventsub.PredictionEvent) string {
	outcome, predictor := topPredictionPredictor(event.Outcomes)
	if outcome == nil || predictor == nil {
		return ""
	}

	if predictor.ChannelPointsUsed < predictionProgressAlertThreshold {
		return ""
	}

	return fmt.Sprintf("%s just put %d on %s.", displayName(predictor.UserName, predictor.UserLogin), predictor.ChannelPointsUsed, strings.TrimSpace(outcome.Title))
}

func topPredictionOutcomes(outcomes []eventsub.PredictionOutcome) (*eventsub.PredictionOutcome, *eventsub.PredictionOutcome) {
	if len(outcomes) == 0 {
		return nil, nil
	}

	var leader *eventsub.PredictionOutcome
	var runnerUp *eventsub.PredictionOutcome
	for i := range outcomes {
		outcome := &outcomes[i]
		if leader == nil || outcome.ChannelPoints > leader.ChannelPoints {
			runnerUp = leader
			leader = outcome
			continue
		}
		if runnerUp == nil || outcome.ChannelPoints > runnerUp.ChannelPoints {
			runnerUp = outcome
		}
	}

	return leader, runnerUp
}

func topPredictionUser(outcome eventsub.PredictionOutcome) *eventsub.PredictionPredictor {
	if len(outcome.TopPredictors) == 0 {
		return nil
	}

	best := &outcome.TopPredictors[0]
	for i := 1; i < len(outcome.TopPredictors); i++ {
		if outcome.TopPredictors[i].ChannelPointsUsed > best.ChannelPointsUsed {
			best = &outcome.TopPredictors[i]
		}
	}

	return best
}

func topPredictionPredictor(outcomes []eventsub.PredictionOutcome) (*eventsub.PredictionOutcome, *eventsub.PredictionPredictor) {
	var bestOutcome *eventsub.PredictionOutcome
	var bestPredictor *eventsub.PredictionPredictor

	for i := range outcomes {
		predictor := topPredictionUser(outcomes[i])
		if predictor == nil {
			continue
		}
		if bestPredictor == nil || predictor.ChannelPointsUsed > bestPredictor.ChannelPointsUsed {
			bestOutcome = &outcomes[i]
			bestPredictor = predictor
		}
	}

	return bestOutcome, bestPredictor
}

func winningPredictionOutcome(event eventsub.PredictionEvent) *eventsub.PredictionOutcome {
	winningID := strings.TrimSpace(event.WinningOutcomeID)
	if winningID == "" {
		return nil
	}

	for i := range event.Outcomes {
		if strings.TrimSpace(event.Outcomes[i].ID) == winningID {
			return &event.Outcomes[i]
		}
	}

	return nil
}

func topPredictionUsersSummary(outcome eventsub.PredictionOutcome, limit int) string {
	if len(outcome.TopPredictors) == 0 || limit <= 0 {
		return ""
	}

	names := make([]string, 0, len(outcome.TopPredictors))
	for _, predictor := range outcome.TopPredictors {
		name := displayName(predictor.UserName, predictor.UserLogin)
		if strings.TrimSpace(name) == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return ""
	}

	if len(names) > limit {
		return strings.Join(names[:limit], ", ") + ", and more.."
	}

	return strings.Join(names, ", ")
}

func humanDuration(value time.Duration) string {
	if value <= 0 {
		return "0 seconds"
	}

	totalSeconds := int(value.Round(time.Second) / time.Second)
	if totalSeconds < 60 {
		if totalSeconds == 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", totalSeconds)
	}

	if totalSeconds%60 == 0 {
		minutes := totalSeconds / 60
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}

	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	if minutes == 1 {
		return fmt.Sprintf("1 minute %d seconds", seconds)
	}
	return fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
}

func formatPollOptions(choices []eventsub.PollChoice) string {
	if len(choices) == 0 {
		return "no options"
	}

	options := make([]string, 0, len(choices))
	for _, choice := range choices {
		title := strings.TrimSpace(choice.Title)
		if title == "" {
			continue
		}
		options = append(options, title)
	}

	if len(options) == 0 {
		return "no options"
	}

	return strings.Join(options, " / ")
}

func pollWinningChoice(choices []eventsub.PollChoice) string {
	var (
		bestTitle string
		bestVotes = -1
	)

	for _, choice := range choices {
		title := strings.TrimSpace(choice.Title)
		if title == "" {
			continue
		}
		if choice.Votes > bestVotes {
			bestVotes = choice.Votes
			bestTitle = title
		}
	}

	return bestTitle
}

func formatPollChannelPointBreakdown(event eventsub.PollEvent) string {
	if !event.ChannelPointsVoting.IsEnabled || event.ChannelPointsVoting.AmountPerVote <= 0 {
		return ""
	}

	parts := make([]string, 0, len(event.Choices))
	for _, choice := range event.Choices {
		title := strings.TrimSpace(choice.Title)
		if title == "" || choice.ChannelPointsVotes <= 0 {
			continue
		}

		spent := choice.ChannelPointsVotes * event.ChannelPointsVoting.AmountPerVote
		parts = append(parts, fmt.Sprintf("%s - %d", title, spent))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " | ")
}
