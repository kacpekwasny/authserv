package authserv

import (
	"time"
)

// Checks if time since last logon is not too long
func (acc *account) durationOK(max_token_age time.Duration) bool {
	return time.Duration(time.Since(acc.Last_login)) < time.Duration(max_token_age)
}

// ret true and prolong acc if login, token and time ok
// if smth wrong return false and msg
func (s *Server) prolongAuth(login, token string) (bool, string) {
	//fmt.Printf("prolongAuth(%v, %v...)\n", login, token[:10])
	succ, msg := s.isAuthenticated(login, token)
	if !succ {
		return false, msg
	}
	s.m.updateLastLogin2Now(login)
	return true, msgSUCCESS
}

// return false and message if failure
// return true and "" if success
func (s *Server) loginAccount(login, pass string) (bool, string, *account) {
	//fmt.Printf("loginAccount(%v, '*******')\n", login)

	// check if there is an account with such a login
	succ, msg, acc := s.m.getAcc(login)
	if !succ {
		return false, msg, nil
	}
	// check if hash matches
	if !passIsCorrect(pass, acc.Pass_hash) {
		return false, msgPASSWORD_MISSMATCH, nil
	}

	if acc.Logged_in {
		return true, msgSUCCESS, acc
	}

	acc.newToken(s.Cnf.TOKEN_LENGTH)
	succ, msg = s.m.updateTokenDB(login, acc.Current_token)
	if !succ {
		return false, msg, nil
	}
	succ, msg = s.m.updateLastLogin2Now(login)
	if !succ {
		return false, msg, nil
	}
	succ, msg = s.m.updateLoggedIn(login, true)
	return succ, msg, acc
}

// Check if:
//  - login exists,
//  - is logged in
//  - tokens match
//  - durationOK
func (s *Server) isAuthenticated(login, token string) (bool, string) {
	// return succ, is_authorised, msg
	//fmt.Printf("isAuthorised(%v, %v...)\n", login, token[:10])
	// check if there is an account with such login
	succ, msg, acc := s.m.getAcc(login)
	if !succ {
		return false, msg
	}
	// check if user is loged out
	if !acc.Logged_in {
		return false, msgACCOUNT_LOGGED_OUT
	}
	// check if tokens match
	if acc.Current_token != token {
		return false, msgTOKEN_MISSMATCH
	}
	// check if this account's last login wasnt to long ago
	if !acc.durationOK(s.Cnf.MAX_TOKEN_AGE) {
		succ, msg := s.m.updateLoggedIn(login, false)
		if !succ {
			return false, msg
		}
		return false, msgTOKEN_TIMEOUT
	}
	return true, msgSUCCESS
}
