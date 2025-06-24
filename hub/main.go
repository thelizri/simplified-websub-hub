package main

/* To Do List:
* Function to register a subscriber
* Function to verify a subscriber
* Function for notifying all subscribers
* Utility function for generating payloads
* Utility function for signing messages
 */

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const DEFAULT_TOPIC string = "/a/topic"

type Subscriber struct {
	CallbackURL string
	Secret      string
	Topic       string
}

type BasicHub struct {
	subs []Subscriber
}

func (b *BasicHub) GetSubscriber(query string) Subscriber {
	values, _ := url.ParseQuery(query)

	var sub Subscriber
	sub.CallbackURL = values.Get("hub.callback")
	sub.Topic = values.Get("hub.topic")
	sub.Secret = values.Get("hub.secret")

	log.Printf("Get subscriber - CallbackURL: %s, Topic: %s, Secret: %s", sub.CallbackURL, sub.Topic, sub.Secret)

	return sub
}

func (b *BasicHub) ValidateSubscriber(w http.ResponseWriter, r *http.Request, sub Subscriber) bool {
	// Generate a random challenge string
	challenge := make([]byte, 16)
	_, err := rand.Read(challenge)
	if err != nil {
		log.Printf("Failed to generate challenge: %v", err)
		return false
	}
	challengeStr := hex.EncodeToString(challenge)

	// Build the verification URL
	verifyURL, err := url.Parse(sub.CallbackURL)
	if err != nil {
		log.Printf("Invalid callback URL: %s", sub.CallbackURL)
		return false
	}

	// Append query parameters
	query := verifyURL.Query()
	query.Set("hub.mode", "subscribe")
	query.Set("hub.topic", sub.Topic)
	query.Set("hub.challenge", challengeStr)
	verifyURL.RawQuery = query.Encode()

	// Make the GET request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(verifyURL.String())
	if err != nil {
		log.Printf("Failed to contact subscriber at %s: %v", verifyURL.String(), err)
		return false
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading subscriber response: %v", err)
		return false
	}
	bodyStr := string(bodyBytes)

	// Compare response with challenge
	if resp.StatusCode == http.StatusOK && bodyStr == challengeStr {
		log.Printf("Successfully validated subscriber at %s", sub.CallbackURL)
		return true
	}

	log.Printf("Subscriber validation failed for %s. Expected challenge: %s, got: %s", sub.CallbackURL, challengeStr, bodyStr)
	return false
}

func (b *BasicHub) NotifySubscribers(content []byte, topic string) {
	for _, sub := range b.subs {
		if sub.Topic != topic {
			continue
		}

		// Sign the message
		signature := signMessage(sub.Secret, content)

		req, err := http.NewRequest("POST", sub.CallbackURL, bytes.NewBuffer(content))
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature", "sha256="+signature)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error sending notification to %s: %v", sub.CallbackURL, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			log.Printf("Warning: Subscriber %s responded with status %d", sub.CallbackURL, resp.StatusCode)
		} else {
			log.Printf("Notification sent to %s", sub.CallbackURL)
		}
	}
}

func (b *BasicHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
		return
	}

	// Read body from http.Request
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	body := string(bodyBytes)
	log.Printf("Request body: %s", body)
	sub := b.GetSubscriber(body)

	//Validate subscriber
	if ok := b.ValidateSubscriber(w, r, sub); ok {
		b.subs = append(b.subs, sub)
		// Respond with success
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Subscriber registered")
		return
	}

	// Respond with failure
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Subscriber not registered")
}

func main() {
	server := &BasicHub{[]Subscriber{}}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			testPayload := generateTestPayload()
			server.NotifySubscribers(testPayload, DEFAULT_TOPIC)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", server))
}
