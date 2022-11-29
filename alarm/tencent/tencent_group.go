package tencent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

/*
企业微信云组报警
*/

const groupAlarmUrl = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="

type options struct {
	key        string // 机器人key
	acceptList string //接受列表  @all
	timeout    time.Duration
}

type Option func(*options)

func WithKey(key string) Option {
	return func(o *options) {
		o.key = key
	}
}

func WithAcceptList(l string) Option {
	return func(o *options) {
		o.acceptList = l
	}
}

func WithTimeout(t time.Duration) Option {
	return func(o *options) {
		o.timeout = t
	}
}

func NewTencentAlarmGroup(opts ...Option) *TencentAlarmGroup {
	defaultOp := options{
		acceptList: "@all",
		timeout:    1000 * time.Millisecond, //默认超时时间
	}

	for _, o := range opts {
		o(&defaultOp)
	}

	return &TencentAlarmGroup{
		key:        defaultOp.key,
		acceptList: defaultOp.acceptList,
		timeout:    defaultOp.timeout,
	}
}

type TencentAlarmGroup struct {
	key        string // 机器人key
	acceptList string //接受列表  @all
	timeout    time.Duration
}

func (t *TencentAlarmGroup) SendText(ctx context.Context, content string) error {
	//设置一个默认超时时间
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	g, ctx := errgroup.WithContext(ctx)
	defer cancel()

	finish := make(chan struct{})

	g.Go(func() error {
		m := map[string]interface{}{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content": content,
				"mentioned_list": []string{
					t.acceptList,
				},
			},
		}
		mJson, _ := json.Marshal(m)
		contentReader := bytes.NewReader(mJson)
		url := fmt.Sprintf("%s%s", groupAlarmUrl, t.key)
		req, _ := http.NewRequest("POST", url, contentReader)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		_, err := client.Do(req)
		finish <- struct{}{}
		return err
	})

	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case <-finish:
		if err := g.Wait(); err != nil {
			return err
		}
	}

	return nil
}
