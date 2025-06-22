package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSendWebhookMissingURL(t *testing.T) {
	os.Unsetenv("DISCORD_WEBHOOK_URL")
	err := SendWebhook(WebhookMessage{Username: "bot"})
	if err == nil {
		t.Fatalf("expected error for missing url")
	}
}

func TestSendWebhookHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadRequest)
	}))
	defer srv.Close()
	os.Setenv("DISCORD_WEBHOOK_URL", srv.URL)
	err := SendWebhook(WebhookMessage{Username: "bot"})
	if err == nil {
		t.Fatalf("expected error from discord server")
	}
}

func TestSendWebhookSuccess(t *testing.T) {
	var body WebhookMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	os.Setenv("DISCORD_WEBHOOK_URL", srv.URL)
	msg := WebhookMessage{Username: "bot"}
	if err := SendWebhook(msg); err != nil {
		t.Fatalf("SendWebhook: %v", err)
	}
	if body.Username != "bot" {
		t.Fatalf("expected username 'bot'")
	}
}
