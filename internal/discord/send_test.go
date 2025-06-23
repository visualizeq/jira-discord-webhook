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

func TestSendWebhook_EmptyPayload(t *testing.T) {
	// Valid URL, but empty payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			t.Errorf("expected non-empty body")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	os.Setenv("DISCORD_WEBHOOK_URL", srv.URL)
	msg := WebhookMessage{}
	if err := SendWebhook(msg); err != nil {
		t.Fatalf("SendWebhook: %v", err)
	}
}

func TestSendWebhook_LargePayload(t *testing.T) {
	large := make([]byte, 2048)
	for i := range large {
		large[i] = 'A'
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	os.Setenv("DISCORD_WEBHOOK_URL", srv.URL)
	msg := WebhookMessage{
		Username: string(large),
	}
	if err := SendWebhook(msg); err != nil {
		t.Fatalf("SendWebhook: %v", err)
	}
}

func TestSendWebhook_UnreachableURL(t *testing.T) {
	os.Setenv("DISCORD_WEBHOOK_URL", "http://127.0.0.1:0") // Unreachable port
	msg := WebhookMessage{Username: "bot"}
	if err := SendWebhook(msg); err == nil {
		t.Fatalf("expected error for unreachable url")
	}
}
