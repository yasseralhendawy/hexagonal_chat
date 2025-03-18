package cstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
)

type _MessageByUser struct {
	userID    string
	messageID string
	chatID    string
	text      string
	sentAt    time.Time
	deletedAt *time.Time
}

func new_MessageByUser(message *chat.Message) *_MessageByUser {
	return &_MessageByUser{
		userID:    message.SenderID,
		messageID: message.MessageID,
		chatID:    message.ChatID,
		text:      message.MessageText,
		sentAt:    message.TimeToPost,
		deletedAt: &message.DeltedAt,
	}
}

func (row *_MessageByUser) edit(message *chat.Message) *_MessageByUser {
	row.text = message.MessageText
	row.deletedAt = &message.DeltedAt
	return row
}

func (row *_MessageByUser) create(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.messages_by_user (user_id,chat_id,message_id,text,sent_at,deleted_at) VALUES (?, ?, ?, ?, ?, ?)`, row.userID, row.chatID, row.messageID, row.text, row.sentAt, row.deletedAt).WithContext(ctx).Exec()
}

func (row *_MessageByUser) update(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`UPDATE `+instance.cfg.Keyspace+`.messages_by_user SET text = ?, deleted_at = ? WHERE message_id = ? AND user_id = ?`,
		row.text, row.deletedAt, row.messageID, row.userID).WithContext(ctx).Exec()
}

func get_MessageByUser(instance *CassandraDB, messageID string, userID string) (*_MessageByUser, error) {
	var table _MessageByUser
	ctx := context.Background()
	err := instance.session.Query(`SELECT user_id,message_id,chat_id,text,sent_at,deleted_at FROM `+instance.cfg.Keyspace+`.messages_by_user WHERE  user_id = ? AND message_id = ? LIMIT 1`, userID, messageID).WithContext(ctx).Scan(&table.userID, &table.messageID, &table.chatID, &table.text, &table.sentAt, &table.deletedAt)
	if err != nil {
		fmt.Println("or here", err)
		return nil, err
	}
	return &table, nil
}

// type _ListOfMessageByUser []_MessageByUser

// func (list *_ListOfMessageByUser) readMany(instance *CassandraDB, userID string) error {
// 	ctx := context.Background()
// 	scanner := instance.session.Query(`SELECT chat_id,message_id,user_id,text,sent_at,deleted_at FROM `+instance.cfg.Keyspace+`.messages_by_user WHERE user_id=?`, userID).WithContext(ctx).Iter().Scanner()
// 	for scanner.Next() {
// 		var row _MessageByUser
// 		err := scanner.Scan(&row.chatID, &row.messageID, &row.userID, &row.text, &row.sentAt, &row.deletedAt)
// 		if err != nil {
// 			return err
// 		}
// 		*list = append(*list, row)
// 	}
// 	if err := scanner.Err(); err != nil {
// 		return err
// 	}
// 	return nil
// }
