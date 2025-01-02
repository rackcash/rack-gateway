package service

import (
	"infra/api/internal/infra/cache"
)

type LockerService struct {
	cache *cache.Cache
}

func NewLockerService(cache *cache.Cache) *LockerService {
	return &LockerService{cache: cache}
}

func (s *LockerService) Lock(key string) {
	s.cache.SetNoExp(key, true)
}

func (s *LockerService) Unlock(key string) {
	s.cache.Del(key)
}

func (s *LockerService) IsLocked(key string) bool {
	return s.cache.Load(key) != nil // locked if not nil
}
