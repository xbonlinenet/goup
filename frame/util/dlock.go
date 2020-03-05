package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// DLock 基于 redis 的分布式锁
type DLock struct {
	Cmd   redis.Cmdable
	Key   string
	ReqID string
}

// NewDLock 创建锁实例
func NewDLock(client redis.Cmdable, key string, reqID string) *DLock {
	return &DLock{
		Cmd:   client,
		Key:   key,
		ReqID: reqID,
	}
}

// TryLock 尝试获取锁
func (lock *DLock) TryLock(expire time.Duration) (bool, error) {
	key := fmt.Sprintf("dlk:%s", lock.Key)
	return lock.Cmd.SetNX(key, lock.ReqID, expire).Result()
}

// TryLockWithWait 尝试获取锁，并等待
func (lock *DLock) TryLockWithWait(expire time.Duration, waitSeconds time.Duration) (bool, error) {
	result, err := lock.TryLock(expire)
	if err != nil {
		return false, err
	}

	if result {
		return result, nil
	}

	if waitSeconds <= 1*time.Second {
		time.Sleep(waitSeconds)
		result, err := lock.TryLock(expire)
		if err != nil {
			return false, err
		}

		if result {
			return result, nil
		}
	} else {
		for waitSeconds > 0 {
			time.Sleep(time.Second)
			waitSeconds = waitSeconds - time.Second
			result, err := lock.TryLock(expire)
			if err != nil {
				return false, err
			}

			if result {
				return result, nil
			}
		}
	}
	return false, nil
}

// ReleaseLock 释放锁
func (lock *DLock) ReleaseLock() (bool, error) {
	script := "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end"
	key := fmt.Sprintf("dlk:%s", lock.Key)
	value, err := lock.Cmd.Eval(script, []string{key}, lock.ReqID).Result()
	if err != nil {
		return false, err
	}

	count, ok := value.(int64)
	if !ok {
		return false, errors.New("redis server reture error")
	}

	if count >= 1 {
		return true, nil
	}
	return false, nil
}
