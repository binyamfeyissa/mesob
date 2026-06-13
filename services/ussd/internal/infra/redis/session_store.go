package redisinfra

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/mesob-wallet/ussd/internal/domain"
)

const sessionTTL = 180 * time.Second

type SessionStore struct {
	Redis *goredis.Client
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	data, err := s.Redis.Get(ctx, "ussd:"+sessionID).Bytes()
	if err != nil {
		return nil, err
	}
	var sess domain.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *SessionStore) Save(ctx context.Context, sess *domain.Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return s.Redis.Set(ctx, "ussd:"+sess.ID, data, sessionTTL).Err()
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.Redis.Del(ctx, "ussd:"+sessionID).Err()
}
