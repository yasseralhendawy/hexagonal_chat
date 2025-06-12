package userchat

import "github.com/yasseralhendawy/hexagonal_chat/domain/user"

type App struct {
	Domain *user.Service
}

func (app App) AddParticipants(req *AddParticipantsRequest) (*user.UserChat, error) {
	return app.Domain.AddParticipants(req.ChatID, req.Participants)
}
func (app App) CreatNewChat(req *CreateChatReq, senderID string) (*user.UserChat, error) {
	return app.Domain.CreatNewChat(req.Participants, senderID, req.text)
}

func (app App) GetUserHistory(userID string) ([]*user.UserChat, error) {
	return app.Domain.GetUserHistory(userID)
}
func (app App) LeaveChat(userID string, req *LeaveChatRequest) (string, error) {
	err := app.Domain.LeaveChat(userID, req.ChatID)
	if err != nil {
		return "failed", err
	}
	return "sucess", nil
}
