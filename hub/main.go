package main

/* To Do List:
* Function to register a subscriber
* Function to verify a subscriber
* Function for notifying all subscribers
* Utility function for generating payloads
* Utility function for signing messages
 */

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Subscriber struct {
	CallbackURL string
	Secret      string
	Topic       string
}

type Hub interface {
	RegisterSubscriber(query string)
	NotifySubscribers(content []byte, topic string)
}

type BasicHub struct {
	subs []Subscriber
}

func (b *BasicHub) RegisterSubscriber(query string) {
	values, _ := url.ParseQuery(query)

	var sub Subscriber
	sub.CallbackURL = values.Get("hub.callback")
	sub.Topic = values.Get("hub.topic")
	sub.Secret = values.Get("hub.secret")

	log.Printf("Registering subscriber - CallbackURL: %s, Topic: %s, Secret: %s", sub.CallbackURL, sub.Topic, sub.Secret)

	b.subs = append(b.subs, sub)
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

	// Pass it as string to next function
	b.RegisterSubscriber(body)

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Subscriber registered")
}

func main() {
	server := &BasicHub{[]Subscriber{}}
	log.Fatal(http.ListenAndServe(":8080", server))
}
