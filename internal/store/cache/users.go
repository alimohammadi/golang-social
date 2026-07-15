package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alimohammadi/golan-social.git/internal/store"
	"github.com/go-redis/redis/v8"
)

type UsersStore struct {
	rdb *redis.Client
}

const UserExpireTime = time.Minute // 1 minutes

func (s *UsersStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)
	result, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}
	
	user := &store.User{}
	err = json.Unmarshal([]byte(result), user)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *UsersStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	jsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	
	err = s.rdb.SetEX(ctx, cacheKey, jsonData, UserExpireTime).Err()
	if err != nil {
		return err
	}
	
	return nil
}
