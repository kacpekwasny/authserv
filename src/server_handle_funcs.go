package authserv

import (
	"encoding/json"
	"net/http"

	cmt "github.com/kacpekwasny/commontools"
)

/////////////////////////////////////
/////// ~~ handle funcs ~~ //////////
// POST handle add account
func (s *Server) handleAddAccount(w http.ResponseWriter, r *http.Request) {
	/*
		INPUT
		w - has to have a json {
			'login': 'xxxx',
			'pass':  'yyyy'
		}

		OUTPUT
		respond with json:
		{
			'status':  'success'/'fail',
			'err_code': check error_codes.txt
			'msg':	   'error message',
		}
	*/
	// response is in json
	s.Cnf.Log1("\n – – – handle AddAccount")
	m := s.intro2handleFuncs(w, fncd_AddAccount, r, "login", "pass")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}

	succ, exists, msg := s.m.loginExistsDB(m["login"])
	if !succ {
		s.Cnf.Log2("loginExistsDB( %v ), failed. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_DB_ERR, msg, nil)
		return
	}

	if exists {
		// login is in use
		s.Cnf.Log2("login in use:  %v", m["login"])
		s.respond(w, msgFAIL, erc_LOGIN_IN_USE, "login is in use allready", nil)
		return
	}

	// check if login and pass match requirements
	login_ok, login_msg := s.Cnf.loginMatchesReq(m["login"])
	pass_ok, pass_msg := s.Cnf.passMatchesReq(m["pass"])
	if !(login_ok && pass_ok) {
		// if one of them is wrong
		s.Cnf.Log2("login, pass dont patch requirements. login_msg: %v, pass_msg: %v", login_msg, pass_msg)
		s.respond(w, msgFAIL, erc_CREDENTIALS_REQUIREMENTS_FAIL,
			"credentials dont match requirements. login: "+login_msg+";  pass: "+pass_msg, nil)
		return
	}

	acc := makeAccount(m["login"], m["pass"], s.Cnf.TOKEN_LENGTH)
	succ, msg = s.m.addAccount(acc)
	if !succ {
		s.Cnf.Log2("addAccount( %v ), failed. "+msg, acc)
		s.respond(w, msgFAIL, erc_CREATE_ACCOUNT_FAIL, "addAccount() failed. msg: "+msg, nil)
		return
	}
	s.Cnf.Log3("Successfully added new account: %v", acc)
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfully created account", nil)
}

// POST remove account from accountBuffer
func (s *Server) handleRemoveAccount(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle RemoveAccount")
	m := s.intro2handleFuncs(w, fncd_RemoveAccount, r, "login")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}

	ok, msg := s.m.removeAccount(m["login"])
	if !ok {
		s.Cnf.Log2("removeAccount( %v ), failed. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_REMOVE_ACCOUNT_FAIL, "failed to remove account. "+msg, nil)
		return
	}
	s.Cnf.Log3("Successfuly removed account: " + m["login"])
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly removed account", nil)
}

// POST handle request for login
func (s *Server) handleLoginAccount(w http.ResponseWriter, r *http.Request) {
	/*
		INPUT
		w - has json with login and password
		Find Login
		Check if pass matches
		newToken()
		uLogTime()
		login()

		OUTPUT
		respond with json:
		{
			'status':  'success'/'fail',
			'err_code': check error_codes.txt
			'msg':	   'error message',
			'token':	acc.token
		}

	*/
	s.Cnf.Log1("\n – – – handle LoginAccount")
	m := s.intro2handleFuncs(w, fncd_LoginAccount, r, "login", "pass")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	succ, msg, acc := s.loginAccount(m["login"], m["pass"])
	if !succ {
		s.Cnf.Log2("loginAccount( %v , ****** ), failed. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_REQUEST_AUTH_FAIL, "request auth failed. "+msg, nil)
		return
	}

	resp := map[string]string{}
	resp["status"] = msgSUCCESS
	resp["err_code"] = erc_SUCCESS
	resp["msg"] = "successfully loged in"
	resp["token"] = acc.Current_token
	cmt.Pc(json.NewEncoder(w).Encode(resp))
	s.Cnf.Log3("Successfuly logged in account: " + acc.Login)
}

// POST handle request to prolong token life
func (s *Server) handleIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	/*
		INPUT json: login, token

		OUTPUT
		respond with json:
		{
			'status':  'success'/'fail',
			'err_code': check error_codes.txt
			'msg':	   'error message',
		}

	*/
	s.Cnf.Log1("\n – – – handle Is Authenticated")
	m := s.intro2handleFuncs(w, fncd_IsAuthenticated, r, "login", "token")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.Cnf.Log2("isAuthenticated( %v, **** ), failed. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_NOT_AUTH_ACC, "account is not authenticated. "+msg, nil)
		return
	}

	s.Cnf.Log3("Successfuly authenticated account: %v", m["login"])
	s.respond(w, msgSUCCESS, erc_SUCCESS, "account is authenticated.", nil)
}

