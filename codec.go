package spi

import "time"

type RowsEncoderContext struct {
	Sink         Sink
	Rownum       bool
	Heading      bool
	TimeLocation *time.Location
	TimeFormat   string
	Precision    int
}

type RowsEncoder interface {
	Open(colums Columns) error
	Close()
	AddRow(values []any) error
	Flush(heading bool)
	ContentType() string
}
