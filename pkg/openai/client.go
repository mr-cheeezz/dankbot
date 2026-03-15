package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const responsesURL = "https://api.openai.com/v1/responses"

type Client struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

type KeywordDecision struct {
	ShouldTrigger bool    `json:"should_trigger"`
	Confidence    float64 `json:"confidence"`
	Reason        string  `json:"reason"`
}

type responseRequest struct {
	Model        string          `json:"model"`
	Instructions string          `json:"instructions,omitempty"`
	Input        string          `json:"input"`
	Store        bool            `json:"store"`
	Text         responseTextCfg `json:"text"`
}

type responseTextCfg struct {
	Format responseFormat `json:"format"`
}

type responseFormat struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type responseEnvelope struct {
	OutputText string `json:"output_text"`
}

type errorEnvelope struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func NewClient(apiKey, model string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
	}
}

func (c *Client) ShouldTriggerKeyword(ctx context.Context, trigger, response, message string) (bool, float64, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return false, 0, fmt.Errorf("openai api key is not configured")
	}
	if strings.TrimSpace(c.model) == "" {
		return false, 0, fmt.Errorf("openai model is not configured")
	}

	requestBody := responseRequest{
		Model: c.model,
		Instructions: strings.TrimSpace(`You decide whether a Twitch chat message is genuinely trying to trigger a configured bot keyword.

Be conservative. Return should_trigger=false if the keyword is only mentioned casually, quoted, or appears in unrelated text.
Return should_trigger=true only if the chatter is clearly trying to invoke the keyword response or directly talking to the bot about that keyword.
Confidence must be between 0 and 1.`),
		Input: buildKeywordPrompt(trigger, response, message),
		Store: false,
		Text: responseTextCfg{
			Format: responseFormat{
				Type:   "json_schema",
				Name:   "keyword_trigger_decision",
				Strict: true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"should_trigger": map[string]any{
							"type": "boolean",
						},
						"confidence": map[string]any{
							"type":    "number",
							"minimum": 0,
							"maximum": 1,
						},
						"reason": map[string]any{
							"type":      "string",
							"minLength": 1,
						},
					},
					"required":             []string{"should_trigger", "confidence", "reason"},
					"additionalProperties": false,
				},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return false, 0, fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, responsesURL, bytes.NewReader(body))
	if err != nil {
		return false, 0, fmt.Errorf("create openai request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, 0, fmt.Errorf("send openai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		var apiErr errorEnvelope
		if err := json.Unmarshal(payload, &apiErr); err == nil && strings.TrimSpace(apiErr.Error.Message) != "" {
			return false, 0, fmt.Errorf("openai responses api error (%s): %s", apiErr.Error.Type, apiErr.Error.Message)
		}
		return false, 0, fmt.Errorf("openai responses api error: status %d", resp.StatusCode)
	}

	var envelope responseEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return false, 0, fmt.Errorf("decode openai response: %w", err)
	}

	var decision KeywordDecision
	if err := json.Unmarshal([]byte(envelope.OutputText), &decision); err != nil {
		return false, 0, fmt.Errorf("decode openai keyword decision: %w", err)
	}

	return decision.ShouldTrigger, decision.Confidence, nil
}

func buildKeywordPrompt(trigger, response, message string) string {
	return fmt.Sprintf(
		"keyword trigger: %s\nconfigured response: %s\nchat message: %s\n\nDecide whether the chatter is genuinely trying to trigger this keyword response.",
		strings.TrimSpace(trigger),
		strings.TrimSpace(response),
		strings.TrimSpace(message),
	)
}
