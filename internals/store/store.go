package store

import (
	"fmt"
	"sync"
	"time"
)

type ValueType int

const (
	TypeString ValueType = iota
	TypeList
	TypeHash
	TypeSet
)

type valueStore struct {
	kind      ValueType
	strVal    string
	listVal   []string
	hashVal   map[string]string
	setVal    map[string]struct{}
	expiresAt time.Time
}

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]valueStore
}

func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		items: make(map[string]valueStore),
	}

	return store
}

func (s *InMemoryStore) Get(key string) (string, bool, error) {
	s.mu.RLock()
	val, found := s.items[key]
	if !found || !isExpired(val) {
		s.mu.RUnlock()
		if !found {
			return "", false, nil
		}
		if err := checktype(val, TypeString); err != nil {
			return "", false, err
		}
		return val.strVal, true, nil
	}

	s.mu.RUnlock()

	// slow path: it was expired, so escalate to a write lock to clean it up
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found = s.items[key] // re-check!
	if !found || !isExpired(val) {
		if !found {
			return "", false, nil
		}
		return val.strVal, true, nil
	}
	delete(s.items, key)
	return "", false, nil
}

func (s *InMemoryStore) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val := valueStore{
		strVal: value,
	}
	s.items[key] = val
}

func (s *InMemoryStore) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found {
		return false
	}
	delete(s.items, key)
	return !isExpired(val)
}

func (s *InMemoryStore) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found {
		return false
	}
	if isExpired(val) {
		delete(s.items, key)
		return false
	}
	return true
}

func (s *InMemoryStore) Expire(key string, seconds int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found || isExpired(val) {
		return false
	}

	val.expiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
	s.items[key] = val
	return true
}

func (s *InMemoryStore) Ttl(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, found := s.items[key]

	if !found || isExpired(val) {
		return -2
	} else if val.expiresAt.IsZero() {
		return -1
	}
	return int64(time.Until(val.expiresAt).Seconds())

}
func (s *InMemoryStore) Lpush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found || isExpired(val) {
		val = valueStore{kind: TypeList}
	} else if err := checktype(val, TypeList); err != nil {
		return -1, err
	}

	newList := make([]string, 0, len(values)+len(val.listVal))
	for i := len(values) - 1; i >= 0; i-- {
		newList = append(newList, values[i])
	}
	newList = append(newList, val.listVal...)

	val.listVal = newList
	s.items[key] = val
	return len(newList), nil
}

func (s *InMemoryStore) Rpush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found || isExpired(val) {
		val = valueStore{kind: TypeList}
	} else if err := checktype(val, TypeList); err != nil {
		return -1, err
	}

	newList := make([]string, 0, len(values)+len(val.listVal))
	newList = append(newList, val.listVal...)
	newList = append(newList, values...)

	val.listVal = newList
	s.items[key] = val
	return len(val.listVal), nil
}

func (s *InMemoryStore) Lrange(key string, start int, stop int) ([]string, bool, error) {
	if start < 0 || stop < 0 {
		return nil, false, fmt.Errorf("WRONG-ARGUMENTS start and stop cant be negative as of now")
	}
	s.mu.RLock()
	val, found := s.items[key]
	if !found || !isExpired(val) {
		s.mu.RUnlock()
		if !found {
			return nil, false, nil
		}
		if err := checktype(val, TypeList); err != nil {
			return nil, false, err
		}
		if stop >= len(val.listVal) {
			stop = len(val.listVal) - 1
		}
		if start < 0 || start > stop || len(val.listVal) == 0 {
			return nil, false, nil

		}
		return []string(val.listVal[start : stop+1]), true, nil
	}

	s.mu.RUnlock()

	// slow path: it was expired, so escalate to a write lock to clean it up
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found = s.items[key] // re-check!
	if !found || !isExpired(val) {
		if !found {
			return nil, false, nil
		}
		return []string(val.listVal[start:stop]), true, nil
	}
	delete(s.items, key)
	return nil, false, nil
}
func (s *InMemoryStore) Hset(key string, field string, value string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found || isExpired(val) {
		val = valueStore{
			kind:    TypeHash,
			hashVal: map[string]string{},
		}
	} else if err := checktype(val, TypeHash); err != nil {
		return false, err
	}
	prevH := val.hashVal
	_, alreadyExist := prevH[field]
	val.hashVal[field] = value
	s.items[key] = val
	if alreadyExist {
		return false, nil
	}
	return true, nil

}

