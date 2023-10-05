package spi

import (
	"context"
	"fmt"
	"net"
	"time"
)

type Database interface {
	// Connect makes a new Conn
	Connect(ctx context.Context, options ...ConnectOption) (Conn, error)
}

// DatabaseServer represents a spi implementation for Database server
type DatabaseServer interface {
	Database
	Startup() error
	Shutdown() error
}

// DatabaseClient represents a spi implementation for Database client
type DatabaseClient interface {
	Database
	Close()
}

type DatabaseAux interface {
	// GetServerInfo gets ServerInfo
	GetServerInfo() (*ServerInfo, error)

	// GetInflights returns list of inflights statements
	GetInflights() ([]*Inflight, error)

	// GetPostflights returns list of postflights statements
	GetPostflights() ([]*Postflight, error)

	// GetServicePorts returns port info
	GetServicePorts(service string) ([]*ServicePort, error)
}

type Explainer interface {
	// Explain retrieves execution plan of the given SQL statement.
	Explain(ctx context.Context, sqlText string, full bool) (string, error)
}

type DatabaseAuth interface {
	UserAuth(user string, password string) (bool, error)
}

type ConnectOption func(Conn)

type Conn interface {
	// Close closes connection
	Close() error

	// ExecContext executes SQL statements that does not return result
	// like 'ALTER', 'CREATE TABLE', 'DROP TABLE', ...
	Exec(ctx context.Context, sqlText string, params ...any) Result

	// Query executes SQL statements that are expected multipe rows as result.
	// Commonly used to execute 'SELECT * FROM <TABLE>'
	//
	// Rows returned by Query() must be closed to prevent server-side-resource leaks.
	//
	//	ctx, cancelFunc := context.WithTimeout(5*time.Second)
	//	defer cancelFunc()
	//
	//	rows, err := conn.Query(ctx, "select * from my_table where name = ?", my_name)
	//	if err != nil {
	//		panic(err)
	//	}
	//	defer rows.Close()
	Query(ctx context.Context, sqlText string, params ...any) (Rows, error)

	// QueryRow executes a SQL statement that expects a single row result.
	//
	//	ctx, cancelFunc := context.WithTimeout(5*time.Second)
	//	defer cancelFunc()
	//
	//	var cnt int
	//	row := conn.QueryRow(ctx, "select count(*) from my_table where name = ?", "my_name")
	//	row.Scan(&cnt)
	QueryRow(ctx context.Context, sqlText string, params ...any) Row

	// Appender creates a new Appender for the given table.
	// Appender should be closed as soon as finshing work, otherwise it may cause server side resource leak.
	//
	//	ctx, cancelFunc := context.WithTimeout(5*time.Second)
	//	defer cancelFunc()
	//
	//	app, _ := conn.Appender(ctx, "MYTABLE")
	//	defer app.Close()
	//	app.Append("name", time.Now(), 3.14)
	Appender(ctx context.Context, tableName string, opts ...AppendOption) (Appender, error)
}

type Pinger interface {
	Ping() (time.Duration, error)
}

type Result interface {
	Err() error
	RowsAffected() int64
	Message() string
}

type Inflight struct {
	Id      string
	Type    string
	SqlText string
	Elapsed time.Duration
}

type Postflight struct {
	SqlText   string
	Count     int64
	TotalTime time.Duration
}

type ServerInfo struct {
	Version Version
	Runtime Runtime
}

type Version struct {
	Major          int32
	Minor          int32
	Patch          int32
	GitSHA         string
	BuildTimestamp string
	BuildCompiler  string
	Engine         string
}

type Runtime struct {
	OS             string
	Arch           string
	Pid            int32
	UptimeInSecond int64
	Processes      int32
	Goroutines     int32
	MemSys         uint64
	MemHeapSys     uint64
	MemHeapAlloc   uint64
	MemHeapInUse   uint64
	MemStackSys    uint64
	MemStackInUse  uint64
}

type ServicePort struct {
	Service string
	Address string
}

type Rows interface {
	// Next returns true if there are at least one more fetchable record remained.
	//
	//  rows, _ := db.Query("select name, value from my_table")
	//	for rows.Next(){
	//		var name string
	//		var value float64
	//		rows.Scan(&name, &value)
	//	}
	Next() bool

	// Scan retrieve values of columns in a row
	//
	//	for rows.Next(){
	//		var name string
	//		var value float64
	//		rows.Scan(&name, &value)
	//	}
	Scan(cols ...any) error

	// Close release all resources that assigned to the Rows
	Close() error

	// IsFetchable returns true if statement that produced this Rows was fetch-able (e.g was select?)
	IsFetchable() bool

	RowsAffected() int64
	Message() string

	// Columns returns list of column info that consists of result of query statement.
	Columns() (Columns, error)
}

type Row interface {
	Success() bool
	Err() error
	Scan(cols ...any) error
	Values() []any
	RowsAffected() int64
	Message() string
}

type Columns []*Column

type Column struct {
	Name   string
	Type   string
	Size   int
	Length int
}

func (cols Columns) Names() []string {
	names := make([]string, len(cols))
	for i := range cols {
		names[i] = cols[i].Name
	}
	return names
}

func (cols Columns) NamesWithTimeLocation(tz *time.Location) []string {
	names := make([]string, len(cols))
	for i := range cols {
		if cols[i].Type == "datetime" {
			names[i] = fmt.Sprintf("%s(%s)", cols[i].Name, tz.String())
		} else {
			names[i] = cols[i].Name
		}
	}
	return names
}

func (cols Columns) Types() []string {
	types := make([]string, len(cols))
	for i := range cols {
		types[i] = cols[i].Type
	}
	return types
}

const (
	ColumnBufferTypeInt16    = "int16"
	ColumnBufferTypeInt32    = "int32"
	ColumnBufferTypeInt64    = "int64"
	ColumnBufferTypeDatetime = "datetime"
	ColumnBufferTypeFloat    = "float"
	ColumnBufferTypeDouble   = "double"
	ColumnBufferTypeIPv4     = "ipv4"
	ColumnBufferTypeIPv6     = "ipv6"
	ColumnBufferTypeString   = "string"
	ColumnBufferTypeBinary   = "binary"
	ColumnBufferTypeBoolean  = "bool"
	ColumnBufferTypeByte     = "int8"
)

func (cols Columns) MakeBuffer() []any {
	rec := make([]any, len(cols))
	for i := range cols {
		switch cols[i].Type {
		case "int16":
			rec[i] = new(int16)
		case "int32":
			rec[i] = new(int32)
		case "int64":
			rec[i] = new(int64)
		case "datetime":
			rec[i] = new(time.Time)
		case "float":
			rec[i] = new(float32)
		case "double":
			rec[i] = new(float64)
		case "ipv4":
			rec[i] = new(net.IP)
		case "ipv6":
			rec[i] = new(net.IP)
		case "string":
			rec[i] = new(string)
		case "binary":
			rec[i] = new([]byte)
		case "bool":
			rec[i] = new(bool)
		case "int8":
			rec[i] = new(byte)
		}
	}
	return rec
}

type Appender interface {
	TableName() string
	TableType() TableType
	Columns() (Columns, error)
	Append(values ...any) error
	AppendWithTimestamp(ts time.Time, values ...any) error
	Close() (int64, int64, error)
}

type AppendOption interface {
	appendoption()
}

func (o AppendTimeformatOption) appendoption() {}

type AppendTimeformatOption string
