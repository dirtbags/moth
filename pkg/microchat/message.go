package main

// Message contains everything sent to the client about a single message
type Message struct {
	// User is the full ID of the user sending this message
	User string

	// Text is the message itself
	Text string
}
