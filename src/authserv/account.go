package authserv

import (
	"time"

	cmt "github.com/kacpekwasny/commontools"
)

type account struct {
	Id            string    `json:"id:`
	Login         string    `json:"login"`
	Pass_hash     string    `json:"pass_hash"`
	Last_login    time.Time `json:"last_login"`
	Current_token string    `json:"current_token"`

	// To de-auth token when in the time span that it would
	// still be valid
	Loged_in bool `json:"logged_in"`
}

// create random ID; DOESNT GUARANTEE UNIQUENESS!
func (acc *account) makeID() {
	acc.Id = cmt.RandString(ID_LENGTH)
}

// set login
func (acc *account) setLogin(login string) {
	acc.Login = login
}

// set hash of password
func (acc *account) setPassHash(pass string) {
	acc.Pass_hash = makeHash(pass)
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
	acc.Loged_in = false
}

func (acc *account) login() {
	acc.Loged_in = true
}

func makeAccount(login, pass string) *account {

	acc := &account{}
	// make unique ID
	for {
		acc.makeID()
		if idIsUnique(acc.Id) {
			break
		}
	}
	acc.setLogin(login)
	acc.setPassHash(pass)
	acc.uLogTime()
	acc.newToken()

	return acc
}
