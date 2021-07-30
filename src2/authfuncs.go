package authserv2

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Checks if time since last logon is not too long
func (acc *Account) DurationOK(max_token_age time.Duration) bool {
	return time.Duration(time.Since(acc.Last_login)) < time.Duration(max_token_age)
}

// ret true and prolong acc if login, token and time ok
// if smth wrong return false and msg
func (m *Manager) ProlongAuth(login, token string) error {
	//fmt.Printf("prolongAuth(%v, %v...)\n", login, token[:10])
	err := m.IsAuthenticated(login, token)
	if err != nil {
		return err
	}
	err = m.UpdateLastLogin2Now(login)
	return err
}

// return false and message if failure
// return true and "" if success
func (m *Manager) LoginAccount(login, pass string) error {
	// check if there is an account with such a login
	acc, err := m.GetAccount(login)
	if err != nil {
		return err
	}
	// check if hash matches
	if !passIsCorrect(pass, acc.Pass_hash) {
		return ErrPasswordMissmatch
	}
	acc.NewToken(CONFIG.TOKEN_LENGTH)
	err = m.UpdateToken(login, acc.Current_token)
	if err != nil {
		return err
	}
	err = m.UpdateLastLogin2Now(login)
	if err != nil {
		return err
	}
	err = m.UpdateLoggedIn(login, true)
	return err
}

func (m *Manager) RegisterAccount(login, pass string) error {
	if !m.connection {
		return ErrDBoffline
	}
	// check if there is an account with such a login
	acc, err := m.GetAccount(login)
	fmt.Println(acc, err)
	if acc == nil {
		if err != ErrLoginNotFound {
			return ErrDBoffline
		}
	} else {
		return ErrLoginInUse
	}
	if !loginMatchesRequirements(login) {
		return ErrLoginRequirements
	}
	if !passMatchesRequirements(pass) {
		return ErrPassRequirements
	}
	acc = MakeAccount(login, pass, CONFIG.TOKEN_LENGTH)
	acc.NewToken(CONFIG.TOKEN_LENGTH)
	m.AddAccount(acc)
	err = m.UpdateToken(login, acc.Current_token)
	if err != nil {
		return err
	}
	err = m.UpdateLastLogin2Now(login)
	if err != nil {
		return err
	}
	err = m.UpdateLoggedIn(login, true)
	return err
}

// Check if:
//  - login exists,
//  - is logged in
//  - tokens match
//  - durationOK
func (m *Manager) IsAuthenticated(login, token string) error {
	// ErrUserNotLogedIn, ErrTokenMissmatch, ErrTokenExpiration
	acc, err := m.GetAccount(login)
	if err != nil {
		return err
	}
	// check if user is loged out
	if !acc.Logged_in {
		return ErrUserNotLogedIn
	}
	// check if tokens match
	if acc.Current_token != token {
		return ErrTokenMissmatch
	}
	// check if this account's last login wasnt to long ago
	if !acc.DurationOK(CONFIG.MAX_TOKEN_AGE) {
		err := m.UpdateLoggedIn(login, false)
		if err != nil {
			return err
		}
		return ErrTokenExpiration
	}
	return nil
}

func passIsCorrect(pass, pass_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(pass_hash), []byte(pass))
	return err == nil
}

func loginMatchesRequirements(login string) bool {
	return (2 < len(login)) && (len(login) < 50)
}
func passMatchesRequirements(pass string) bool {
	return (2 < len(pass)) && (len(pass) < 50)
}
