package authserv2

import (
	"time"

	cmt "github.com/kacpekwasny/commontools"
)

type Account struct {
	Login         string    `json:"login"`
	Pass_hash     string    `json:"pass_hash"`
	Last_login    time.Time `json:"last_login"`
	Current_token string    `json:"token"`

	// To de-auth token when in the time span that it would
	// still be valid
	Logged_in bool `json:"logged_in"`
}

// set login
func (acc *Account) SetLogin(login string) {
	acc.Login = login
}

// set hash of password
func (acc *Account) SetPassHash(pass string) {
	acc.Pass_hash = MakeHash(pass)
}

func (acc *Account) UpdatePassHash(pass_hash string) {
	acc.Pass_hash = pass_hash
}

// Update last_login time
func (acc *Account) UpdateLogTime() {
	acc.Last_login = time.Now()
}

// Create new token
func (acc *Account) NewToken(length int) {
	acc.Current_token = cmt.RandString(length)
}

func MakeAccount(login, pass string, token_length int) *Account {

	acc := &Account{}
	acc.SetLogin(login)
	acc.SetPassHash(pass)
	acc.UpdateLogTime()
	acc.NewToken(token_length)

	return acc
}
