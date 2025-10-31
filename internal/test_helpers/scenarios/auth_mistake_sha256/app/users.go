package main

import (
	"crypto/sha1"
	"encoding/hex"
)

type User struct {
	passwords []string
	nopass    bool
}

type UserStore struct {
	users map[string]*User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*User),
	}
}

func (us *UserStore) GetUser(username string) (*User, bool) {
	user, exists := us.users[username]
	return user, exists
}

func (us *UserStore) CreateUser(username string) *User {
	user := &User{
		passwords: make([]string, 0),
		nopass:    true,
	}
	us.users[username] = user
	return user
}

func (us *UserStore) GetOrCreateUser(username string) *User {
	user, exists := us.users[username]
	if !exists {
		user = us.CreateUser(username)
	}
	return user
}

func (u *User) AddPassword(password string) {
	hashedPassword := hashPassword(password)
	u.passwords = append(u.passwords, hashedPassword)
	u.nopass = false
}

func (u *User) GetPasswords() []string {
	return u.passwords
}

func (u *User) HasNopass() bool {
	return u.nopass
}

func hashPassword(password string) string {
	hash := sha1.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}
