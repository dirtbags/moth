package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheResolver is a UserResolver that caches whatever's returned
type CacheResolver struct {
	resolver   UserResolver
	rdb        *redis.Client
	expiration time.Duration
}

// NewCacheResolver returns a new CacheResolver
//
// Items will be cached in rdb with an expration of expiration.
func NewCacheResolver(resolver UserResolver, rdb *redis.Client, expiration time.Duration) *CacheResolver {
	return &CacheResolver{
		rdb:        rdb,
		resolver:   resolver,
		expiration: expiration,
	}
}

// Resolve resolves an eventID and userID.
//
// It checks the cache first. If a match is found, that is returned.
// If not, it passes the request along to the upstream Resolver,
// caches the result, and returns it.
func (cr *CacheResolver) Resolve(eventID string, userID string) (string, error) {
	key := fmt.Sprintf("username:%s|%s", eventID, userID)
	name, err := cr.rdb.Get(context.TODO(), key).Result()
	if err == nil {
		// Cache hit
		return name, nil
	}

	name, err = cr.resolver.Resolve(eventID, userID)
	if err != nil {
		return "", err
	}

	cr.rdb.Set(context.TODO(), key, name, cr.expiration)

	return name, nil
}
