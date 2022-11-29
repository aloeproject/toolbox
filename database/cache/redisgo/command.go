package redisgo

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"time"
)

func (r *Redisgo) Close() error {
	return r.pool.Close()
}

func (r *Redisgo) TTL(ctx context.Context, key string) (res int64, err error) {
	return redis.Int64(r.Do(ctx, "TTL", key))
}

func (r *Redisgo) RPop(ctx context.Context, key string) (res string, err error) {
	reply, err := r.do(ctx, "RPOP", redisBytes, key)
	if err != nil {
		return "", err
	}
	return string(reply.([]byte)[:]), nil
}

func (r *Redisgo) LPush(ctx context.Context, name string, fields ...interface{}) error {
	keys := []interface{}{name}
	keys = append(keys, fields...)
	_, err := r.do(ctx, "LPUSH", nil, keys...)
	return err
}

func (r *Redisgo) Send(ctx context.Context, name string, fields ...interface{}) error {
	keys := []interface{}{name}
	keys = append(keys, fields...)
	_, err := r.do(ctx, "RPUSH", nil, keys...)
	return err
}

// DoCtx 函数与Do函数相比， 增加了一个context参数， 提供了超时的功能
func (r *Redisgo) DoCtx(ctx context.Context, cmd string, args ...interface{}) (interface{}, error) {
	var (
		ch    = make(chan struct{})
		reply interface{}
		err   error
	)
	go func() {
		defer close(ch)
		reply, err = r.do(ctx, cmd, nil, args...)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return reply, err
	}
}

// Set 返回两个参数，err不为空为服务器内服错误， 当命令执行成功时，ret为true
// 使用这个函数时需要同时判断这两个返回值
func (r *Redisgo) Set(ctx context.Context, key, value interface{}) (ret bool, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "SET", redisString, key, value)
	if err != nil {
		return
	}

	rsp := reply.(string)

	if rsp == "OK" {
		ret = true
	}

	return
}

//不存在则设置，存在则不设置
func (r *Redisgo) SetNX(ctx context.Context, key string, value interface{}) (num int, err error) {
	num, err = redis.Int(r.Do(ctx, "SETNX", key, value))
	return
}

func (r *Redisgo) SetNEX(ctx context.Context, key string, value interface{}, sec int) (bt bool, err error) {
	_, err = redis.String(r.Do(ctx, "SET", key, value, "EX", sec, "NX"))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Redisgo) SetExSecond(ctx context.Context, key, value interface{}, dur int) (ret string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "SET", redisString, key, value, "EX", dur)
	if err != nil {
		return
	}
	ret = reply.(string)
	return
}

func (r *Redisgo) Get(ctx context.Context, key string) (ret []byte, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "GET", redisBytes, key)
	if err != nil {
		return
	}
	ret = reply.([]byte)
	return
}

func (r *Redisgo) GetString(ctx context.Context, key string) (ret string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "GET", redisString, key)
	if err != nil {
		return
	}
	ret = reply.(string)
	return
}

func (r *Redisgo) GetFloat64(ctx context.Context, key string) (ret float64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "GET", redisFloat64, key)
	if err != nil {
		return
	}
	ret = reply.(float64)
	return
}

func (r *Redisgo) GetInt(ctx context.Context, key string) (ret int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "GET", redisInt, key)
	if err != nil {
		return
	}
	ret = reply.(int)
	return
}

func (r *Redisgo) GetInt64(ctx context.Context, key string) (ret int64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "GET", redisInt64, key)
	if err != nil {
		return
	}
	ret = reply.(int64)
	return
}

func (r *Redisgo) MGet(ctx context.Context, keys ...interface{}) (ret [][]byte, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "MGET", redisByteSlices, keys...)
	if err != nil {
		return
	}
	ret = reply.([][]byte)
	return
}

func (r *Redisgo) MSet(ctx context.Context, keys ...interface{}) (ret string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "MSET", redisString, keys...)
	if err != nil {
		return
	}
	ret = reply.(string)
	return
}

func (r *Redisgo) Del(ctx context.Context, args ...interface{}) (count int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "Del", redisInt, args...)
	if err != nil {
		return
	}
	count = reply.(int)
	return
}

func (r *Redisgo) Exists(ctx context.Context, key string) (res bool, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "Exists", redisBool, key)
	if err != nil {
		return
	}
	res = reply.(bool)
	return
}

