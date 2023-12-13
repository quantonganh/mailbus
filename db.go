package mailbus

type Database interface {
	Open() error
	Close() error
}
