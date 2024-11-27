package cstorage

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type AuthRepo struct {
	instance *CassandraDB
	metric   appmetrics.Metrics
}

func (s *CassandraDB) NewAuthRepo(metric appmetrics.Metrics) (auth.IAuthRepo, error) {
	var wg sync.WaitGroup
	wg.Add(3)
	errCh := make(chan error, 3)
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
			errCh <- err
			return
		}
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	return AuthRepo{
		metric:   metric,
		instance: s,
	}, nil
}

// this function made only for testing purposes and
// because the domain will read the interface and it's not descriped in IAuthRepo it's fine to be here
func (a AuthRepo) DropTables() error {
	var wg sync.WaitGroup
	wg.Add(3)
	errCh := make(chan error, 3)
	go func() {
		defer wg.Done()
		err := a.instance.session.Query("DROP TABLE " + a.instance.cfg.Keyspace + ".user_by_email").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := a.instance.session.Query("DROP TABLE " + a.instance.cfg.Keyspace + ".person_cql").Exec()
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := a.instance.session.Query("DROP TABLE " + a.instance.cfg.Keyspace + ".user_data").Exec()
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

type user_tables struct {
	Data        user_data
	Person      person_cql
	Credentials []credential
}

type person_cql struct {
	ID        string `cql:"id"`
	Username  string `cql:"username"`
	Firstname string `cql:"firstname"`
	Lastname  string `cql:"lastname"`
}

type credential interface {
	pass() string
	identifier() string
	method() auth.LoginMethod
}

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

type user_data struct {
	ID       string `cql:"id"`
	Email    string `cql:"email"`
	Username string `cql:"username"`
	Phone    string `cql:"phone"`
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

// CreateNewUser implements auth.IAuthRepo.
func (a AuthRepo) CreateNewUser(user *auth.User) error {
	var wg sync.WaitGroup
	wg.Add(3)
	errCh := make(chan error, 3)

	go func() {
		defer wg.Done()
		ctx := context.Background()
		err := a.instance.session.Query(`INSERT INTO `+a.instance.cfg.Keyspace+`.user_data (id,email,username,phone) VALUES (?, ?, ?, ?)`, user.UserId, user.Email, user.Username, user.MobileNumber).WithContext(ctx).Exec()
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_data{}).String(), "Create", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_data{}).String(), "Create", "Success")
	}()

	go func() {
		defer wg.Done()
		ctx := context.Background()
		err := a.instance.session.Query(`INSERT INTO `+a.instance.cfg.Keyspace+`.person_cql (id,username,firstname,lastname) VALUES (?, ?, ?, ?)`, user.UserId, user.Username, user.FirstName, user.LastName).WithContext(ctx).Exec()
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(person_cql{}).String(), "Create", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(person_cql{}).String(), "Create", "Success")
	}()

	go func() {
		defer wg.Done()
		ctx := context.Background()
		err := a.instance.session.Query(`INSERT INTO `+a.instance.cfg.Keyspace+`.user_by_email (id,email,pass) VALUES (?, ?, ?)`, user.UserId, user.Email, user.HashPassword).WithContext(ctx).Exec()
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_by_email{}).String(), "Create", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_by_email{}).String(), "Create", "Success")
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

