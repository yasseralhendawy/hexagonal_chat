package cstorage

import (
	"context"

	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
)

type user_tables struct {
	Data        _User_data
	Person      _Person_cql
	Credentials []credential
}

func (u *user_tables) toEntity() *auth.User {
	if len(u.Credentials) == 0 {
		return nil
	}
	var methods []auth.LoginMethod
	for _, m := range u.Credentials {
		methods = append(methods, m.method())
	}
	return &auth.User{
		UserId:       u.Data.ID,
		Email:        u.Data.Email,
		Username:     u.Data.Username,
		MobileNumber: u.Data.Phone,
		HashPassword: u.Credentials[0].pass(),
		FirstName:    u.Person.Firstname,
		LastName:     u.Person.Lastname,
		LoginMethod:  methods,
	}
}

func NewUserTables(user *auth.User) *user_tables {
	tables := &user_tables{}
	for _, v := range user.LoginMethod {
		switch v {
		case auth.Email:
			tables.Credentials = append(tables.Credentials, &user_by_email{
				ID:       user.UserId,
				Email:    user.Email,
				Password: user.HashPassword,
			})
		// other login methods cases
		default:
		}
	}
	tables.Data = _User_data{
		ID:       user.UserId,
		Email:    user.Email,
		Username: user.Username,
		Phone:    user.MobileNumber,
	}
	tables.Person = _Person_cql{
		ID:        user.UserId,
		Username:  user.Username,
		Firstname: user.FirstName,
		Lastname:  user.LastName,
	}
	return tables
}

// table represent user info
type _Person_cql struct {
	ID        string `cql:"id"`
	Username  string `cql:"username"`
	Firstname string `cql:"firstname"`
	Lastname  string `cql:"lastname"`
}

func (user *_Person_cql) create(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.person_cql (id,username,firstname,lastname) VALUES (?, ?, ?, ?)`, user.ID, user.Username, user.Firstname, user.Lastname).WithContext(ctx).Exec()
}
func (user *_Person_cql) update(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`UPDATE `+instance.cfg.Keyspace+`.person_cql SET username=?, firstname=? ,lastname=? WHERE id=?`, user.Username, user.Firstname, user.Lastname, user.ID).WithContext(ctx).Exec()
}

func (user *_Person_cql) readOne(instance *CassandraDB, id string) error {
	ctx := context.Background()
	err := instance.session.Query(`SELECT id,username,firstname,lastname FROM `+instance.cfg.Keyspace+`.person_cql WHERE id = ? LIMIT 1`, id).WithContext(ctx).Scan(&user.ID, &user.Username, &user.Firstname, &user.Lastname)
	if err != nil {
		return err
	}
	return nil
}

// interface represent tables which represent the credential method
type credential interface {
	pass() string
	identifier() string
	method() auth.LoginMethod
	create(instance *CassandraDB) error
	update(instance *CassandraDB) error
	readOne(*CassandraDB, string) error
	check(*CassandraDB) (bool, error)
}

// the default credential methof table which represent the login by email
type user_by_email struct {
	Email    string `cql:"email"`
	ID       string `cql:"id"`
	Password string `cql:"pass"`
}

func (u *user_by_email) pass() string {
	return u.Password
}
func (u *user_by_email) identifier() string {
	return u.ID
}
func (u *user_by_email) method() auth.LoginMethod {
	return auth.Email
}

func (user *user_by_email) create(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.user_by_email (id,email,pass) VALUES (?, ?, ?)`, user.ID, user.Email, user.Password).WithContext(ctx).Exec()
}
func (user *user_by_email) update(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`UPDATE `+instance.cfg.Keyspace+`.user_by_email SET pass=? WHERE email=? AND id=?`, user.Password, user.Email, user.ID).WithContext(ctx).Exec()
}

func (user *user_by_email) readOne(instance *CassandraDB, email string) error {
	ctx := context.Background()
	err := instance.session.Query(`SELECT id,email,pass FROM `+instance.cfg.Keyspace+`.user_by_email WHERE email = ? LIMIT 1`, email).WithContext(ctx).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (user *user_by_email) check(instance *CassandraDB) (bool, error) {
	ctx := context.Background()
	scanner := instance.session.Query(`SELECT * FROM `+instance.cfg.Keyspace+`.user_by_email WHERE email = ?`, user.Email).WithContext(ctx).Iter().Scanner()
	return scanner.Next(), scanner.Err()
}

// table which represnt table sensitive info
type _User_data struct {
	ID       string `cql:"id"`
	Email    string `cql:"email"`
	Username string `cql:"username"`
	Phone    string `cql:"phone"`
}

func (user *_User_data) create(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.user_data (id,email,username,phone) VALUES (?, ?, ?, ?)`, user.ID, user.Email, user.Username, user.Phone).WithContext(ctx).Exec()

}
func (user *_User_data) update(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`UPDATE `+instance.cfg.Keyspace+`.user_data SET phone=?, username=? WHERE id=?`, user.Phone, user.Username, user.ID).WithContext(ctx).Exec()
}

func (user *_User_data) readOne(instance *CassandraDB, id string) error {
	ctx := context.Background()
	return instance.session.Query(`SELECT id,email,username,phone FROM `+instance.cfg.Keyspace+`.user_data WHERE id = ? LIMIT 1`, id).WithContext(ctx).Scan(&user.ID, &user.Email, &user.Username, &user.Phone)
}
