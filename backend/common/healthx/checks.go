package healthx

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
)

func GormPing(db *gorm.DB) Check {
	return Check{
		Name: "mysql",
		Fn: func(ctx context.Context) error {
			if db == nil {
				return fmt.Errorf("nil gorm db")
			}
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.PingContext(ctx)
		},
	}
}

func RedisPing(name string, rdb *redis.Client) Check {
	return Check{
		Name: name,
		Fn: func(ctx context.Context) error {
			if rdb == nil {
				return fmt.Errorf("nil redis client")
			}
			return rdb.Ping(ctx).Err()
		},
	}
}

func GRPCHealthCheck(name string, conn *grpc.ClientConn, service string, timeout time.Duration) Check {
	return Check{
		Name: name,
		Fn: func(ctx context.Context) error {
			if conn == nil {
				return fmt.Errorf("nil grpc conn")
			}
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			cli := grpc_health_v1.NewHealthClient(conn)
			_, err := cli.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
			return err
		},
	}
}

func TCPDial(name, addr string, timeout time.Duration) Check {
	return Check{
		Name: name,
		Fn: func(ctx context.Context) error {
			if addr == "" {
				return fmt.Errorf("empty addr")
			}
			d := net.Dialer{}
			if timeout > 0 {
				d.Timeout = timeout
			}
			c, err := d.DialContext(ctx, "tcp", addr)
			if err != nil {
				return err
			}
			_ = c.Close()
			return nil
		},
	}
}
