package infra

import "github.com/thebravebyte/numeris/domain"

// User : Master struct model for user data to use in the application
func UserFromDB(user User) *domain.User {
	return &domain.User{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Password:    user.Password,
		PhoneNumber: user.PhoneNumber,
		
	}
}
