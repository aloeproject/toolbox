package redisgo

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"math/rand"
	"strings"
	"time"
)

type option struct {
	RedisConfig
}

type Option func(*option)

type RedisConfig struct {
	//以下均来自redisgo原生配置
	// 内部重试次数
	Retry int `json:"retry"`
	// Redis服务器的host和port "localhost:6379"
	Addr string `json:"addr"`

	// 连接到redis服务器的密码
	Password string `json:"password"`

	// 在连接池中可以存在的最大空闲连接数
	MaxIdle int `json:"max_idle"`

	//连接池中最大空闲连接数
	MaxActive int `json:"max_active"`

	//在此空闲时间后关闭连接。如果该值为零，则不关闭空闲连接。应用程序应将超时设置为小于服务器超时的值 单位：秒
	IdleTimeout int `json:"idle_timeout"`

	//以下单位为：毫秒
	ConnectTimeout int `json:"connect_timeout"`
	ReadTimeout    int `json:"read_timeout"`
	WriteTimeout   int `json:"write_timeout"`

	Database int `json:"database"`
}

func WithAddr(addr string) Option {
	return func(o *option) {
		o.Addr = addr
	}
}

func WithPassword(ps string) Option {
	return func(o *option) {
		o.Password = ps
	}
}

func WithMaxIdle(maxIdle int) Option {
	return func(o *option) {
		o.MaxIdle = maxIdle
	}
}

func WithMaxActive(maxActive int) Option {
	return func(o *option) {
		o.MaxActive = maxActive
	}
}

func WithIdleTimeout(idleTimeout int) Option {
	return func(o *option) {
		o.IdleTimeout = idleTimeout
	}
}

func WithConnectTimeout(connectTimeout int) Option {
	return func(o *option) {
		o.ConnectTimeout = connectTimeout
	}
}

func WithReadTimeout(readTimeout int) Option {
	return func(o *option) {
		o.ReadTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout int) Option {
	return func(o *option) {
		o.WriteTimeout = writeTimeout
	}
}

func WithDatabase(database int) Option {
	return func(o *option) {
		o.Database = database
	}
}

func NewRedisgo(opts ...Option) *Redisgo {
	defaultOpt := &option{
		RedisConfig: RedisConfig{
			Addr:           "127.0.0.1:6379",
			MaxIdle:        100,
			MaxActive:      100,
			IdleTimeout:    20,
			ConnectTimeout: 300, //单位为毫秒
			ReadTimeout:    50,
			WriteTimeout:   50,
			Database:       0,
			Retry:          0,
		},
	}
	for _, o := range opts {
		o(defaultOpt)
	}

	return newRedis(&defaultOpt.RedisConfig)
}

func NewConfRedisgo(conf RedisConfig) *Redisgo {
	return newRedis(&conf)
}

func newRedis(o *RedisConfig) *Redisgo {
	opts := []redis.DialOption{}
	opts = append(opts, redis.DialConnectTimeout(time.Duration(o.ConnectTimeout)*time.Millisecond))
	opts = append(opts, redis.DialReadTimeout(time.Duration(o.ReadTimeout)*time.Millisecond))
	opts = append(opts, redis.DialWriteTimeout(time.Duration(o.WriteTimeout)*time.Millisecond))
	if len(o.Password) != 0 {
		opts = append(opts, redis.DialPassword(o.Password))
	}
	opts = append(opts, redis.DialDatabase(o.Database))
	pool := redisinit(o.Addr, o.Password, o.MaxIdle, o.IdleTimeout, o.MaxActive, opts...)
	oo := *o

	return &Redisgo{
		pool:     pool,
		opts:     &oo,
		lastTime: time.Now().UnixNano(),
	}
}

func redisinit(server, password string, maxIdle, idleTimeout, maxActive int, options ...redis.DialOption) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		MaxActive:   maxActive,
		Dial: func() (redis.Conn, error) {
			var c redis.Conn
			var err error
			protocol := "tcp"
			if strings.HasPrefix(server, "unix://") {
				server = strings.TrimLeft(server, "unix://")
				protocol = "unix"
			}
			c, err = redis.Dial(protocol, server, options...)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// Do 函数为通用的函数， 可以执行任何redis服务器支持的命令 如果涉及到没有封装的命令 可以用此命令调用原始命令
func (r *Redisgo) Do(ctx context.Context, cmd string, args ...interface{}) (reply interface{}, err error) {
	return r.do(ctx, cmd, nil, args...)
}

func (r *Redisgo) randomDuration(n int64) time.Duration {
	s := rand.NewSource(r.lastTime)
	return time.Duration(rand.New(s).Int63n(n) + 1)
}

func (r *Redisgo) do(ctx context.Context, cmd string, f func(interface{}, error) (interface{}, error), args ...interface{}) (reply interface{}, err error) {
	var (
		count = 0
	)

retry1:
	client, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err = client.Do(cmd, args...)

	if f != nil {
		reply, err = f(reply, err)
	}

	if err == redis.ErrNil {
		err = nil
	}

	// if err is not redis.Error, it will retry
	if _, ok := err.(redis.Error); err != nil && !ok {
		var rterr error
		switch err {
		case redis.ErrPoolExhausted:
			rterr = ErrConnExhausted
		default:
			if strings.Contains(err.Error(), "timeout") {
				rterr = ErrTimeout
			} else {
				rterr = err
			}
		}

		if r.opts.Retry > 0 && count < r.opts.Retry {
			count++
			time.Sleep(time.Millisecond * r.randomDuration(10))
			goto retry1
		}

		return nil, rterr
	}

	return
}

type Redisgo struct {
	pool     *redis.Pool
	opts     *RedisConfig
	lastTime int64
}
