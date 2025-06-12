package chatmessage

type AddMessgaeRequest struct {
	ChatID string
	Text   string
}

type EditMessgaeRequest struct {
	ChatID    string
	MessageID string
	Text      string
}

type GetMessageRequest struct {
	ChatID string
}
