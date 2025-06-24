package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

func logHttpMessage(w http.ResponseWriter, req *http.Request) {
	// Dump full request, including body
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		http.Error(w, "Failed to dump request", http.StatusInternalServerError)
		return
	}

	// Write to file
	f, err := os.OpenFile("requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Failed to open log file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	f.WriteString("----- New Request -----\n")
	f.Write(dump)
	f.WriteString("\n\n")

	// Respond for debug visibility
	fmt.Fprintf(w, "Request logged.\n")
}

// Sign message using HMAC SHA1
func signMessage(secret string, message []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

// Generate some default JSON
func generateTestPayload() []byte {
	payload := map[string]string{
		"event":   "test_event",
		"message": "Hello, subscriber!",
		"status":  "success",
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to generate test payload: %v", err)
		return []byte(`{}`)
	}

	return jsonBytes
}
