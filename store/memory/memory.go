package memory

type Store struct {
	records map[string]*Record
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
	value, ok := s.records[key]

	if ok {
		return value
	}

	return nil
}

func (s Store) Set(record *Record) {
	s.records[record.Key] = record
}

func (s Store) Delete(key string) {
	delete(s.records, key)
}

func (s Store) Flush() {
	s.records = make(map[string]*Record)
}
