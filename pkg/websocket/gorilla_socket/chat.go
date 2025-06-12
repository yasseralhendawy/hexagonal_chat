package gorillasocket

import (
	"sync"
)

type ChatRoom struct {
	sync.RWMutex
	ID           string
	Participants map[string]*Client
	Broadcast    chan *[]byte
}

func (c *ChatRoom) Run() {
	for msg := range c.Broadcast {
		c.RLock()
		for _, cl := range c.Participants {
			if cl == nil {
				continue
			}
			select {
			case cl.messageCh <- msg:
			default:
				cl.Close()
			}
		}
		c.RUnlock()
	}
}

func newChat(id string) *ChatRoom {
	return &ChatRoom{
		ID:           id,
		Participants: make(map[string]*Client),
		Broadcast:    make(chan *[]byte),
	}
}

// even if the Client is nil we will set it to the map till the client establish
func (c *ChatRoom) AddParticipant(clientID string, conn *Client) {
	if conn == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	c.Participants[clientID] = conn
}
func (c *ChatRoom) RemoveParticipant(participant string) {
	c.Lock()
	defer c.Unlock()
	if conn := c.Participants[participant]; conn != nil {
		conn.RemoveChat(c.ID)
	}
	delete(c.Participants, participant)
}
