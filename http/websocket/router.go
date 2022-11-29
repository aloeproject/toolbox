package websocket

import (
	"context"
	"errors"
	"strings"
)

var ErrRouterRule = errors.New("路由规则错误")
var ErrRouterGroupNoExists = errors.New("路由组不存在")
var ErrRouterControllerNoExists = errors.New("路由名称不存在")

type IMessageManager interface {
	SendMsg(connId string, msg interface{})
	Broadcast(msg interface{})
}

type Message struct {
	Body interface{} `json:"body"`
}

/*
请求消息
*/
type ReqMessage struct {
	MsgManger IMessageManager `json:"-"`
	ClientId  string          `json:"-"`
	PathName  string          `json:"path_name"` //路由名称  路由名称格式  groupsName.controllerName
	Body      []byte          `json:"body"`
}

/*
响应消息
*/
type RepMessage struct {
	Data interface{} `json:"data"`
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
}

type HandlerFunc func(ctx context.Context, req *ReqMessage) (*RepMessage, error)

func NewGroup() *Group {
	return &Group{make(map[string]HandlerFunc)}
}

type Group struct {
	HandlerMap map[string]HandlerFunc
}

func (g *Group) AddRouter(name string, h HandlerFunc) {
	g.HandlerMap[name] = h
}

func NewRouter() *Router {
	return &Router{
		group: make(map[string]*Group),
	}
}

type Router struct {
	group map[string]*Group
}

func (r *Router) AddGroup(name string) *Group {
	if g, ok := r.group[name]; !ok {
		g = NewGroup()
		r.group[name] = g
		return g
	} else {
		return g
	}
}

func (r *Router) Exec(ctx context.Context, req *ReqMessage) (*RepMessage, error) {
	/*
		定义 account.login
		第一个为组标识  第二个为业务标识
	*/

	strArr := strings.Split(req.PathName, ".")

	if len(strArr) == 0 || len(strArr) != 2 {
		return nil, ErrRouterRule
	}

	prefix := strArr[0]
	name := strArr[1]
	if _, ok := r.group[prefix]; !ok {
		return nil, ErrRouterGroupNoExists
	}

	routerGroup := r.group[prefix]
	if _, ok := routerGroup.HandlerMap[name]; !ok {
		return nil, ErrRouterControllerNoExists
	}

	return routerGroup.HandlerMap[name](ctx, req)
}
