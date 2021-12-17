package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/go-redis/redis/v8"
	"github.com/dirtbags/moth/pkg/award"
)

const (
	REDIS_KEY_PREFIX = "moth"
	REDIS_KEY_MESSAGE = "message"
	REDIS_KEY_TEAMS = "teams"
	REDIS_KEY_TEAM_IDS = "team_ids"
	REDIS_KEY_POINT_LOG = "points"
	REDIS_KEY_EVENT_LOG = "events"
	REDIS_KEY_ENABLED = "enabled"
)

type RedisState struct {
	ctx				context.Context

	redis_client	*redis.Client

	// Enabled tracks whether the current State system is processing updates
	Enabled bool

	instance_id		string

	eventStream     chan map[string]interface{}
}

type RedisAward struct {
	when	int64
	teamID	string
	category	string
	points	int
}

// NewRedisState returns a new State struct backed by the given Fs
func NewRedisState(redis_addr string, redis_db int, instance_id string) *RedisState {

	rdb := redis.NewClient(&redis.Options{
        Addr:     redis_addr,
        Password: "", // no password set
        DB:       redis_db,  // use default DB
    })

	s := &RedisState{
		ctx: context.Background(),
		redis_client: rdb,
		Enabled:     true,
		instance_id:	instance_id,
		eventStream: make(chan map[string]interface{}, 80),
	}
	
	//s.initialize()

	return s
}

func (s *RedisState) formatRedisKey(key string) string {
	return fmt.Sprintf("%s:%s:%s", REDIS_KEY_PREFIX, s.instance_id, key)
}

func (s *RedisState) initialize() {
	s.SetMessagesOverride("", false)
}


// ************ Message-related operations ****************

// Messages retrieves the current messages.
func (s *RedisState) Messages() string {
	val, err := s.redis_client.Get(s.ctx, s.formatRedisKey(REDIS_KEY_MESSAGE)).Result()

	if err != nil {
		return ""
	}

	return val
}

func (s *RedisState) SetMessages(message string) error {
	return s.SetMessagesOverride(message, true)
}

func (s *RedisState) SetMessagesOverride(message string, override bool) error {
	if override {
		return s.redis_client.Set(s.ctx, s.formatRedisKey(REDIS_KEY_MESSAGE), message, 0).Err()
	} else {
		return s.redis_client.SetNX(s.ctx, s.formatRedisKey(REDIS_KEY_MESSAGE), message, 0).Err()
	}
}

// ******************** Team operations ******************

func (s *RedisState) TeamIDs() ([]string, error) {
	return s.redis_client.SMembers(s.ctx, s.formatRedisKey(REDIS_KEY_TEAM_IDS)).Result()
}

func (s *RedisState) AddTeamID(teamID string) error {
	return s.redis_client.SAdd(s.ctx, s.formatRedisKey(REDIS_KEY_TEAM_IDS), teamID).Err()
}

func (s *RedisState) TeamNames() (map[string]string, error) {
	return s.redis_client.HGetAll(s.ctx, s.formatRedisKey(REDIS_KEY_TEAMS)).Result()
}

// TeamName returns team name given a team ID.
func (s *RedisState) TeamName(teamID string) (string, error) {
	team_name, err := s.redis_client.HGet(s.ctx, s.formatRedisKey(REDIS_KEY_TEAMS), teamID).Result()

	if err != nil {
		return "", fmt.Errorf("unregistered team ID: %s", teamID)
	}

	return team_name, nil
}

// SetTeamName writes out team name.
// This can only be done once per team.
func (s *RedisState) SetTeamName(teamID, teamName string) error {
	valid_id, err := s.redis_client.SIsMember(s.ctx, s.formatRedisKey(REDIS_KEY_TEAM_IDS), teamID).Result()

	if err != nil {
		return fmt.Errorf("Unexpected error while validating team ID: %s", teamID)
	} else if (!valid_id) {
		return fmt.Errorf("team ID: (%s) not found in list of valid team IDs", teamID)
	}

	exists, err := s.redis_client.HExists(s.ctx, s.formatRedisKey(REDIS_KEY_TEAMS), teamID).Result()

	if exists {
		return nil
	}

	success, err := s.redis_client.HSetNX(s.ctx,  s.formatRedisKey(REDIS_KEY_TEAMS), teamID, teamName).Result()

	if err != nil {
		fmt.Println(err)
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
		Min: "-inf",
		Max: "+inf",
	}
	scores, err := s.redis_client.ZRangeByScoreWithScores(s.ctx, s.formatRedisKey(REDIS_KEY_POINT_LOG), redis_args).Result()

	if err != nil {
		fmt.Println("Encountered an error processing points")
		return make(award.List, 0)
	}
	
	var point_log award.List
	
	for _, item := range scores {
		point_entry := award.T{}

		point_string := strings.TrimSpace(item.Member.(string))

		point_entry.When = int64(item.Score)
		n, err := fmt.Sscanf(point_string, "%s %s %d", &point_entry.TeamID, &point_entry.Category, &point_entry.Points)

		if err != nil {
			// Do nothing
			fmt.Println("Encountered an error while extracting fields from ", point_string)
		} else if n != 3 {
			// Wrong number of fields, do nothing
			fmt.Println("Point entry is malformed, ", point_string)
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

	point_string := fmt.Sprintf("%s %s %d", teamID, category, points)

	new_member := redis.Z{
		Score: float64(awardTime),
		Member: point_string,
	}

	redis_args.Members = append(redis_args.Members, new_member)

	_, err := s.redis_client.ZAddArgs(s.ctx, s.formatRedisKey(REDIS_KEY_POINT_LOG), redis_args).Result()

	if err != nil {
		return err
	}

	return nil
}

// LogEvent writes to the event log
func (s *RedisState) LogEvent(event string, participantID string, teamID string, cat string, points int, extra ...string) {
	extra_data, _ := json.Marshal(extra)

	s.eventStream <- 
		map[string]interface{}{
			"When": time.Now().Unix(),
			"Event": event,
			"ParticipantID": participantID,
			"TeamID": teamID,
			"Category": cat,
			"Points": strconv.Itoa(points),
			"Extra": extra_data,
		}
}

func (s *RedisState) writeEvent(event map[string]interface{}) {
	redis_args := redis.XAddArgs {
		Stream: s.formatRedisKey(REDIS_KEY_EVENT_LOG),
		Values: event,
	}

	_, err := s.redis_client.XAdd(s.ctx, &redis_args).Result()

	if err != nil {
		fmt.Println("Error when processing event stream")
		fmt.Println(err)
	}
}


func (s *RedisState) Maintain(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)

	for {
		select {
		case msg := <-s.eventStream:
			s.writeEvent(msg)
		case <-ticker.C:
			/* 	There are no maintanance tasks for this provider, currently. Maybe some state-saving mechanism, at some point?	*/
		}
	}
}