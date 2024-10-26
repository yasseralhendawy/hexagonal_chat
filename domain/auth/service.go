package auth

import "errors"

type IAuthRepo interface {
	CreateNewUser(*User) error
	EditUser(*User) error
	GetUserByEmail(string) (*User, error)
	CheckUserEmailExist(string) (bool, error)
	// GetActiveUser(string) (*User, error)
	// GetUserByID(string) (*User, error)
}

type Service struct {
	storage IAuthRepo
}

func New(storage IAuthRepo) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) GetUser(email string, password string) (*User, error) {
	//get the user from the database
	user, err := s.storage.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	//check the password match
	err = user.checkPasswordMatch(password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) CreateNewUser(email string, password string) (*User, error) {
	userExist, err := s.storage.CheckUserEmailExist(email)
	if err != nil {
		return nil, err
	}
	if userExist {
		return nil, errors.New(email + " is already exist")
	}
	user, err := NewUser(email, password)
	if err != nil {
		return nil, err
	}
	err = s.storage.CreateNewUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
