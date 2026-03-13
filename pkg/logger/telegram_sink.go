package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// TelegramSink posts selected log entries to a Telegram chat.
type TelegramSink struct {
	botToken string
	chatID   string
	minLevel Level
	client   *http.Client
}

// NewTelegramSink constructs a Telegram sink.
func NewTelegramSink(botToken, chatID string, minLevel Level) (*TelegramSink, error) {
	if botToken == "" || chatID == "" {
		return nil, errors.New("botToken and chatID required")
	}
	return &TelegramSink{
		botToken: botToken,
		chatID:   chatID,
		minLevel: minLevel,
		client:   &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// NewTelegramSinkFromEnv creates a Telegram sink from environment variables.
func NewTelegramSinkFromEnv(minLevel Level) (*TelegramSink, error) {
	return NewTelegramSink(os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID"), minLevel)
}

// Log forwards the entry to Telegram when it passes the configured level threshold.
func (t *TelegramSink) Log(e Entry) error {
	if levelFromString(e.Level) < t.minLevel {
		return nil
	}

	fields, err := json.Marshal(e.Fields)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	body.WriteString("[")
	body.WriteString(e.Level)
	body.WriteString("] ")
	body.WriteString(e.Msg)
	body.WriteString("\ntime: ")
	body.WriteString(e.Time.Format(time.RFC3339))
	if len(fields) > 0 {
		body.WriteString("\nfields: ")
		body.Write(fields)
	}

	apiURL := "https://api.telegram.org/bot" + url.PathEscape(t.botToken) + "/sendMessage"
	form := url.Values{}
	form.Set("chat_id", t.chatID)
	form.Set("text", body.String())

	resp, err := t.client.PostForm(apiURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("telegram sink: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func levelFromString(s string) Level {
	switch s {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}
