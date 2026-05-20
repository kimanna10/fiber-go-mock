package patterns

import (
	"fiber-go/internal/errs"
	"fiber-go/internal/models"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// UserUpdateBuilder помогает строить обновленного юзера на основе патч структуры
type UserUpdateBuilder struct {
	user models.User
}

func NewUserUpdateBuilder(u models.User) *UserUpdateBuilder {
	return &UserUpdateBuilder{user: u}
}

func (b *UserUpdateBuilder) ApplyPatch(req models.UserUpdateRequest) (*UserUpdateBuilder, error) {
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, errs.ErrFieldEmpty
		}
		b.user.Name = *req.Name
	}

	if req.Age != nil {
		if *req.Age < 0 {
			return nil, errs.ErrInvalidData
		}
		b.user.Age = *req.Age
	}

	if req.Email != nil {
		if strings.TrimSpace(*req.Email) == "" {
			return nil, errs.ErrFieldEmpty
		}
		b.user.Email = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		b.user.Password = string(hashedPassword)
	}
	return b, nil
}

func (b *UserUpdateBuilder) Build() (models.User, error) {
	if strings.TrimSpace(b.user.Email) == "" {
		return models.User{}, errs.ErrFieldEmpty
	}
	return b.user, nil
}

// UserBuilder помогает строить нового юзера на основе запроса
type UserBuilder struct {
	user models.User
	err  error
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{}
}

func (b *UserBuilder) SetName(name string) *UserBuilder {
	if b.err != nil {
		return b
	}
	if strings.TrimSpace(name) == "" {
		b.err = errs.ErrFieldEmpty
		return b
	}
	b.user.Name = name
	return b
}

func (b *UserBuilder) SetAge(age int) *UserBuilder {
	if b.err != nil {
		return b
	}
	if age < 0 {
		b.err = errs.ErrInvalidData
		return b
	}
	b.user.Age = age
	return b
}

func (b *UserBuilder) SetEmail(email string) *UserBuilder {
	if b.err != nil {
		return b
	}
	if strings.TrimSpace(email) == "" {
		b.err = errs.ErrFieldEmpty
		return b
	}
	b.user.Email = email
	return b
}

func (b *UserBuilder) SetPassword(password string) *UserBuilder {
	if b.err != nil {
		return b
	}
	if strings.TrimSpace(password) == "" {
		b.err = errs.ErrFieldEmpty
		return b
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		b.err = err
		return b
	}
	b.user.Password = string(hashedPassword)
	return b
}

func (b *UserBuilder) DefaultRole() *UserBuilder {
	b.user.Role = "user"
	return b
}

func (b *UserBuilder) Build() (models.User, error) {
	if b.err != nil {
		return models.User{}, b.err
	}
	return b.user, nil
}
