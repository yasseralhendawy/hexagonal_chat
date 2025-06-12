package gorillasocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	sync.RWMutex

	Server *Server
	conn   *websocket.Conn

	ClientID  string
	ticker    *time.Ticker
	messageCh chan *[]byte

	chatList []string
}

func (c *Client) RemoveChat(d string) {
	c.Lock()
	defer c.Unlock()
	for i, chatID := range c.chatList {
		if chatID == d {
			c.chatList = append((c.chatList)[:i], (c.chatList)[i+1:]...)
			break
		}
	}
}

func (c *Client) AddChat(chatId string) {
	c.Lock()
	defer c.Unlock()
	c.chatList = append(c.chatList, chatId)
}

func (c *Client) GetChatIDs() []string {
	c.RLock()
	defer c.RUnlock()
	return c.chatList
}

func NewClient(id string, chatList []string, server *Server, conn *websocket.Conn) *Client {
	return &Client{
		ClientID:  id,
		Server:    server,
		conn:      conn,
		ticker:    time.NewTicker(server.pingTime),
		messageCh: make(chan *[]byte),
		chatList:  chatList,
	}
}

func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()
	c.ticker.Stop()
	c.conn.Close()
	close(c.messageCh)
}

func (c *Client) write() {
	for {
		select {
		case msg, ok := <-c.messageCh:
			if ok {
				err := c.conn.WriteMessage(websocket.TextMessage, *msg)
				if err != nil {
					//log here
					return
				}
			}
		case <-c.ticker.C:
			err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				//add log
				return
			}
		}
	}
}

func (c *Client) read() {
	c.conn.SetReadLimit(c.Server.maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.Server.pongTime))
	c.conn.SetPongHandler(c.pongHandler)
	for {
		_, msgData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//log here
			}
			break
		}
		msg, err := UnmarshalMessage(msgData)
		if err != nil {
			//log here
			break
		}
		fmt.Println("recived: ", msg.Type)

		err = c.Server.routeEvent(msg, c)
		if err != nil {
			//log here
			break
		}
		// c.server.Broadcast(msg)
	}
}
func (c *Client) pongHandler(msg string) error {
	return c.conn.SetReadDeadline(time.Now().Add(c.Server.pongTime))
}
