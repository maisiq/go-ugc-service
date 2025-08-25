package clickhouse

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/maisiq/go-ugc-service/pkg/config"
)

var (
	client driver.Conn
	once   sync.Once
)

func InitClickhouseClient(ctx context.Context, cfg *config.ClickhouseConfig) (driver.Conn, error) {

	var dialCount int32
	var conn driver.Conn
	var err error

	once.Do(func() {
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{cfg.DSN},
			Auth: clickhouse.Auth{
				Database: cfg.DatabaseName,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
				dialCount++
				var d net.Dialer
				return d.DialContext(ctx, "tcp", addr)
			},
			Debug: true,
			Debugf: func(format string, v ...any) {
				fmt.Printf(format+"\n", v...)
			},
			Settings: clickhouse.Settings{
				"max_execution_time": 60,
			},
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
			DialTimeout:          time.Second * 30,
			MaxOpenConns:         5,
			MaxIdleConns:         5,
			ConnMaxLifetime:      time.Duration(10) * time.Minute,
			ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
			BlockBufferSize:      10,
			MaxCompressionBuffer: 10240,
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "api", Version: "0.1.0"},
				},
			},
		})
	})

	if err != nil {
		return nil, err
	}

	if client == nil {
		client = conn
	}

	return client, err
}

func GetClickhouseClient() driver.Conn {
	if client == nil {
		panic("Clickhouse Client is not initialized")
	}
	return client
}
