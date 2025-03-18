package cstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
)

type _MessageByChat struct {
	chatID    string
	messageID string
	userID    string
	sentAt    time.Time
	deletedAt *time.Time
	text      string
}

func new_MessageByChat(message *chat.Message) *_MessageByChat {
	return &_MessageByChat{
		chatID:    message.ChatID,
		userID:    message.SenderID,
		messageID: message.MessageID,
		text:      message.MessageText,
		sentAt:    message.TimeToPost,
		deletedAt: &message.DeltedAt,
	}
}

func (row *_MessageByChat) edit(message *chat.Message) *_MessageByChat {
	row.text = message.MessageText
	row.deletedAt = &message.DeltedAt
	return row
}

func (row *_MessageByChat) create(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.message_by_chat (chat_id,message_id,user_id,text,sent_at,deleted_at) VALUES (?, ?, ?, ?, ?, ?)`, row.chatID, row.messageID, row.userID, row.text, row.sentAt, row.deletedAt).WithContext(ctx).Exec()
}

func (row *_MessageByChat) update(instance *CassandraDB) error {
	fmt.Println(*row)
	ctx := context.Background()
	return instance.session.Query(`UPDATE `+instance.cfg.Keyspace+`.message_by_chat SET text = ?, deleted_at = ? WHERE message_id = ? AND chat_id = ? AND sent_at = ?`,
		row.text, row.deletedAt, row.messageID, row.chatID, row.sentAt).WithContext(ctx).Exec()
}

type _ListOfMessageByChat []_MessageByChat

func get_MessageByChat(instance *CassandraDB, messageID string, chatID string, sentAt time.Time) (*_MessageByChat, error) {
	var table _MessageByChat
	ctx := context.Background()
	err := instance.session.Query(`SELECT chat_id,message_id,user_id,text,sent_at,deleted_at FROM `+instance.cfg.Keyspace+`.message_by_chat WHERE  chat_id = ? AND message_id = ? AND sent_at = ? LIMIT 1`, chatID, messageID, sentAt).WithContext(ctx).Scan(&table.chatID, &table.messageID, &table.userID, &table.text, &table.sentAt, &table.deletedAt)
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (list *_ListOfMessageByChat) getOnlyLast(instance *CassandraDB, chatID string) error {
	var row _MessageByChat
	ctx := context.Background()
	err := instance.session.Query(`SELECT chat_id,message_id,user_id,text,sent_at,deleted_at FROM `+instance.cfg.Keyspace+`.message_by_chat WHERE chat_id=? ORDER BY sent_at DESC LIMIT 1`, chatID).WithContext(ctx).Scan(&row.chatID, &row.messageID, &row.userID, &row.text, &row.sentAt, &row.deletedAt)
	if err != nil {
		if err.Error() == "not found" {
			return nil
		}
		return err
	}
	*list = append(*list, row)
	return nil
}

func (list *_ListOfMessageByChat) readMany(instance *CassandraDB, chatID string) error {
	ctx := context.Background()
	scanner := instance.session.Query(`SELECT chat_id,message_id,user_id,text,sent_at,deleted_at FROM `+instance.cfg.Keyspace+`.message_by_chat WHERE chat_id=? ORDER BY sent_at DESC`, chatID).WithContext(ctx).Iter().Scanner()
	for scanner.Next() {
		var row _MessageByChat
		err := scanner.Scan(&row.chatID, &row.messageID, &row.userID, &row.text, &row.sentAt, &row.deletedAt)
		if err != nil {
			return err
		}
		*list = append(*list, row)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