// EditUser implements auth.IAuthRepo.
func (a AuthRepo) EditUser(user *auth.User) error {
	if user == nil {
		return errors.New("user can not be nil")
	}
	userFound, err := a.CheckUserEmailExist(user.Email)
	if err != nil {
		return err
	}
	if !userFound {
		return errors.New("user can not be found")
	}
	var wg sync.WaitGroup
	wg.Add(3)
	errCh := make(chan error, 3)

	go func() {
		defer wg.Done()
		ctx := context.Background()
		err := a.instance.session.Query(`UPDATE `+a.instance.cfg.Keyspace+`.user_data SET phone=?, username=? WHERE id=?`, user.MobileNumber, user.Username, user.UserId).WithContext(ctx).Exec()
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_data{}).String(), "Update", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_data{}).String(), "Update", "Success")
	}()

	go func() {
		defer wg.Done()
		ctx := context.Background()
		err := a.instance.session.Query(`UPDATE `+a.instance.cfg.Keyspace+`.person_cql SET username=?, firstname=? ,lastname=? WHERE id=?`, user.Username, user.FirstName, user.LastName, user.UserId).WithContext(ctx).Exec()
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(person_cql{}).String(), "Update", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(person_cql{}).String(), "Update", "Success")
	}()

	go func() {
		defer wg.Done()
		ctx := context.Background()
		err := a.instance.session.Query(`UPDATE `+a.instance.cfg.Keyspace+`.user_by_email SET pass=? WHERE email=? AND id=?`, user.HashPassword, user.Email, user.UserId).WithContext(ctx).Exec()
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_by_email{}).String(), "Update", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_by_email{}).String(), "Update", "Success")
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

// GetUser implements auth.IAuthRepo.
func (a AuthRepo) GetUserByEmail(email string) (*auth.User, error) {
	c, err := a.getCredentialByEmail(email)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	dataCh := make(chan *user_data, 1)
	personCh := make(chan *person_cql, 1)
	errCh := make(chan error, 2)
	go func() {
		defer wg.Done()
		res, err := a.getUserData(c.ID)
		if err != nil {
			errCh <- err
			return
		}
		dataCh <- res
	}()
	go func() {
		defer wg.Done()
		res, err := a.getPersonalData(c.ID)
		if err != nil {
			errCh <- err
			return
		}
		personCh <- res
	}()

	wg.Wait()
	close(dataCh)
	close(personCh)
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	dao := &user_tables{
		Data:        *<-dataCh,
		Person:      *<-personCh,
		Credentials: []credential{c},
	}
	user := dao.toEntity()
	return user, nil
}

func (a AuthRepo) CheckUserEmailExist(email string) (bool, error) {
	ctx := context.Background()
	scanner := a.instance.session.Query(`SELECT * FROM `+a.instance.cfg.Keyspace+`.user_by_email WHERE email = ?`, email).WithContext(ctx).Iter().Scanner()
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user_by_email{}).String(), "Select", "")
	return scanner.Next(), scanner.Err()
}

func (a AuthRepo) getCredentialByEmail(email string) (*user_by_email, error) {
	user := &user_by_email{}
	ctx := context.Background()
	err := a.instance.session.Query(`SELECT id,email,pass FROM `+a.instance.cfg.Keyspace+`.user_by_email WHERE email = ? LIMIT 1`, email).WithContext(ctx).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Fail")
		return nil, err
	}
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Success")
	return user, nil
}

func (a AuthRepo) getPersonalData(id string) (*person_cql, error) {
	user, err := getPersonalData(a.instance, id)
	if err != nil {
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Fail")
		return nil, err
	}
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Success")
	return user, nil
}

func (a AuthRepo) getUserData(id string) (*user_data, error) {
	user, err := getUserData(a.instance, id)
	if err != nil {
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Fail")
		return nil, err
	}
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Success")
	return user, nil
}

func getPersonalData(instance *CassandraDB, id string) (*person_cql, error) {
	user := &person_cql{}
	ctx := context.Background()
	err := instance.session.Query(`SELECT id,username,firstname,lastname FROM `+instance.cfg.Keyspace+`.person_cql WHERE id = ? LIMIT 1`, id).WithContext(ctx).Scan(&user.ID, &user.Username, &user.Firstname, &user.Lastname)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func getUserData(instance *CassandraDB, id string) (*user_data, error) {
	user := &user_data{}
	ctx := context.Background()
	err := instance.session.Query(`SELECT id,email,username,phone FROM `+instance.cfg.Keyspace+`.user_data WHERE id = ? LIMIT 1`, id).WithContext(ctx).Scan(&user.ID, &user.Email, &user.Username, &user.Phone)
	if err != nil {
		return nil, err
	}
	return user, nil
}
