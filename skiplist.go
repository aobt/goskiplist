package goskiplist

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	DefaultSkipStepSize = 4
	DefaultSkipLevel    = 16
	MaxSkipLevel        = 64
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Key interface {
	~string |
		~int | ~int64 | ~int32 | ~int16 | ~int8 |
		~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8 |
		~float64 | ~float32
}

func NewSkipList[K Key, V any](stepSize, level int) *SkipList[K, V] {

	if stepSize < 2 {
		stepSize = DefaultSkipStepSize
	}

	if level < 0 || level > MaxSkipLevel {
		level = DefaultSkipLevel
	}

	sl := SkipList[K, V]{
		stepSize: stepSize,
		level:    level,
		lock:     &sync.RWMutex{},
		head: &Elem[K, V]{
			next: make([]*Elem[K, V], level+1),
		},
	}
	return &sl
}

type SkipList[K Key, V any] struct {
	stepSize int
	level    int

	lock   *sync.RWMutex
	length int32
	head   *Elem[K, V] // must not be nil
}

type Elem[K Key, V any] struct {
	next []*Elem[K, V] // skip index list, level0, level1, level2...
	// kv
	k K // read only
	v *V
}

//================================================================
// SkipList

// StepSize return the stepSize of list
func (s *SkipList[K, V]) StepSize() int {
	return s.stepSize
}

// Level return the level of list
func (s *SkipList[K, V]) Level() int {
	return s.level
}

// Length return the length of list
func (s *SkipList[K, V]) Length() int32 {
	return atomic.LoadInt32(&s.length)
}

// randomLevel
func (s *SkipList[K, V]) randomLevel() int {
	l := 0
	for l < s.level && rand.Intn(s.stepSize) == 0 {
		l++
	}
	return l
}

// Put elem into list
func (s *SkipList[K, V]) Put(k K, v *V) error {
	lv := s.randomLevel()
	e := Elem[K, V]{
		next: make([]*Elem[K, V], lv+1),
		k:    k,
		v:    v,
	}
	pres := make([]*Elem[K, V], lv+1)

	s.lock.Lock()
	defer s.lock.Unlock()
	el := s.head
	for i := s.level; i > -1; i-- {
		for el.next[i] != nil && e.k >= el.next[i].k {
			el = el.next[i]
			if e.k == el.k { // just replace
				el.v = e.v
				return nil
			}
		}
		if i > lv {
			continue
		}
		pres[i] = el
	}
	for i := lv; i > -1; i-- {
		e.next[i] = pres[i].next[i]
		pres[i].next[i] = &e
	}
	atomic.AddInt32(&s.length, 1)
	return nil
}

// Find elem from list
func (s *SkipList[K, V]) Find(k K) (*V, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	el := s.head
	for i := s.level; i > -1; i-- {
		for el.next[i] != nil && k >= el.next[i].k {
			el = el.next[i]
			if k == el.k {
				return el.v, nil
			}
		}
	}
	return nil, ErrNotFound
}

// FindMin find the min elem from list
func (s *SkipList[K, V]) FindMin() (*V, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.head.next[0] == nil {
		return nil, ErrNotFound
	}
	return s.head.next[0].v, nil
}

// FindMax find the max elem from list
func (s *SkipList[K, V]) FindMax() (*V, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	el := s.head
	for i := s.level; i > -1; i-- {
		for el.next[i] != nil {
			el = el.next[i]
		}
	}
	if el == s.head {
		return nil, ErrNotFound
	}
	return el.v, nil
}

// Pop elem from list
func (s *SkipList[K, V]) Pop(k K) (*V, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	el := s.head
	var (
		has bool
		v   *V
	)
	for i := s.level; i > -1; i-- {
		for el.next[i] != nil && k >= el.next[i].k {
			if k == el.next[i].k {
				has = true
				v = el.next[i].v
				el.next[i] = el.next[i].next[i]
				break
			}
			el = el.next[i]
		}
	}
	if has {
		atomic.AddInt32(&s.length, -1)
		return v, nil
	}
	return nil, ErrNotFound
}

// PopMin pop the min elem from list
func (s *SkipList[K, V]) PopMin() (*V, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.head.next[0] == nil {
		return nil, ErrNotFound
	}
	el := s.head.next[0]
	for i := len(el.next) - 1; i > -1; i-- {
		s.head.next[i] = el.next[i]
	}
	atomic.AddInt32(&s.length, -1)
	return el.v, nil
}

// PopMax pop the max elem from list
func (s *SkipList[K, V]) PopMax() (*V, error) {
	pres := make([]*Elem[K, V], s.level+1)

	s.lock.Lock()
	defer s.lock.Unlock()
	el := s.head
	var pre *Elem[K, V]
	for i := s.level; i > -1; i-- {
		for el.next[i] != nil {
			pre = el
			el = el.next[i]
		}
		for pre != nil && pre.next[i].k < el.k {
			pre = pre.next[i]
		}
		pres[i] = pre
	}
	if el == s.head {
		return nil, ErrNotFound
	}
	for i := len(el.next) - 1; i > -1; i-- {
		pres[i].next[i] = nil
	}
	atomic.AddInt32(&s.length, -1)
	return el.v, nil
}
