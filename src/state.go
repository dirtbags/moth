package main

import (
	"errors"
)

type MOTHState interface {
	PointsLog(teamId string) []*Award
	AwardPoints(teamID string, category string, points int) error

	TeamName(teamId string) (string, error)
	isEnabled() bool
	getConfig(configName string) (string, error)
	getTeams() map[string]struct{}
	login(teamId string, token string) (bool, error)
	Initialize() (bool, error)
}

var ErrAlreadyRegistered = errors.New("This team ID has already been registered")
var ErrInvalidTeamID = errors.New("Invalid team ID")
var ErrRegistrationError = errors.New("Unable to register. Perhaps a teammate has already registered")
