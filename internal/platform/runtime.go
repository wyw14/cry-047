package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type RandomIDs struct{}

func (RandomIDs) New(prefix string) domain.ID {
	buffer := make([]byte, 10)
	if _, err := rand.Read(buffer); err != nil {
		return domain.ID(fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()))
	}
	return domain.ID(prefix + "-" + hex.EncodeToString(buffer))
}

type ObjectCatalog struct {
	mu      sync.RWMutex
	objects map[string]bool
	failure error
}

func NewObjectCatalog(keys ...string) *ObjectCatalog {
	catalog := &ObjectCatalog{objects: make(map[string]bool)}
	for _, key := range keys {
		catalog.objects[key] = true
	}
	return catalog
}

func (c *ObjectCatalog) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.failure != nil {
		return false, c.failure
	}
	return c.objects[key], nil
}

func (c *ObjectCatalog) Add(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.objects[key] = true
}

func (c *ObjectCatalog) Fail(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failure = err
}