func (r *Redisgo) Expire(ctx context.Context, key string, expire time.Duration) error {
	_, err := r.do(ctx, "EXPIRE", nil, key, int64(expire.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

/*
*	hash
 */
func (r *Redisgo) HDel(ctx context.Context, key interface{}, fields ...interface{}) (res int, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, fields...)

	reply, err = r.do(ctx, "HDEL", redisInt, keys...)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) HSet(ctx context.Context, key, fieldk string, fieldv interface{}) (res int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "HSET", redisInt, key, fieldk, fieldv)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) HGet(ctx context.Context, key, field string) (res string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "HGET", redisString, key, field)
	if err != nil {
		return
	}
	res = reply.(string)
	return
}

func (r *Redisgo) HGetFloat(ctx context.Context, key, field string) (res float64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "HGET", redisFloat64, key, field)
	if err != nil {
		return
	}
	res = reply.(float64)
	return
}

func (r *Redisgo) HGetBytes(ctx context.Context, key, field string) ([]byte, error) {
	return redis.Bytes(r.Do(ctx, "HGET", key, field))
}

func (r *Redisgo) HGetInt(ctx context.Context, key, field string) (res int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "HGET", redisInt, key, field)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) HGetUint64(ctx context.Context, key, field string) (res uint64, err error) {
	return redis.Uint64(r.Do(ctx, "HGET", key, field))
}

func (r *Redisgo) HExists(ctx context.Context, key, field string) (res bool, err error) {
	data, err := redis.Int(r.Do(ctx, "HEXISTS", key, field))
	if err != nil {
		return false, err
	}
	return data == 1, nil
}

func (r *Redisgo) HMGet(ctx context.Context, key string, fields ...interface{}) (res []string, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, fields...)
	reply, err = r.do(ctx, "HMGET", redisStrings, keys...)
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) HMSetStruct(ctx context.Context, key string, model interface{}) (string, error) {
	return redis.String(r.Do(ctx, "HMSET", redis.Args{}.Add(key).AddFlat(model)...))
}

func (r *Redisgo) HGetStruct(ctx context.Context, key string, model interface{}) error {
	value, err := redis.Values(r.Do(ctx, "HGETALL", key))
	if err != nil {
		return err
	}
	if len(value) == 0 {
		return ErrKeyNoExist
	}
	return redis.ScanStruct(value, model)
}

func (r *Redisgo) HGetStructSlice(ctx context.Context, key string, model interface{}) error {
	value, err := redis.Values(r.Do(ctx, "HGETALL", key))
	if err != nil {
		return err
	}
	if len(value) == 0 {
		return ErrKeyNoExist
	}
	return redis.ScanSlice(value, model)
}

func (r *Redisgo) HMSet(ctx context.Context, key string, fields ...interface{}) (res string, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, fields...)
	reply, err = r.do(ctx, "HMSET", redisString, keys...)
	if err != nil {
		return
	}
	res = reply.(string)
	return
}

func (r *Redisgo) HGetAll(ctx context.Context, key string) (res map[string]string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "HGETALL", redisStringMap, key)
	if err != nil {
		return
	}
	res = reply.(map[string]string)
	return
}

func (r *Redisgo) HKeys(ctx context.Context, key string) (res []string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "HKEYS", redisStrings, key)
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) HIncrby(ctx context.Context, key, field string, incr int) (res int64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "HINCRBY", redisInt64, key, field, incr)
	if err != nil {
		return
	}
	res = reply.(int64)
	return
}

func (r *Redisgo) HIncrbyFloat(ctx context.Context, key, field string, incr float64) (res float64, err error) {
	return redis.Float64(r.Do(ctx, "HINCRBYFLOAT", key, field, incr))
}

/*
*	set
 */
func (r *Redisgo) SAdd(ctx context.Context, key string, members ...interface{}) (res int, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, members...)
	reply, err = r.do(ctx, "SADD", redisInt, keys...)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) SRem(ctx context.Context, key string, members ...interface{}) (res int, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, members...)
	reply, err = r.do(ctx, "SREM", redisInt, keys...)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) SIsMember(ctx context.Context, key string, member string) (res bool, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "SISMEMBER", redisBool, key, member)
	if err != nil {
		return
	}
	res = reply.(bool)

	return
}

