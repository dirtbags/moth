package main

import (
	"errors"
)

type MOTHState interface {
	// Perform any setup needed
	Initialize() (bool, error)

	// Return a list of awarded points
	PointsLog(teamId string) []*Award

	// Award points to a team
	AwardPoints(teamID string, category string, points int) error

	// Given a team hash/token/id, retrieve the team name, if possible
	TeamName(teamId string) (string, error)

	// Returns true if the event is currently enabled
	isEnabled() bool

	// Attempt to read an arbitrary configuration item, if possible
	getConfig(configName string) (string, error)

	// Return a list of valid team IDs
	getValidTeamIds() map[string]struct{}

	// Attempt to register/log in.
	login(teamName string, token string) (bool, error)
}

var ErrAlreadyRegistered = errors.New("This team ID has already been registered")
var ErrInvalidTeamID = errors.New("Invalid team ID")
var ErrRegistrationError = errors.New("Unable to register. Perhaps a teammate has already registered")
