package store

type Store interface {
	Get(key string) (bool, Record)
	Set(key, value string, length, flags, ttl int64)
	Delete(key string)
	Flush()
}

type Record struct {
	Key    string
	Value  string
	Flags  int64
	Ttl    int64
	Length int64
}
