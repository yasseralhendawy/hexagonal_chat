package cstorage

import (
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

func (s *CassandraDB) NewAuthRepo(metric appmetrics.Metrics) (*AuthRepo, error) {
	return &AuthRepo{
		metric:   metric,
		instance: s,
	}, nil
}

// func (a *AuthRepo) IAuthRepo() auth.IAuthRepo {
// 	return a
// }

// CreateNewUser implements auth.IAuthRepo.
func (a *AuthRepo) CreateNewUser(user *auth.User) error {
	user_tables := NewUserTables(user)
	n := len(user_tables.Credentials) + 2
	if n <= 2 {
		return errors.New("there is no credentials")
	}

	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan error, n)

	go func() {
		defer wg.Done()
		err := user_tables.Data.create(a.instance)
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(_User_data{}).String(), "Create", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(_User_data{}).String(), "Create", "Success")
	}()

	go func() {
		defer wg.Done()
		err := user_tables.Person.create(a.instance)
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(_Person_cql{}).String(), "Create", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(_Person_cql{}).String(), "Create", "Success")
	}()

	for _, credential := range user_tables.Credentials {
		go func() {
			defer wg.Done()
			err := credential.create(a.instance)
			if err != nil {
				a.metric.DBCallsWithLabelValues(reflect.TypeOf(credential).String(), "Create", "Fail")
				errCh <- err
				return
			}
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(credential).String(), "Create", "Success")
		}()
	}

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
func (a *AuthRepo) EditUser(user *auth.User) error {
	if user == nil {
		return errors.New("user can not be nil")
	}
	user_tables := NewUserTables(user)
	userFound, err := a.CheckUserExist(user_tables)
	if err != nil {
		return err
	}
	n := len(user_tables.Credentials) + 2
	if !userFound {
		return errors.New("user can not be found")
	}
	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan error, n)

	go func() {
		defer wg.Done()
		err := user_tables.Data.update(a.instance)
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(_User_data{}).String(), "Update", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(_User_data{}).String(), "Update", "Success")
	}()

	go func() {
		defer wg.Done()
		err := user_tables.Person.update(a.instance)
		if err != nil {
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(_Person_cql{}).String(), "Update", "Fail")
			errCh <- err
			return
		}
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(_Person_cql{}).String(), "Update", "Success")
	}()
	for _, credential := range user_tables.Credentials {
		go func() {
			defer wg.Done()
			err := credential.update(a.instance)
			if err != nil {
				a.metric.DBCallsWithLabelValues(reflect.TypeOf(credential).String(), "Update", "Fail")
				errCh <- err
				return
			}
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(credential).String(), "Update", "Success")
		}()
	}

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
func (a *AuthRepo) GetUserByEmail(email string) (*auth.User, error) {
	c, err := a.getCredentialByEmail(email)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	dataCh := make(chan *_User_data, 1)
	personCh := make(chan *_Person_cql, 1)
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
func (a *AuthRepo) CheckUserEmailExist(email string) (bool, error) {
	user := &user_by_email{Email: email}
	found, err := user.check(a.instance)
	var mValue string
	if found {
		mValue = "Found"
	} else {
		mValue = "Not Found"
	}
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "CheckExist", mValue)
	return found, err
}

func (a *AuthRepo) CheckUserExist(user *user_tables) (bool, error) {
	n := len(user.Credentials)
	var wg sync.WaitGroup
	wg.Add(n)
	foundCh := make(chan bool, n)
	errCh := make(chan error, n)
	for _, credential := range user.Credentials {
		go func() {
			defer wg.Done()
			check, err := credential.check(a.instance)
			if err != nil {
				errCh <- err
				return
			}
			var mValue string
			if check {
				mValue = "Found"
			} else {
				mValue = "Not Found"
			}
			a.metric.DBCallsWithLabelValues(reflect.TypeOf(credential).String(), "CheckExist", mValue)
			foundCh <- check
		}()

	}
	wg.Wait()
	close(errCh)
	close(foundCh)
	for err := range errCh {
		if err != nil {
			return true, err
		}
	}
	for f := range foundCh {
		if f {
			return f, nil
		}
	}
	return false, nil
}

func (a *AuthRepo) getCredentialByEmail(email string) (*user_by_email, error) {
	user := &user_by_email{}
	err := user.readOne(a.instance, email)
	if err != nil {
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Fail")
		return nil, err
	}
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Success")
	return user, nil
}

func (a *AuthRepo) getPersonalData(id string) (*_Person_cql, error) {
	user := &_Person_cql{}
	err := user.readOne(a.instance, id)
	if err != nil {
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Fail")
		return nil, err
	}
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Success")
	return user, nil
}

func (a *AuthRepo) getUserData(id string) (*_User_data, error) {
	user := &_User_data{}
	err := user.readOne(a.instance, id)
	if err != nil {
		a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Fail")
		return nil, err
	}
	a.metric.DBCallsWithLabelValues(reflect.TypeOf(user).String(), "Select", "Success")
	return user, nil
}

// func getPersonalData(instance *CassandraDB, id string) (*_Person_cql, error) {
// 	user := &_Person_cql{}
// 	ctx := context.Background()
// 	err := instance.session.Query(`SELECT id,username,firstname,lastname FROM `+instance.cfg.Keyspace+`.person_cql WHERE id = ? LIMIT 1`, id).WithContext(ctx).Scan(&user.ID, &user.Username, &user.Firstname, &user.Lastname)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return user, nil
// }

// func getUserData(instance *CassandraDB, id string) (*_User_data, error) {
// 	user := &_User_data{}
// 	ctx := context.Background()
// 	err := instance.session.Query(`SELECT id,email,username,phone FROM `+instance.cfg.Keyspace+`.user_data WHERE id = ? LIMIT 1`, id).WithContext(ctx).Scan(&user.ID, &user.Email, &user.Username, &user.Phone)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return user, nil
// }
