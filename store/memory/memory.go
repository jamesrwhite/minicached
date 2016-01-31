package memory

import (
	"sync"
)

type Store struct {
	records map[string]*Record
	// TODO: in certain situations can we only lock the record instead of the entire Store?
	sync.RWMutex
}

type Record struct {
	Key    string
	Value  []byte
	Flags  int64
	Ttl    int64
	Length int64
}

func Init() *Store {
	return &Store{
		records: make(map[string]*Record),
	}
}

func (s Store) Get(key string) *Record {
	s.RLock()
	defer s.RUnlock()

	value, ok := s.records[key]

	if ok {
		return value
	}

	return nil
}

func (s Store) Set(record *Record) {
	s.Lock()
	defer s.Unlock()

	s.records[record.Key] = record
}

func (s Store) Delete(key string) {
	s.Lock()
	defer s.Unlock()

	delete(s.records, key)
}

func (s Store) Flush() {
	s.Lock()
	defer s.Unlock()

	s.records = make(map[string]*Record)
}