func (r *Redisgo) SCard(ctx context.Context, key string) (ret int64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "SCARD", redisInt64, key)
	if err != nil {
		return
	}
	ret = reply.(int64)
	return
}

func (r *Redisgo) SMembers(ctx context.Context, key string) (res []string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "SMEMBERS", redisStrings, key)
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) ZAdd(ctx context.Context, key string, args ...interface{}) (res int, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, args...)
	reply, err = r.do(ctx, "ZADD", redisInt, keys...)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) ZRange(ctx context.Context, key string, args ...interface{}) (res []string, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, args...)
	reply, err = r.do(ctx, "ZRANGE", redisStrings, keys...)
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) ZRangeInt(ctx context.Context, key string, start, stop int) (res []int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZRANGE", redisInts, key, start, stop)
	if err != nil {
		return
	}
	res = reply.([]int)
	return
}

func (r *Redisgo) ZRangeWithScore(ctx context.Context, key string, start, stop int) (res []string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZRANGE", redisStrings, key, start, stop, "WITHSCORES")
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) ZRevRangeWithScore(ctx context.Context, key string, start, stop int) (res []string, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZREVRANGE", redisStrings, key, start, stop, "WITHSCORES")
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) ZCount(ctx context.Context, key string, min, max int64) (res int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZCOUNT", redisInt, key, min, max)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) ZCard(ctx context.Context, key string) (res int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZCARD", redisInt, key)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) LLen(ctx context.Context, key string) (res int64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "LLEN", redisInt64, key)
	if err != nil {
		return
	}
	res = reply.(int64)
	return
}

func (r *Redisgo) Incrby(ctx context.Context, key string, incr int) (res int64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "INCRBY", redisInt64, key, incr)
	if err != nil {
		return
	}
	res = reply.(int64)
	return
}

func (r *Redisgo) Incr(ctx context.Context, key string) (res int64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "INCR", redisInt64, key)
	if err != nil {
		return
	}
	res = reply.(int64)
	return
}

func (r *Redisgo) ZIncrby(ctx context.Context, key string, incr int, member string) (res int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZINCRBY", redisInt, key, incr, member)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

/*
* If the member not in the zset or key not exits, ZRank will return ErrNil
 */
func (r *Redisgo) ZRank(ctx context.Context, key string, member string) (res int, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZRANK", redisInt, key, member)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

/*
* 如果key 或者 member 不存在则会返回 ErrNil
 */
func (r *Redisgo) ZRem(ctx context.Context, key string, members ...interface{}) (res int, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, members...)

	reply, err = r.do(ctx, "ZREM", redisInt, keys...)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

func (r *Redisgo) ZRemrangebyrank(ctx context.Context, key string, members ...interface{}) (res int, err error) {
	var reply interface{}
	keys := []interface{}{key}
	keys = append(keys, members...)

	reply, err = r.do(ctx, "ZREMRANGEBYRANK", redisInt, keys...)
	if err != nil {
		return
	}
	res = reply.(int)
	return
}

/*
* 如果key 或者 member不存在则返回 ErrNil
 */
func (r *Redisgo) ZScore(ctx context.Context, key, member string) (res float64, err error) {
	var reply interface{}
	reply, err = r.do(ctx, "ZSCORE", redisFloat64, key, member)
	if err != nil {
		return
	}
	res = reply.(float64)
	return
}

func (r *Redisgo) Zrevrange(ctx context.Context, key string, args ...interface{}) (res []string, err error) {
	var reply interface{}
	argss := []interface{}{key}
	argss = append(argss, args...)
	reply, err = r.do(ctx, "ZREVRANGE", redisStrings, argss...)
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) Zrevrangebyscore(ctx context.Context, key string, args ...interface{}) (res []string, err error) {
	var reply interface{}
	argss := []interface{}{key}
	argss = append(argss, args...)
	reply, err = r.do(ctx, "ZREVRANGEBYSCORE", redisStrings, argss...)
	if err != nil {
		return
	}
	res = reply.([]string)
	return
}

func (r *Redisgo) ZrevrangebyscoreInt(ctx context.Context, key string, args ...interface{}) (res []int, err error) {
	var reply interface{}
	argss := []interface{}{key}
	argss = append(argss, args...)
	reply, err = r.do(ctx, "ZREVRANGEBYSCORE", redisInts, argss...)
	if err != nil {
		return
	}
	res = reply.([]int)
	return
}
