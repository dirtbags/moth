package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dirtbags/moth/pkg/award"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	rdb    *redis.Client
	prefix string
}

// NewRedis returns a new Redis structure
func NewRedis(rdb *redis.Client, prefix string) Redis {
	return Redis{
		rdb:    rdb,
		prefix: prefix,
	}
}

func (r Redis) key(path ...string) string {
	path = append([]string{r.prefix}, path...)
	return strings.Join(path, "/")
}

// Messages returns all broadcast messages
func (r Redis) Messages() string {
	messages, err := r.rdb.Get(
		context.TODO(),
		r.key("messages.html"),
	).Result()
	if err != nil {
		return ""
	}
	return messages
}

func (r Redis) PointsLog() award.List {
	pointsLog, err := r.rdb.ZRangeArgsWithScores(
		context.TODO(),
		redis.ZRangeArgs{
			Key:   r.key("points.log"),
			Start: math.Inf(-1),
			Stop:  math.Inf(+1),
		},
	).Result()
	if err != nil {
		return nil
	}

	ret := make(award.List, len(pointsLog), 0)
	for _, entry := range pointsLog {
		// XXX: Fix this kludge
		line := fmt.Sprintf("%d %s", int64(math.Round(entry.Score)), entry.Member.(string))
		aent, err := award.Parse(line)
		if err != nil {
			continue
		}
		ret = append(ret, aent)
	}
	return ret
}

func (r Redis) TeamName(teamID string) (string, error) {
	name, err := r.rdb.Get(
		context.TODO(),
		r.key("teams", teamID),
	).Result()
	if err != nil {
		return "", fmt.Errorf("No such team")
	}
	return name, nil
}

func (r Redis) SetTeamName(teamID, teamName string) error {
	_, err := r.rdb.SetArgs(
		context.TODO(),
		r.key("teams", teamID),
		teamName,
		redis.SetArgs{
			Mode: "NX",
		},
	).Result()
	return err
}

func (r Redis) AwardPoints(teamID string, cat string, points int) error {
	// XXX: Add to award something that makes a string without timestamp
}

func (r Redis) Maintain(updateInterval time.Duration) {
	// No maintenance required
}
