package websocket

import (
	"context"
	"encoding/json"
	"github.com/aloeproject/toolbox/logger"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

func NewClient(ctx context.Context, log logger.ILogger, ser *Server, conn *ws.Conn, router *Router) *Client {
	id := uuid.New().String()
	return &Client{
		ctx:    ctx,
		id:     id,
		server: ser,
		conn:   conn,
		log:    log,
		router: router,
		//数据处理buf
		msg: make(chan interface{}, 100),
	}
}

type Client struct {
	ctx    context.Context
	id     string
	server *Server
	conn   *ws.Conn
	router *Router
	log    logger.ILogger
	msg    chan interface{}
}

func (c *Client) Start() {
	c.server.AddClient(c)
	go c.read()
	go c.write()
}

func (c *Client) GetId() string {
	return c.id
}

func (c *Client) SendMsg(msg interface{}) {
	c.msg <- msg
}

func (c *Client) Close() {
	//注销
	c.server.RemoveClient(c)
	c.conn.Close()
}

func (c *Client) read() {
	defer c.Close()

	for {
		_, body, err := c.conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseNormalClosure, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				c.log.Errorw(c.ctx, "read message error: %v", err)
			}
			return
		} else {
			req := ReqMessage{}
			err = json.Unmarshal(body, &req)
			if err != nil {
				c.log.Errorw(c.ctx, "read message Unmarshal error %v", err)
				continue
			}
			if req.PathName != "" {
				rep, err := c.router.Exec(c.ctx, &ReqMessage{
					MsgManger: c.server,
					ClientId:  c.id,
					PathName:  req.PathName,
					Body:      req.Body,
				})
				if err != nil {
					c.log.Errorw(c.ctx, "read message exec error:%v", err)
					continue
				}
				c.msg <- rep
			}
		}
	}
}

func (c *Client) write() {
	defer c.Close()
	for {
		select {
		case msg := <-c.msg:
			body, err := json.Marshal(msg)
			if err != nil {
				c.log.Errorw(c.ctx, "write message error:%v", err)
				continue
			}

			if err = c.conn.WriteMessage(ws.BinaryMessage, body); err != nil {
				c.log.Errorw(c.ctx, "write writeMessage error:%v", err)
			}
		}
	}
}
