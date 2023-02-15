package spi

type OutputStream interface {
	Write([]byte) (int, error)
	Flush() error
	Close() error
}
