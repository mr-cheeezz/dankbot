package eventsub

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func DesiredSubscriptions(streamerID string) []DesiredSubscription {
	return []DesiredSubscription{
		{Type: "stream.online", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}},
		{Type: "stream.offline", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}},
		{Type: "channel.ad_break.begin", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:ads"}},
		{Type: "channel.subscribe", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:subscriptions"}},
		{Type: "channel.subscription.gift", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:subscriptions"}},
		{Type: "channel.cheer", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"bits:read"}},
		{Type: "channel.channel_points_custom_reward_redemption.add", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:redemptions"}},
		{Type: "channel.poll.begin", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:polls"}},
		{Type: "channel.poll.progress", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:polls"}},
		{Type: "channel.poll.end", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:polls"}},
		{Type: "channel.prediction.begin", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:predictions"}},
		{Type: "channel.prediction.progress", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:predictions"}},
		{Type: "channel.prediction.lock", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:predictions"}},
		{Type: "channel.prediction.end", Version: "1", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:predictions"}},
		{Type: "channel.hype_train.begin", Version: "2", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:hype_train"}},
		{Type: "channel.hype_train.progress", Version: "2", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:hype_train"}},
		{Type: "channel.hype_train.end", Version: "2", Condition: map[string]string{"broadcaster_user_id": streamerID}, RequiredScopes: []string{"channel:read:hype_train"}},
	}
}

func SubscriptionKey(subscriptionType, version string, condition map[string]string) string {
	return subscriptionType + "|" + version + "|" + conditionKey(condition)
}

func SecretFingerprint(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func MissingScopes(actual, required []string) []string {
	if len(required) == 0 {
		return nil
	}

	actualSet := make(map[string]struct{}, len(actual))
	for _, scope := range actual {
		actualSet[scope] = struct{}{}
	}

	var missing []string
	for _, scope := range required {
		if _, ok := actualSet[scope]; !ok {
			missing = append(missing, scope)
		}
	}

	sort.Strings(missing)
	return missing
}

func conditionKey(condition map[string]string) string {
	if len(condition) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(condition))
	for key := range condition {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ordered := make(map[string]string, len(condition))
	for _, key := range keys {
		ordered[key] = condition[key]
	}

	body, _ := json.Marshal(ordered)
	return string(body)
}
