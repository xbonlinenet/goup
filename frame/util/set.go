package util

import "sync"

// 线程不安全的 set
type UnsafeSet struct {
	m map[string]struct{}
}

func (set UnsafeSet) Contains(s string) bool {
	if _, ok := set.m[s]; ok {
		return true
	}

	return false
}

func NewUnsafeSet(ss ...string) *UnsafeSet {
	set := new(UnsafeSet)
	set.m = make(map[string]struct{}, len(ss))

	for _, s := range ss {
		set.m[s] = struct{}{}
	}

	return set
}

type SafeSet struct {
	m     map[string]struct{}
	mLock sync.RWMutex
}

func NewSafeSet(ss ...string) *SafeSet {
	set := new(SafeSet)
	set.m = make(map[string]struct{}, 2*len(ss))
	set.Add(ss...)
	return set
}

func (set SafeSet) Add(ss ...string) {
	set.mLock.Lock()
	defer set.mLock.Unlock()

	for _, s := range ss {
		set.m[s] = struct{}{}
	}
}

func (set SafeSet) Del(ss ...string) {
	set.mLock.Lock()
	defer set.mLock.Unlock()

	for _, s := range ss {
		delete(set.m, s)
	}
}

func (set SafeSet) Contains(s string) bool {
	set.mLock.RLock()
	defer set.mLock.RUnlock()

	if _, ok := set.m[s]; ok {
		return true
	}

	return false
}
