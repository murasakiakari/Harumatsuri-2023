package service

import (
	"fmt"
	"gatewayserver/backend"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	WRITE_WAIT       = 10 * time.Second
	PONG_WAIT        = 60 * time.Second
	PING_PERIOD      = PONG_WAIT * 9 / 10
	MAX_MESSAGE_SIZE = 512

	READ_BUFFER_SIZE  = 1024
	WRITE_BUFFER_SIZE = 1024
)

var (
	queuingSystem *QueuingSystem
	upgrader      = websocket.Upgrader{
		ReadBufferSize:  READ_BUFFER_SIZE,
		WriteBufferSize: WRITE_BUFFER_SIZE,
	}

	pattern = regexp.MustCompile(`^R?\d+$`)
)

func init() {
	queuingSystem = NewQueuingSystem()
}

type AllowEntryT struct {
	AllowEntry int `form:"allowEntry" binding:"required"`
}

type OffsetT struct {
	Offset float64 `form:"offset"`
}

type QueuingSystem struct {
	// Client related
	clients    map[*Client]*struct{}
	register   chan *Client
	unregister chan *Client
	sendAll    chan []byte

	// Queuing related
	queuingMessage []byte

	// General
	quit chan *struct{}
}

func NewQueuingSystem() *QueuingSystem {
	message, err := backend.GetMessage()
	if err != nil {
		fmt.Println(err)
	}

	queuingSystem := &QueuingSystem{
		clients:        make(map[*Client]*struct{}),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		sendAll:        make(chan []byte),
		queuingMessage: []byte(message),
		quit:           make(chan *struct{}),
	}
	go queuingSystem.clientRoutine()
	return queuingSystem
}

func (s *QueuingSystem) clientRoutine() {
Loop:
	for {
		select {
		case client := <-s.register:
			client.out <- s.queuingMessage
			s.clients[client] = nil
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.out)
			}
		case s.queuingMessage = <-s.sendAll:
			for client := range s.clients {
				go func(client *Client) {
					// select {
					// case client.out <- s.queuingMessage:
					// default:
					// 	s.unregister <- client
					// }
					client.out <- s.queuingMessage
				}(client)
			}
		case <-s.quit:
			for client := range s.clients {
				delete(s.clients, client)
				close(client.out)
			}
			break Loop
		}
	}
}

func (s *QueuingSystem) StartQueuing(ctx *gin.Context) {
	submit := &AllowEntryT{}
	if err := ctx.ShouldBind(submit); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind: " + err.Error()})
		return
	}

	message, err := backend.StartQueuing(submit.AllowEntry)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	queuingSystem.sendAll <- message
	ctx.JSON(http.StatusOK, nil)
}

func (s *QueuingSystem) AllowEntry(ctx *gin.Context) {
	submit := &AllowEntryT{}
	if err := ctx.ShouldBind(submit); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind: " + err.Error()})
		return
	}

	message, err := backend.AllowEntry(submit.AllowEntry)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	queuingSystem.sendAll <- message
	ctx.JSON(http.StatusOK, nil)
}

func (s *QueuingSystem) SetOffset(ctx *gin.Context) {
	submit := &OffsetT{}
	if err := ctx.ShouldBind(submit); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind: " + err.Error()})
		return
	}

	message, err := backend.SetOffset(submit.Offset)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	queuingSystem.sendAll <- message
	ctx.JSON(http.StatusOK, nil)
}

func (s *QueuingSystem) ConnectClient(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to upgrade to websocket: %w" + err.Error()})
		return
	}

	client := &Client{
		queuingSystem: s,
		conn:          conn,
		out:           make(chan []byte, 512),
	}
	go client.read()
	go client.write()
	s.register <- client
}

func (s *QueuingSystem) UpdateFirstEntry(peopleEntries int) error {
	message, err := backend.UpdateFirstEntry(peopleEntries)
	if err != nil {
		return err
	}

	s.sendAll <- message
	return nil
}

func (s *QueuingSystem) UpdateSecondEntry(peopleEntries int) error {
	message, err := backend.UpdateSecondEntry(peopleEntries)
	if err != nil {
		return err
	}

	s.sendAll <- message
	return nil
}

func (s *QueuingSystem) GetIndex(ctx *gin.Context) {
	number := ctx.Param("number")
	if !pattern.Match([]byte(number)) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid number"})
		return
	}

	order, err := backend.GetOrder(number)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"order": order})
}

type Client struct {
	queuingSystem *QueuingSystem
	conn          *websocket.Conn
	out           chan []byte
}

func (c *Client) read() {
	c.conn.SetReadLimit(MAX_MESSAGE_SIZE)
	c.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	c.queuingSystem.unregister <- c
	c.conn.Close()
}

func (c *Client) write() {
	ticker := time.NewTicker(PING_PERIOD)

Loop:
	for {
		select {
		case message, ok := <-c.out:
			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				break Loop
			}
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println(err)
				break Loop
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				break Loop
			}
		}
	}

	ticker.Stop()
}
