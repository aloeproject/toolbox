package websocket

import (
	"context"
	"crypto/tls"
	"github.com/aloeproject/toolbox/logger"
	ws "github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type options struct {
	address string
	timeout time.Duration
	log     logger.ILogger
	network string
	urlPath string //请求路径
	tlsConf *tls.Config
}

type Options func(o *options)

func WithTimeout(t time.Duration) Options {
	return func(o *options) {
		o.timeout = t
	}
}

func WithLogger(log logger.ILogger) Options {
	return func(o *options) {
		o.log = log
	}
}

func WithTLSConfig(c *tls.Config) Options {
	return func(o *options) {
		o.tlsConf = c
	}
}

func WithAddress(s string) Options {
	return func(o *options) {
		o.address = s
	}
}

var _ IMessageManager = (*Server)(nil)

type Server struct {
	httpServer *http.Server
	tlsConf    *tls.Config

	address string
	timeout time.Duration
	log     logger.ILogger
	clients map[string]*Client
	urlPath string //请求路径

	upgrader *ws.Upgrader
	router   *Router
	lock     *sync.RWMutex
}

func NewServer(router *Router, opts ...Options) *Server {
	conf := options{
		timeout: 3 * time.Second,
		log:     nil, //todo
		urlPath: "/",
		network: "tcp",
	}

	for _, o := range opts {
		o(&conf)
	}

	s := &Server{
		httpServer: &http.Server{
			Addr: conf.address,
		},
		address: conf.address,
		timeout: conf.timeout,
		log:     conf.log,
		clients: make(map[string]*Client),
		urlPath: conf.urlPath,
		upgrader: &ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,

			//允许跨域
			CheckOrigin: func(r *http.Request) bool {
				return true
			}},
		router: router,
		lock:   new(sync.RWMutex),
	}

	return s
}

/*
单播
*/
func (s *Server) SendMsg(connId string, msg interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if c, ok := s.clients[connId]; ok {
		c.SendMsg(msg)
	}
}

/*
广播
*/
func (s *Server) Broadcast(msg interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, c := range s.clients {
		c.SendMsg(msg)
	}
}

/*
客户端连接数量
*/
func (s *Server) ClientCount() int {
	return len(s.clients)
}

/*
程序开始
*/
func (s *Server) Start(ctx context.Context) error {
	http.HandleFunc(s.urlPath, s.wsHandle)
	var err error
	if s.address != "" {
		s.log.Infof("websocket server listening on: %s", s.address)
	} else {
		s.log.Infof("websocket server start")
	}
	if s.tlsConf != nil {
		err = s.httpServer.ListenAndServeTLS("", "")
	} else {
		err = s.httpServer.ListenAndServe()
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) StartByHandle(ctx context.Context, writer http.ResponseWriter, req *http.Request) {
	s.wsHandle(writer, req)
	select {
	case <-ctx.Done():
		return
	}
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) wsHandle(writer http.ResponseWriter, req *http.Request) {
	conn, err := s.upgrader.Upgrade(writer, req, nil)
	if err != nil {
		s.log.Errorf("upgrade exception:", err)
		return
	}

	client := NewClient(req.Context(), s.log, s, conn, s.router)
	client.Start()
}

/*
增加客户端
*/
func (s *Server) AddClient(c *Client) {
	s.lock.Lock()
	s.clients[c.GetId()] = c
	s.log.Infof("client connect id:%v current_counts:%d", c.GetId(), s.ClientCount())
	s.lock.Unlock()
}

/*
退出客户端
*/
func (s *Server) RemoveClient(c *Client) {
	if _, ok := s.clients[c.GetId()]; ok {
		s.lock.Lock()
		delete(s.clients, c.GetId())
		s.log.Infof("client disconnect id:%v current_counts:%d", c.GetId(), s.ClientCount())
		s.lock.Unlock()
	}
}
