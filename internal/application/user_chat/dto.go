package userchat

type CreateChatReq struct {
	Participants []string
	text         string
}

type AddParticipantsRequest struct {
	ChatID       string
	Participants []string
}

type LeaveChatRequest struct {
	ChatID string
}
