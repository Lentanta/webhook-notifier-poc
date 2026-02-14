package sender

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSendWebhook_Success tests successful webhook delivery
func TestSendWebhook_Success(t *testing.T) {
	// Create a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test data
	testData := []byte(`{"test": "data"}`)

	// Call sendWebhook with test server URL
	err, shouldRetry := sendWebhook(server.URL, testData)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if shouldRetry {
		t.Errorf("Expected shouldRetry=false, got true")
	}
}
