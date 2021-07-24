package sessions

type Store interface {
	Get(id string) (*Session, error)
	Set(id string, session *Session) error
	Del(id string) error
	Range(page, size int64) ([]*Session, error)
	Count() (int64, error)
	Close() error
}
