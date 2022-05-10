package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

// HmacResolverSeparator is the string used to separater username from hmac
const HmacResolverSeparator = "::"

// HmacResolver resolves usernames using SHA256 HMAC
type HmacResolver struct {
	key string
}

// Resolve resolves usernames using HMAC.
//
// User strings are expected to be the concatenation of:
//   desired username, HmacResolverSeparator, MAC
//
// If there is no separator, the correct user string is computed and printed to the log.
// So you can use this to compute the correct usernames.
func (h *HmacResolver) Resolve(event string, user string) (string, error) {
	userparts := strings.Split(user, HmacResolverSeparator)
	username := userparts[0]

	mac := hmac.New(sha256.New, []byte(h.key))
	fmt.Fprint(mac, event)
	fmt.Fprint(mac, user)
	expectedMAC := mac.Sum(nil)

	if len(userparts) == 1 {
		expectedEnc := base64.URLEncoding.EncodeToString(expectedMAC)
		log.Printf("Authenticated username: %s%s%s", username, HmacResolverSeparator, expectedEnc)
		return "", fmt.Errorf("No authentication provided")
	}
	givenMAC, err := base64.URLEncoding.DecodeString(userparts[1])
	if err != nil {
		return "", err
	}

	if hmac.Equal(givenMAC, expectedMAC) {
		return username, nil
	}

	return "", fmt.Errorf("Authentication failed")
}
