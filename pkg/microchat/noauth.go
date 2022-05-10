package main

import "fmt"

// NoAuthResolver is a pass-through resolver
type NoAuthResolver struct {
}

// Resolve just returns user, no authentication whatsover is performed
func (n NoAuthResolver) Resolve(event string, user string) (string, error) {
	if (event == "") || (user == "") {
		return user, fmt.Errorf("User and event must be specified")
	}
	if (len(event) > 40) || (len(user) > 40) {
		return "", fmt.Errorf("Too large for me to handle!")
	}
	return user, nil
}