func (s *InMemoryStore) Hget(key string, field string) (string, bool, error) {
	s.mu.RLock()
	val, found := s.items[key]
	if !found || !isExpired(val) {
		s.mu.RUnlock()
		if !found {
			return "", false, nil
		}
		if err := checktype(val, TypeHash); err != nil {
			return "", false, err
		}
		fieldVal, fieldCheck := val.hashVal[field]
		if fieldCheck {
			return fieldVal, true, nil
		} else {
			return "", false, nil
		}
	}

	s.mu.RUnlock()

	// slow path: it was expired, so escalate to a write lock to clean it up
	s.mu.Lock()
	defer s.mu.Unlock()
	val, found = s.items[key] // re-check!
	if !found || !isExpired(val) {
		if !found {
			return "", false, nil
		}
		fieldVal, fieldFound := val.hashVal[field]
		return fieldVal, fieldFound, nil
	}
	delete(s.items, key)
	return "", false, nil
}
func (s *InMemoryStore) HgetALL(key string) (map[string]string, bool, error) {
	s.mu.RLock()
	val, found := s.items[key]
	if !found || !isExpired(val) {
		s.mu.RUnlock()
		if !found {
			return nil, false, nil
		}
		if err := checktype(val, TypeHash); err != nil {
			return nil, false, err
		}
		copied := make(map[string]string, len(val.hashVal))
		for k, v := range val.hashVal {
			copied[k] = v
		}
		return copied, true, nil
	}

	s.mu.RUnlock()

	// slow path: it was expired, so escalate to a write lock to clean it up
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found = s.items[key] // re-check!
	if !found || !isExpired(val) {
		if !found {
			return nil, false, nil
		}
		copied := make(map[string]string, len(val.hashVal))
		for k, v := range val.hashVal {
			copied[k] = v
		}
		return copied, true, nil
	}
	delete(s.items, key)
	return nil, false, nil
}

func (s *InMemoryStore) Sadd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]

	if !found || isExpired(val) {
		val = valueStore{kind: TypeSet, setVal: map[string]struct{}{}}
	} else if err := checktype(val, TypeSet); err != nil {
		return 0, err
	}
	newCnt := 0
	for _, member := range members {
		if _, exists := val.setVal[member]; !exists {
			val.setVal[member] = struct{}{}
			newCnt++
		}
	}
	s.items[key] = val
	return newCnt, nil
}

func (s *InMemoryStore) Smembers(key string) ([]string, error) {
	s.mu.RLock()
	val, found := s.items[key]
	if !found || !isExpired(val) {
		s.mu.RUnlock()
		if !found {
			return []string{}, nil
		}
		if err := checktype(val, TypeSet); err != nil {
			return []string{}, err
		}
		members := make([]string, 0, len(val.setVal))
		for member := range val.setVal {
			members = append(members, member)
		}
		return members, nil
	}

	s.mu.RUnlock()

	// slow path: it was expired, so escalate to a write lock to clean it up
	s.mu.Lock()
	defer s.mu.Unlock()
	val, found = s.items[key] // re-check!
	if !found || !isExpired(val) {
		if !found {
			return []string{}, nil
		}
		members := make([]string, 0, len(val.setVal))
		for member := range val.setVal {
			members = append(members, member)
		}
		return members, nil
	}
	delete(s.items, key)
	return []string{}, nil
}
