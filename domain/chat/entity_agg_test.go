package chat_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
)

func TestEditMessage(t *testing.T) {
	var testCases = []struct {
		messageOptions []chat.MessageOpt
		epectedError   bool
	}{
		{[]chat.MessageOpt{}, true},
		{[]chat.MessageOpt{chat.MarkMessageAsDeleted()}, false},
		{[]chat.MessageOpt{chat.EditText("lol")}, false},
		{[]chat.MessageOpt{chat.EditText("lol"), chat.MarkMessageAsDeleted()}, false}, // a test case that does not make sense but it's ok lol
	}
	message := chat.NewMessage(uuid.NewString(), uuid.NewString(), "hellow world")
	assert := assert.New(t)
	for _, c := range testCases {
		oldMessage := *message
		message, err := message.EditMessage(c.messageOptions...)
		if c.epectedError {
			assert.Nil(message)
			assert.Error(err)
		} else {
			assert.NotEqual(oldMessage, message)
			assert.Equal(oldMessage.MessageID, message.MessageID)
			assert.Equal(oldMessage.ChatID, message.ChatID)
			assert.Nil(err)
		}
	}
}

func TestGetMessage(t *testing.T) {
	var testCases = []struct {
		chat        chat.Chat
		messageID   string
		expectError bool
	}{
		{chat.Chat{
			ChatID:          "id",
			ParticipantsIDs: []string{"id1,id2"},
			Messages: []*chat.Message{
				{
					MessageID:   "mid1",
					SenderID:    "senderID",
					ChatID:      "chatID",
					MessageText: "text",
				},
			},
			LastMessage: time.Now().Add(-1 * time.Minute),
		},
			"mid1", false,
		},
		{chat.Chat{
			ChatID:          "id",
			ParticipantsIDs: []string{"id1,id2"},
			Messages: []*chat.Message{
				{
					MessageID:   "mid1",
					SenderID:    "senderID",
					ChatID:      "chatID",
					MessageText: "text",
				},
			},
			LastMessage: time.Now().Add(-1 * time.Minute),
		},
			"id1", true,
		},
	}

	assert := assert.New(t)
	for _, c := range testCases {
		message, err := c.chat.GetMessage(c.messageID)
		if c.expectError {
			assert.Error(err)
			assert.Nil(message)
		} else {
			assert.Nil(err)
			assert.NotNil(message)
		}
	}
}

func TestEditChat(t *testing.T) {
	var testCases = []struct {
		chat        chat.Chat
		editOptions []chat.ChatOpt
		expectError bool
	}{
		//no options add
		{
			chat:        chatTemp,
			expectError: true,
		},
		//edit messages test cases
		{
			chat: chatTemp,
			editOptions: []chat.ChatOpt{
				chat.EditMessage(&chat.Message{
					MessageID:   "mid1",
					SenderID:    "senderID",
					ChatID:      "chatID", //chat id is different
					MessageText: "textmessage",
				}),
			},
			expectError: true,
		},
		{
			chat: chatTemp,
			editOptions: []chat.ChatOpt{
				chat.EditMessage(&chat.Message{
					MessageID:   "mid2",
					SenderID:    "senderID",
					ChatID:      "id",
					MessageText: "textmessage",
				}),
			},
			expectError: true,
		},
		{
			chat: chatTemp,
			editOptions: []chat.ChatOpt{
				chat.EditMessage(&chat.Message{
					MessageID:   "mid1",
					SenderID:    "senderID",
					ChatID:      "id",
					MessageText: "textmessage",
				}),
			},
			expectError: false,
		},
		// add message test cases
		{
			chat: chatTemp,
			editOptions: []chat.ChatOpt{
				chat.AddMessage(&chat.Message{
					MessageID:   "mid1",
					SenderID:    "senderID",
					ChatID:      "chatID", // the chat id is different
					MessageText: "textmessage",
				}),
			},
			expectError: true,
		},
		{
			chat: chatTemp,
			editOptions: []chat.ChatOpt{
				chat.AddMessage(&chat.Message{
					MessageID:   "mid2",
					SenderID:    "senderID",
					ChatID:      "id",
					MessageText: "textmessage",
				}),
				chat.AddMessage(&chat.Message{
					MessageID:   "mid2", //this message is already exist
					SenderID:    "senderID",
					ChatID:      "id",
					MessageText: "textmessage",
				}),
			},
			expectError: true,
		},
		{
			chat: chatTemp,
			editOptions: []chat.ChatOpt{
				chat.AddMessage(&chat.Message{
					MessageID:   "mid2",
					SenderID:    "senderID",
					ChatID:      "id",
					MessageText: "textmessage",
				}),
			},
			expectError: false,
		},
	}

	assert := assert.New(t)
	for _, c := range testCases {
		newchat, err := c.chat.EditChat(c.editOptions...)
		if c.expectError {
			assert.Error(err)
			assert.Nil(newchat)
		} else {
			assert.Nil(err)
			assert.Equal(c.chat.ChatID, newchat.ChatID)
			assert.NotEqual(c.chat, newchat)
		}
	}
}
