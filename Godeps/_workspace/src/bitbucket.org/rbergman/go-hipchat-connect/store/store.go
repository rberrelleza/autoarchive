package store

type Store interface {
	Key(scope string) string
	Del(k string) error
	Get(k string) ([]byte, error)
	Set(k string, v []byte) error
	SetEx(k string, v []byte, sec int) error
	Sub(scope string) Store
}