// POST handle request to prolong token life
func (s *Server) handleProlongAuth(w http.ResponseWriter, r *http.Request) {
	/*
		INPUT json: login, token

		OUTPUT
		respond with json:
		{
			'status':  'success'/'fail',
			'err_code': check error_codes.txt
			'msg':	   'error message',
		}

	*/
	s.Cnf.Log1("\n – – – handle ProlongAuth")
	m := s.intro2handleFuncs(w, fncd_ProlongAuth, r, "login", "token")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	ok, msg := s.prolongAuth(m["login"], m["token"])
	if !ok {
		s.Cnf.Log2("prolongAuth( %v, ***** ), failed. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_PROLONG_SESSION_FAIL, "failed to prolong session. "+msg, nil)
		return
	}

	s.Cnf.Log3("Successfuly prolonged auth, login: %v ", m["login"])
	s.respond(w, msgSUCCESS, erc_SUCCESS, "succesfully prolonged session.", nil)
}

// POST handle Logout request
func (s *Server) handleLogoutAccount(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle Logout Account")
	m := s.intro2handleFuncs(w, fncd_LogoutAccount, r, "login", "token")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.Cnf.Log2("isAuthenticated( %v, ****** ), false. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_PROLONG_SESSION_FAIL, "need to be authenticated to logout. "+msg, nil)
		return
	}

	succ, msg := s.m.updateLoggedIn(m["login"], false)
	if !succ {
		s.Cnf.Log2("updateLoggedIn( %v, false ), failed. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_DB_ERR, msg, nil)
		return
	}
	s.Cnf.Log3("Successfuly logged out: %v", m["login"])
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successful logout", nil)
}

////////////////////////////////
/// ~ CHANGE LOGIN / PASS ~~ ///
func (s *Server) handleChangeLogin(w http.ResponseWriter, r *http.Request) {
	/*
		required aditional json fields:
			login, token, new_login
	*/
	s.Cnf.Log1("\n – – – handle Change Login")
	m := s.intro2handleFuncs(w, fncd_ChangeLogin, r, "login", "token", "new_login")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}

	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.Cnf.Log2("login: '%v', token: xxxxxxx", m["login"])
		s.respond(w, msgFAIL, erc_UNAUTHENTICATED, "need to be authenticated to change login. "+msg, nil)
		return
	}
	succ, msg := s.m.updateLogin(m["login"], m["new_login"])
	if !succ {
		s.Cnf.Log2("updateLogin(...) unsuccessfull. msg: " + msg)
		s.respond(w, msgFAIL, erc_CHANGE_LOGIN_FAILED, msg, nil)
		return
	}
	s.Cnf.Log3("Successfuly changed login to: " + m["new_login"])
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly changed login to: "+m["new_login"], nil)
}

func (s *Server) handleChangePass(w http.ResponseWriter, r *http.Request) {
	/*
		required aditional json fields:
			login, token, new_pass
	*/
	s.Cnf.Log1("\n – – – handle Change Pass")
	m := s.intro2handleFuncs(w, fncd_ChangePass, r, "login", "token", "new_pass")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}

	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.Cnf.Log2("login: '%v', token: xxxxxxx", m["login"])
		s.respond(w, msgFAIL, erc_UNAUTHENTICATED, "need to be authenticated to change pass. "+msg, nil)
		return
	}

	succ, msg := s.m.updatePassHash(m["login"], makeHash(m["new_pass"]))
	if !succ {
		s.Cnf.Log2("updatePass(...) unsuccessfull. msg: " + msg)
		s.respond(w, msgFAIL, erc_CHANGE_PASS_FAILED, msg, nil)
		return
	}
	s.Cnf.Log3("Successfuly changed pass for user: " + m["login"])
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly changed pass", nil)
}
