package memory

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/jamesrwhite/minicached/store"
)

var memoryStore = make(map[string]store.Record)
var lock = sync.RWMutex{}

func init() {
	log.Info("Initialising memory storage engine")
}

func Get(key string) (found bool, record store.Record) {
	lock.RLock()
	defer lock.RUnlock()

	value, ok := memoryStore[key]

	if ok {
		return true, value
	}

	return false, store.Record{}
}

func Set(key, value string, length, flags, ttl int64) {
	lock.Lock()
	defer lock.Unlock()

	memoryStore[key] = store.Record{
		Key:    key,
		Value:  value,
		Flags:  flags,
		Ttl:    ttl,
		Length: length,
	}
}

func Delete(key string) {
	lock.Lock()
	defer lock.Unlock()

	delete(memoryStore, key)
}

func Flush() {
	lock.Lock()
	defer lock.Unlock()

	memoryStore = make(map[string]store.Record)
}
