package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AlfioUserResolver struct {
	apiUrl string
}

// NewAlfioUserResolver returns an AlfioUserResolver for the provided API URL
func NewAlfioUserResolver(apiUrl string) AlfioUserResolver {
	return AlfioUserResolver{
		apiUrl: apiUrl,
	}
}

// AlfioTicket defines the parts of the alfio ticket that we care about
type AlfioTicket struct {
	FullName           string `json:"fullName"`
	TicketCategoryName string `json:"ticketCategoryName"`
}

// Resolve looks up a ticket to resolve into "${fullName} (${ticketCategory})"
func (a AlfioUserResolver) Resolve(event string, user string) (string, error) {
	url := fmt.Sprintf("%s/event/%s/ticket/%s", a.apiUrl, event, user)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf(res.Status)
	}

	var ticket AlfioTicket
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&ticket); err != nil {
		return "", err
	}

	username := fmt.Sprintf("%s (%s)", ticket.FullName, ticket.TicketCategoryName)
	return username, nil
}
