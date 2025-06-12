package gorillasocket

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MethodHandler func(message *Message, c *Client) error

type Server struct {
	sync.RWMutex

	Upgrader *websocket.Upgrader

	Handlers map[string]MethodHandler

	chats map[string]*ChatRoom
	conns map[string]*Client

	pingTime time.Duration
	pongTime time.Duration

	maxMessageSize int64
}

// from config
func NewServer() *Server {
	return &Server{
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == "http://localhost:8080"
			},
		},
		conns:          make(map[string]*Client),
		maxMessageSize: 512,
		pingTime:       9 * time.Second,
		pongTime:       10 * time.Second,
	}
}

func (s *Server) Serve(conn *websocket.Conn, clientID string, chatIDs []string) {
	c := s.addConnection(conn, clientID, chatIDs)
	defer s.removeConnection(c)
	done := make(chan bool)

	go func() {
		c.read()
		done <- true
	}()
	go func() {
		c.write()
		done <- true
	}()
	<-done
}

func (s *Server) removeConnection(c *Client) {
	s.Lock()
	ids := c.GetChatIDs()
	delete(s.conns, c.ClientID)
	s.Unlock()
	c.Close()

	for _, chatID := range ids {
		chatRoom := s.GetChat(chatID)
		if chatRoom != nil {
			chatRoom.RemoveParticipant(c.ClientID)
		}
	}

}

func (s *Server) addConnection(conn *websocket.Conn, clientID string, chatIDs []string) *Client {
	s.Lock()

	c := NewClient(clientID, chatIDs, s, conn)
	s.conns[clientID] = c
	s.Unlock()
	for _, v := range chatIDs {
		room := s.GetChat(v)
		room.AddParticipant(clientID, c)
	}
	return c
}

func (s *Server) routeEvent(message *Message, c *Client) error {
	if handler, ok := s.Handlers[string(message.Type)]; ok {
		err := handler(message, c)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("request is not exist")
}

func (s *Server) GetChat(chatID string) *ChatRoom {
	s.RLock()
	c, ok := s.chats[chatID]
	s.RUnlock()
	if ok {
		return c
	}
	s.Lock()
	defer s.Unlock()
	c = newChat(chatID)
	s.chats[chatID] = c
	go c.Run()
	return c
}

func (s *Server) GetClient(clientID string) *Client {
	s.RLock()
	defer s.RUnlock()
	return s.conns[clientID]
}
