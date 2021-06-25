package authserv

import (
	"time"

	cmt "github.com/kacpekwasny/commontools"
)

type account struct {
	Login         string    `json:"login"`
	Pass_hash     string    `json:"pass_hash"`
	Last_login    time.Time `json:"last_login"`
	Current_token string    `json:"current_token"`

	// To de-auth token when in the time span that it would
	// still be valid
	Logged_in bool `json:"logged_in"`
}

// set login
func (acc *account) setLogin(login string) {
	acc.Login = login
}

// set hash of password
func (acc *account) setPassHash(pass string) {
	acc.Pass_hash = makeHash(pass)
}

func (acc *account) updatePassHash(pass_hash string) {
	acc.Pass_hash = pass_hash
}

// Update last_login time
func (acc *account) uLogTime() {
	acc.Last_login = time.Now()
}

// Create new token
func (acc *account) newToken() {
	acc.Current_token = cmt.RandString(TOKEN_LENGTH)
}

func (acc *account) logout() {
	acc.Logged_in = false
}

func (acc *account) login() {
	acc.Logged_in = true
}

func makeAccount(login, pass string) *account {

	acc := &account{}
	acc.setLogin(login)
	acc.setPassHash(pass)
	acc.uLogTime()
	acc.newToken()

	return acc
}
