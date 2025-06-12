package gorillasocket

import "encoding/json"

type MessageType string

type Message struct {
	Type    MessageType
	Payload []byte
}

func UnmarshalMessage(data []byte) (*Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(*m)

}
