package cstorage

import (
	"fmt"
	"sync"

	"github.com/gocql/gocql"
	"github.com/yasseralhendawy/hexagonal_chat/config"
)

type CassandraDB struct {
	cfg     *gocql.ClusterConfig
	session *gocql.Session
}

func NewCassandraSession(cfg config.CassandraDB) (*CassandraDB, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cfg.UserName,
		Password: cfg.Password,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	err = session.Query(" CREATE KEYSPACE IF NOT EXISTS " + cfg.KeySpace + " WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };").Exec()
	if err != nil {
		return nil, err
	}
	cluster.Keyspace = cfg.KeySpace

	db := &CassandraDB{cfg: cluster, session: session}
	err = db.createTables()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *CassandraDB) StopSession() {
	s.session.Close()
}

func (s *CassandraDB) createTables() error {
	var wg sync.WaitGroup
	wg.Add(8)
	errCh := make(chan error, 8)
	go func() {
		defer wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".user_by_email (id text,email text,pass text,PRIMARY KEY(email,id));").Exec()
		if err != nil {
			errCh <- err
			return
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".person_cql (id text,username text,firstname text,lastname text,PRIMARY KEY (id));").Exec()
		if err != nil {
			errCh <- err
			return
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".user_data (id text,email text,phone text,username text,PRIMARY KEY (id));").Exec()
		if err != nil {
			fmt.Println("3")
			errCh <- err
			return
		}
	}()
	go func() {
		wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".chat (chat_id text,chat_name text,created_at timestamp,PRIMARY KEY(chat_id));").Exec()
		if err != nil {
			fmt.Println("4")
			errCh <- err
			return
		}
	}()
	go func() {
		wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".user_by_chat (chat_id text,user_id text,is_admin boolean,join_at timestamp,leave_at timestamp,PRIMARY KEY(chat_id,user_id));").Exec()
		if err != nil {
			errCh <- err
			return
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".chat_by_user(user_id text, chat_id text, chat_name text, is_fav boolean, join_at timestamp, leave_at timestamp, PRIMARY KEY(user_id, chat_id));").Exec()
		if err != nil {
			errCh <- err
			return
		}
	}()
	go func() {
		wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".message_by_chat(chat_id text, message_id text, user_id text, text text, sent_at timestamp, deleted_at timestamp, PRIMARY KEY(chat_id, sent_at, message_id)) WITH CLUSTERING ORDER BY (sent_at DESC, message_id ASC)").Exec()
		if err != nil {
			errCh <- err
			return
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("CREATE TABLE IF NOT EXISTS " + s.cfg.Keyspace + ".messages_by_user (user_id text, message_id text, chat_id text, text text, sent_at timestamp, deleted_at timestamp, PRIMARY KEY(user_id, message_id));").Exec()
		if err != nil {
			errCh <- err
			return
		}
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CassandraDB) DropTables() error {
	var wg sync.WaitGroup
	wg.Add(8)
	errCh := make(chan error, 8)
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".user_by_email").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".person_cql").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".user_data").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".chat").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".user_by_chat").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".chat_by_user").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".message_by_chat").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := s.session.Query("DROP TABLE " + s.cfg.Keyspace + ".messages_by_user").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
