package line

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type lineAPI struct {
	http               *http.Client
	baseURL            string
	channelAccessToken string
}

func newLineAPI(httpClient *http.Client, baseURL, channelAccessToken string) *lineAPI {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	baseURL = strings.TrimSpace(strings.TrimRight(baseURL, "/"))
	if baseURL == "" {
		baseURL = "https://api.line.me"
	}
	return &lineAPI{
		http:               httpClient,
		baseURL:            baseURL,
		channelAccessToken: strings.TrimSpace(channelAccessToken),
	}
}

type lineTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type lineBotInfoResponse struct {
	UserID string `json:"userId,omitempty"`
}

type lineReplyRequest struct {
	ReplyToken string            `json:"replyToken"`
	Messages   []lineTextMessage `json:"messages"`
}

type linePushRequest struct {
	To       string            `json:"to"`
	Messages []lineTextMessage `json:"messages"`
}

func (api *lineAPI) replyMessage(ctx context.Context, replyToken string, text string) error {
	if api == nil {
		return fmt.Errorf("line api is not initialized")
	}
	replyToken = strings.TrimSpace(replyToken)
	if replyToken == "" {
		return fmt.Errorf("line reply token is required")
	}
	return api.postJSON(ctx, "/v2/bot/message/reply", lineReplyRequest{
		ReplyToken: replyToken,
		Messages:   []lineTextMessage{{Type: "text", Text: strings.TrimSpace(text)}},
	})
}

func (api *lineAPI) pushMessage(ctx context.Context, chatID string, text string) error {
	if api == nil {
		return fmt.Errorf("line api is not initialized")
	}
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return fmt.Errorf("line chat id is required")
	}
	return api.postJSON(ctx, "/v2/bot/message/push", linePushRequest{
		To:       chatID,
		Messages: []lineTextMessage{{Type: "text", Text: strings.TrimSpace(text)}},
	})
}

func (api *lineAPI) botUserID(ctx context.Context) (string, error) {
	if api == nil {
		return "", fmt.Errorf("line api is not initialized")
	}
	url := api.baseURL + "/v2/bot/info"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+api.channelAccessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := api.http.Do(req)
	if err != nil {
		return "", err
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", parseLineAPIError(resp.StatusCode, raw)
	}
	var out lineBotInfoResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	userID := strings.TrimSpace(out.UserID)
	if userID == "" {
		return "", fmt.Errorf("line bot info returned empty user id")
	}
	return userID, nil
}

type lineAPIError struct {
	Status  int
	Message string
	Details []string
}

func (e *lineAPIError) Error() string {
	if e == nil {
		return "line api error"
	}
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		msg = "unknown_error"
	}
	if len(e.Details) == 0 {
		return fmt.Sprintf("line api http %d: %s", e.Status, msg)
	}
	return fmt.Sprintf("line api http %d: %s (%s)", e.Status, msg, strings.Join(e.Details, "; "))
}

type lineErrorResponse struct {
	Message string `json:"message,omitempty"`
	Details []struct {
		Message  string `json:"message,omitempty"`
		Property string `json:"property,omitempty"`
	} `json:"details,omitempty"`
}

func (api *lineAPI) postJSON(ctx context.Context, path string, payload any) error {
	if api == nil {
		return fmt.Errorf("line api is not initialized")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("line api path is required")
	}
	bodyRaw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := api.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyRaw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+api.channelAccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := api.http.Do(req)
	if err != nil {
		return err
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return parseLineAPIError(resp.StatusCode, raw)
}

func parseLineAPIError(status int, raw []byte) error {
	out := lineErrorResponse{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &lineAPIError{
			Status:  status,
			Message: strings.TrimSpace(string(raw)),
		}
	}
	details := make([]string, 0, len(out.Details))
	for _, detail := range out.Details {
		part := strings.TrimSpace(detail.Message)
		property := strings.TrimSpace(detail.Property)
		if property != "" {
			if part != "" {
				part = property + ": " + part
			} else {
				part = property
			}
		}
		if part == "" {
			continue
		}
		details = append(details, part)
	}
	return &lineAPIError{
		Status:  status,
		Message: strings.TrimSpace(out.Message),
		Details: details,
	}
}

func shouldFallbackToLinePush(err error) bool {
	var apiErr *lineAPIError
	if !errors.As(err, &apiErr) || apiErr == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(apiErr.Message))
	if msg == "" {
		msg = strings.ToLower(strings.TrimSpace(strings.Join(apiErr.Details, " ")))
	}
	msg = strings.ReplaceAll(msg, "_", " ")
	return strings.Contains(msg, "reply token") || strings.Contains(msg, "replytoken")
}

func sendLineText(ctx context.Context, api *lineAPI, logger *slog.Logger, chatID string, text string, replyToken string) error {
	if api == nil {
		return fmt.Errorf("line api is not initialized")
	}
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return fmt.Errorf("line chat id is required")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("line text is required")
	}
	replyToken = strings.TrimSpace(replyToken)
	if replyToken != "" {
		if err := api.replyMessage(ctx, replyToken, text); err == nil {
			return nil
		} else if shouldFallbackToLinePush(err) {
			if logger != nil {
				logger.Warn("line_reply_failed_fallback_push",
					"chat_id", chatID,
					"error_class", "reply_token_invalid_or_expired",
					"error", err.Error(),
				)
			}
		} else {
			return err
		}
	}
	return api.pushMessage(ctx, chatID, text)
}
