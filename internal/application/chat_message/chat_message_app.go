package chatmessage

import (
	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
)

type App struct {
	Domain *chat.ChatService
}

func (app App) AddMesage(senderID string, req AddMessgaeRequest) (*chat.Chat, error) {
	return app.Domain.AddMesage(req.ChatID, senderID, req.Text)
}
func (app App) EditMessageText(userID string, req EditMessgaeRequest) (*chat.Chat, error) {
	return app.Domain.EditMessageText(userID, req.ChatID, req.MessageID, req.Text)
}
func (app App) GetChat(req GetMessageRequest) (*chat.Chat, error) {
	return app.Domain.GetChat(req.ChatID)
}
