package cstorage

import (
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
	session.Close()
	cluster.Keyspace = cfg.KeySpace
	session, err = cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &CassandraDB{cfg: cluster, session: session}, nil
}

func (s *CassandraDB) StopSession() {
	s.session.Close()
}
