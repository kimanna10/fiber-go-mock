package cache

import (
	"context"
	"time"
)

// Cache - интерфейс для работы с кэшем. В данном случае, это может быть Redis или любой другой кэш.
type Cache interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
}
