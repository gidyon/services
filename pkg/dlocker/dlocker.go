// Package dlocker is a distributed locker for distributed objects whose access requires synchronization. The objects may reside in same process or different processes.
package dlocker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var lockLeaseTime = 10 * time.Second

var (
	// ErrLeaseTimeExceeded happens when lock is held for longer that lease time
	ErrLeaseTimeExceeded = fmt.Errorf("maximum lease time is %d seconds", 10)
	// ErrLockNotHeld happens when calling Release on a locker that is Released
	ErrLockNotHeld = errors.New("lock is released")
)

// Locker is a distributed locker for clients to implement
type Locker interface {
	Acquire(context.Context) error
	Release(context.Context) error
	Released() <-chan struct{}
}

type redisLocker struct {
	client  *redis.Client
	channel <-chan *redis.Message
	lockID  string
	set     string
}

// NewRedisLocker creates a distributed locker using redis pub sub and set data structure
func NewRedisLocker(ctx context.Context, client *redis.Client, lockID string) Locker {
	return &redisLocker{
		client:  client,
		channel: client.Subscribe(ctx, lockID).Channel(),
		lockID:  lockID,
		set:     uuid.New().String(),
	}
}

func (rl *redisLocker) acquireLock(ctx context.Context) error {
	v, err := rl.client.SAdd(ctx, rl.set, rl.lockID).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire token: %w", err)
	}
	if v == 1 {
		return nil
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to acquire token: %v", err)
	case <-rl.channel:
		return rl.acquireLock(ctx)
	}
}

func (rl *redisLocker) Acquire(ctx context.Context) error {
	if t, ok := ctx.Deadline(); ok {
		rem := t.Sub(time.Now())
		if rem > lockLeaseTime {
			return ErrLeaseTimeExceeded
		}
	}
	ch := make(chan struct{})
	var err error
	go func() {
		err = rl.acquireLock(ctx)
		close(ch)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return err
	}
}

func (rl *redisLocker) Release(ctx context.Context) error {
	pipeliner := rl.client.Pipeline()

	v, err := pipeliner.SRem(ctx, rl.set, rl.lockID).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if v == 0 {
		return ErrLockNotHeld
	}

	err = pipeliner.Publish(ctx, rl.lockID, "RELEASED").Err()
	if err != nil {
		return fmt.Errorf("failed to publish lock release: %v", err)
	}

	_, err = pipeliner.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute lock release: %v", err)
	}
	return nil
}

func (rl *redisLocker) Released() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	<-rl.channel
	return ch
}
