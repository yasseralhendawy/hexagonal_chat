package auth

type IAuthRepo interface {
	CreateNewUser(*User) error
	EditUser(*User) error
	GetUser(string) (*User, error)
	// GetActiveUser(string) (*User, error)
	// GetUserByID(string) (*User, error)
}

type Service struct {
	Storage IAuthRepo
}

func (s *Service) GetUser(email string, password string) (*User, error) {
	//get the user from the database
	user, err := s.Storage.GetUser(email)
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

// TODO: we are going to transfare the input to dto object later
func (s *Service) CreateNewUser(email string, password string) (*User, error) {
	user, err := NewUser(email, password)
	if err != nil {
		return nil, err
	}
	err = s.Storage.CreateNewUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
