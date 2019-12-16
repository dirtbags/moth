package main

import (
	"log"
	"errors"
	"database/sql"
	"path"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
)

var SQLite3ErrAwardingPointsUnique = errors.New("Your team has already solved this puzzle")

type SQLiteMOTHState struct {
	database	*sql.DB
	StateDir	string
}

func (state *SQLiteMOTHState) Initialize() (bool, error) {
	database, err := sql.Open("sqlite3", state.StatePath("moth.sqlite3"))
	if err != nil {
		return false, err
	}
	state.database = database

	if _, err := database.Exec("CREATE TABLE IF NOT EXISTS config (id INTEGER PRIMARY KEY, key TEXT UNIQUE, value TEXT)"); err != nil {
		return false, err
	}

	if _, err := database.Exec("CREATE TABLE IF NOT EXISTS teams (id INTEGER PRIMARY KEY, team_name TEXT, team_hash TEXT, UNIQUE(team_name, team_hash))"); err != nil {
		return false, err
	}

	if _, err := database.Exec("CREATE TABLE IF NOT EXISTS valid_team_ids (id INTEGER PRIMARY KEY, team_hash TEXT UNIQUE)"); err != nil {
		return false, err
	}


	if _, err := database.Exec("CREATE TABLE IF NOT EXISTS points (id INTEGER PRIMARY KEY, time INTEGER, team_id TEXT, category TEXT, points INTEGER, UNIQUE(team_id, category, points))"); err != nil {
		return false, err
	}

	// Only do these things if we haven't been initialized
	if _, err := state.getConfig("initialized"); err != nil {
		log.Printf("Initialized config missing, re-initializing")

		if len(state.getValidTeamIds()) == 0 {
			for i := 0; i <= 100; i += 1 {
				state.database.Exec("REPLACE INTO valid_team_ids (team_hash) VALUES (?)", mktoken())
			}
		}

		state.setConfig("initialized", "true")
	}

	return true, nil
}

func (state *SQLiteMOTHState) StatePath(parts ...string) string {
        tail := pathCleanse(parts)
        return path.Join(state.StateDir, tail)
}

func (state *SQLiteMOTHState) PointsLog(teamId string) []*Award {
	var ret []*Award

	var When int64
	var TeamId string
	var Category string
	var Points int

	var rows *sql.Rows

	if len(teamId) > 0 {
		rows, _ = state.database.Query("SELECT time, team_id, category, points FROM points WHERE team_id = ?", teamId)

	} else {
		rows, _ = state.database.Query("SELECT time, team_id, category, points FROM points")
	}

	for rows.Next() {
		rows.Scan(&When, &TeamId, &Category, &Points)
		new_award := Award{}
		new_award.When = time.Unix(When, 0)
		new_award.TeamId = TeamId
		new_award.Category = Category
		new_award.Points = Points
		ret = append(ret, &new_award)
	}

	return ret
}

func (state *SQLiteMOTHState) AwardPoints(teamID string, category string, points int) error {
	_, err := state.database.Exec("INSERT INTO points (time, team_id, category, points) VALUES (?, ?, ?, ?)", time.Now().Unix(), teamID, category, points)
	if sqliteErr, ok := err.(sqlite3.Error); ok {
		if sqliteErr.Code == sqlite3.ErrConstraint {
			return SQLite3ErrAwardingPointsUnique
		}
	}

	return err
}

func (state *SQLiteMOTHState) TeamName(teamId string) (string, error) {
	res := state.database.QueryRow("SELECT team_name FROM teams WHERE team_hash = ?", teamId)

	var teamName string

	err := res.Scan(&teamName)

	return teamName, err
}

func (state *SQLiteMOTHState) isEnabled() bool {
	if res, _ := state.getConfig("disabled"); res == "true" {
		return false
	}

	if res, err := state.getConfig("until"); err == nil {
                untilspecs := strings.TrimSpace(res)
                until, err := time.Parse(time.RFC3339, untilspecs)
                if err != nil {
                        log.Printf("Suspended: Unparseable until date: %s", untilspecs)
                        return false
                }
                if until.Before(time.Now()) {
                        log.Print("Suspended: until time reached, suspending maintenance")
                        return false
                }
        }


	return true
}

func (state *SQLiteMOTHState) getConfig(configName string) (string, error) {
	res := state.database.QueryRow("SELECT value FROM config WHERE key = ?", configName)

	var value string

	err := res.Scan(&value)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (state *SQLiteMOTHState) setConfig(configName string, value string) error {
	_, err := state.database.Exec("REPLACE INTO config (key, value) VALUES (?, ?)", configName, value)
	return err
}

func (state *SQLiteMOTHState) getValidTeamIds() map[string]struct{} {
	teams := make(map[string]struct{})
	var team_id string

	rows, _ := state.database.Query("SELECT team_hash FROM valid_team_ids")
	for rows.Next() {
		rows.Scan(&team_id)
		teams[team_id] = struct{}{}
	}

	return teams
}

func (state *SQLiteMOTHState) login(teamName string, token string) (bool, error) {
	if _, err := state.TeamName(token); err == nil{
		return true, ErrAlreadyRegistered
	}

	row := state.database.QueryRow("SELECT count(*) FROM valid_team_ids WHERE team_hash = ?", token)

        var count int

	err := row.Scan(&count)
	if err != nil {
		return false, ErrRegistrationError
	}

	if count == 0 {
		return false, ErrInvalidTeamID
	}

	if _, err := state.database.Exec("INSERT INTO teams (team_name, team_hash) VALUES (?, ?)", teamName, token); err != nil {
		return false, ErrRegistrationError
	}

	return true, nil
}
