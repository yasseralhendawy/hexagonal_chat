package cstorage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yasseralhendawy/hexagonal_chat/config"
	cstorage "github.com/yasseralhendawy/hexagonal_chat/internal/infrastructure/cassandra"
)

var cfg = config.CassandraDB{
	UserName: "admin",
	Password: "admin",
	KeySpace: "test",
	Hosts:    []string{"localhost"},
}

func TestNewCassandraSession(t *testing.T) {
	assert := assert.New(t)

	session, err := cstorage.NewCassandraSession(cfg)
	assert.NoError(err)
	assert.NotEmpty(session)
}
