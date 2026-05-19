package users

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id                int    `db:"id"`
	Name              string `db:"name"`
	Email             string `db:"email"`
	EncryptedPassword string `db:"encrypted_password"`
}

func Register(ctx context.Context, name, email, password string) (*User, error) {
	u := User{Name: name, Email: email}
	u.SetPassword(password)
	if err := create(ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (u *User) SetPassword(p string) error {
	enc, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("set password: %w", err)
	}
	u.EncryptedPassword = string(enc)
	return nil
}

// VerifyPassword checks if p matches EncryptedPassword
// in a
func (u *User) VerifyPassword(p string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(p))
	if err == nil {
		return true
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false
	}
	panic(fmt.Errorf("password compare: %w", err))
}
