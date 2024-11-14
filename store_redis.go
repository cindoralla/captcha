// Contributed 2024 by Cindoralla
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

const (
	// DefaultMaxRedisKeys default max redis keys per expiration
	DefaultMaxRedisKeys = 500000
	// DefaultRedisPrefixKey default redis prefix key
	DefaultRedisPrefixKey = "captcha"
)

// redisStore is an internal store for captcha ids and their values.
type redisStore struct {
	redisClient *redis.Client
	expiration  time.Duration
	maxKeys     int64
	prefixKey   string
}

// NewRedisStore returns new Redis memory store
func NewRedisStore(redisOptions *redis.Options, expiration time.Duration, maxKeys int64, prefixKey string) (Store, error) {
	if redisOptions == nil {
		return nil, fmt.Errorf("invalid redis options: %v", redisOptions)
	}
	s := new(redisStore)
	s.redisClient = redis.NewClient(redisOptions)
	s.expiration = expiration
	s.maxKeys = maxKeys
	if s.maxKeys <= 100 {
		s.maxKeys = DefaultMaxRedisKeys
	}
	s.prefixKey = prefixKey
	if s.prefixKey == "" {
		s.prefixKey = DefaultRedisPrefixKey
	}

	return s, nil
}

// TransferDigitsToStr transfers digits' bytes to string
func TransferDigitsToStr(digits []byte) string {

	var strs []string

	for _, d := range digits {
		strs = append(strs, string(d+'0'))
	}

	return strings.Join(strs, "")
}

// TransferStrToDigits transfers string to digits' bytes
func TransferStrToDigits(str string) []byte {
	ns := make([]byte, len(str))

	for i := range ns {
		d := str[i]
		ns[i] = d - '0'
	}

	return ns
}

func (s *redisStore) Set(id string, digits []byte) {
	c, err := s.redisClient.DbSize().Result()
	if err != nil {
		panic(err)
	}
	if c > s.maxKeys {
		panic(fmt.Errorf("to many keys > %v", s.maxKeys))
	}

	id = fmt.Sprintf("%s:%s", s.prefixKey, id)
	_, err = s.redisClient.Get(id).Result()
	if err == redis.Nil {
		str := TransferDigitsToStr(digits)
		s.redisClient.Set(id, str, s.expiration)
	}
}

func (s *redisStore) Get(id string, clear bool) (digits []byte) {
	id = fmt.Sprintf("%s:%s", s.prefixKey, id)
	val, err := s.redisClient.Get(id).Result()
	if err == redis.Nil {
		return digits
	}

	// digits = []byte(val)
	digits = TransferStrToDigits(val)

	if clear {
		if err != redis.Nil {
			s.redisClient.Del(id)
		}
	}
	return digits
}
