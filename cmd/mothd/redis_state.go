package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"github.com/go-redis/redis/v8"
	"github.com/dirtbags/moth/pkg/award"
)

const (
	REDIS_KEY_MESSAGE = "moth:message"
	REDIS_KEY_TEAMS = "moth:teams"
	REDIS_KEY_TEAM_IDS = "moth:team_ids"
	REDIS_KEY_POINT_LOG = "moth:points"	
)

type RedisState struct {
	ctx				context.Context

	redis_client	*redis.Client

	// Enabled tracks whether the current State system is processing updates
	Enabled bool

	eventStream     chan []string

	//lock      sync.RWMutex
}

type RedisEventEntry struct {
	/*
	[]string{
		strconv.FormatInt(time.Now().Unix(), 10),
		event,
		participantID,
		teamID,
		cat,
		strconv.Itoa(points),
	},
	extra...,*/
	// Unix epoch time of this event
	When     int64
	Event	string
	ParticipantID	string
	TeamID   string
	Category string
	Points   int
	Extra	[]string
}

// MarshalJSON returns the award event, encoded as a list.
func (e RedisEventEntry) MarshalJSON() ([]byte, error) {
	/*ao := []interface{}{
		a.When,
		a.TeamID,
		a.Category,
		a.Points,
	}*/

	return json.Marshal(e)
}

// UnmarshalJSON decodes the JSON string b.
func (e RedisEventEntry) UnmarshalJSON(b []byte) error {
	//r := bytes.NewReader(b)
	//dec := json.NewDecoder(r)
	//dec.UseNumber() // Don't use floats
	json.Unmarshal(b, &e)

	return nil
}

// NewRedisState returns a new State struct backed by the given Fs
func NewRedisState(redis_addr string, redis_db int) *RedisState {

	rdb := redis.NewClient(&redis.Options{
        Addr:     redis_addr,
        Password: "", // no password set
        DB:       redis_db,  // use default DB
    })

	s := &RedisState{
		Enabled:     true,
		eventStream: make(chan []string, 80),

		ctx: context.Background(),
		redis_client: rdb,
	}
	
	return s
}

// Messages retrieves the current messages.
func (s *RedisState) Messages() string {
	val, err := s.redis_client.Get(s.ctx, REDIS_KEY_MESSAGE).Result()

	if err != nil {
		return ""
	}

	return val
}

// TeamName returns team name given a team ID.
func (s *RedisState) TeamName(teamID string) (string, error) {
	team_name, err := s.redis_client.HGet(s.ctx, REDIS_KEY_TEAMS, teamID).Result()

	if err != nil {
		return "", fmt.Errorf("unregistered team ID: %s", teamID)
	}

	return team_name, nil
}

// SetTeamName writes out team name.
// This can only be done once per team.
func (s *RedisState) SetTeamName(teamID, teamName string) error {
	valid_id, err := s.redis_client.SIsMember(s.ctx, REDIS_KEY_TEAM_IDS, teamID).Result()

	if err != nil {
		return fmt.Errorf("Unexpected error while validating team ID: %s", teamID)
	} else if (!valid_id) {
		return fmt.Errorf("team ID: (%s) not found in list of valid team IDs", teamID)
	}

	success, err := s.redis_client.HSetNX(s.ctx, REDIS_KEY_TEAMS, teamID, teamName).Result()

	if err != nil {
		return fmt.Errorf("Unexpected error while setting team ID: %s and team Name: %s", teamID, teamName)
	}

	if (success) {
		return nil
	}

	return fmt.Errorf("Team ID: %s is already set", teamID)
}

// PointsLog retrieves the current points log.
func (s *RedisState) PointsLog() award.List {
	redis_args := &redis.ZRangeBy{
		Min: "0",
		Max: "-1",
	}
	scores, err := s.redis_client.ZRangeByScoreWithScores(s.ctx, REDIS_KEY_POINT_LOG, redis_args).Result()

	if err != nil {
		return make(award.List, 0)
	}
	
	point_log := make(award.List, len(scores))
	
	for _, item := range scores {
		point_entry := award.T{}

		point_string := strings.TrimSpace(item.Member.(string))

		point_entry.When = int64(item.Score)
		n, err := fmt.Sscanf(point_string, "%s %s %d", &point_entry.TeamID, &point_entry.Category, &point_entry.Points)

		if err != nil {
			// Do nothing
		} else if n != 3 {
			// Wrong number of fields, do nothing
		} else {
			point_log = append(point_log, point_entry)
		}
	}

	return point_log
}

func (s *RedisState) AwardPoints(teamID, category string, points int) error {
	redis_args := redis.ZAddArgs {
		LT: true,
	}
	awardTime := time.Now().Unix()

	point_string := fmt.Sprintf("%s %s %s", teamID, category, points)

	new_member := redis.Z{
		Score: float64(awardTime),
		Member: point_string,
	}

	redis_args.Members = append(redis_args.Members, new_member)

	_, err := s.redis_client.ZAddArgs(s.ctx, REDIS_KEY_POINT_LOG, redis_args).Result()

	if err != nil {
		return err
	}

	return nil
}

// LogEvent writes to the event log
func (s *RedisState) LogEvent(event, participantID, teamID, cat string, points int, extra ...string) {
	new_event := RedisEventEntry {
		When: time.Now().Unix(),
		Event: event,
		ParticipantID: participantID,
		TeamID: teamID,
		Category: cat,
		Points: points,
		Extra: extra,
	}

	message := new_event.MarshalJSON()
	/*
	s.eventStream <- append(
		[]string{
			strconv.FormatInt(time.Now().Unix(), 10),
			event,
			participantID,
			teamID,
			cat,
			strconv.Itoa(points),
		},
		extra...,
	)*/


}


func (s *RedisState) Maintain(updateInterval time.Duration) {
	/*
	ticker := time.NewTicker(updateInterval)
	s.refresh()
	for {
		select {
		case msg := <-s.eventStream:
			s.eventWriter.Write(msg)
			s.eventWriter.Flush()
			s.eventWriterFile.Sync()
		case <-ticker.C:
			s.refresh()
		case <-s.refreshNow:
			s.refresh()
		}
	}
	*/
}