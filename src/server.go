package authserv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cmt "github.com/kacpekwasny/commontools"
)

type Server struct {
	m        *Manager
	myr      *mux.Router
	URL_ROOT string
}

func InitServer(URL_ROOT, db_username, db_password, db_address string, dp_port int,
	db_database_name, db_tablename string, ping_interval time.Duration, max_accountBuffer_len int,
	observe_interval time.Duration) *Server {

	s := &Server{}

	s.m = InitManager(db_username, db_password, db_address, dp_port,
		db_database_name, db_tablename, ping_interval, max_accountBuffer_len,
		observe_interval)

	s.myr = &mux.Router{}
	s.setURL_ROOT(URL_ROOT)
	return s
}

func (s *Server) setURL_ROOT(root string) {
	l := len(root)
	if l == 0 {
		s.URL_ROOT = "/"
		return
	}
	if root[l-1] != '/' {
		root = root + "/"
		s.URL_ROOT = root
		return
	}
	s.URL_ROOT = root
}

func (s *Server) StopTestDBConn() {
	s.m.StopTestDBConn()
}

// json from Request to map[string]string
func json2map(r *http.Request) (map[string]string, error) {
	// read data that will be used to create an account
	reqBody, _ := ioutil.ReadAll(r.Body)
	m := &map[string]string{}
	err := json.Unmarshal(reqBody, m)
	return *m, err
}

// Auth client entity (NOT USER/ACOUNT), return true for AUTHORISED and false for UNAUTH.
func (s *Server) authClient(w http.ResponseWriter, m map[string]string) (bool, string) {
	ok := checkKeys(m, "client_id", "client_password")
	if !ok {
		s.respond(w, msgFAIL, erc_JSON_KEYS_MISSING, "client_id or password missing from json")
		return false, "client_id or password missing from json"
	}
	client_id := m["client_id"]
	client_pass := m["client_password"]

	// get good password assigned to client_id
	good_password, id_exists := CLIENT_PASSWORD[client_id]
	if !id_exists {
		// id not found in client database (not accounts DB)
		s.respond(w, msgFAIL, erc_CLIENT_ID_NOT_FOUND, "client failed to authorise! client login not found, (credentials to authorise request to this server)")
		return false, "client failed to authorise! client login not found, (credentials to authorise request to this server)"
	}
	if good_password != client_pass {
		s.respond(w, msgFAIL, erc_CLIENT_PASSWORD_MISSMATCH, "client failed to authorise! client password doesnt match")
		return false, "client failed to authorise! client password doesnt match"
	}

	return true, ""
}

// Shortcut for usual response
func (s *Server) respond(w http.ResponseWriter, status, err_code, msg string) {
	w.Header().Add("Content-Type", "application/json")
	resp := map[string]string{
		"status":   status,
		"err_code": err_code,
		"msg":      msg,
	}
	cmt.Pc(json.NewEncoder(w).Encode(resp))
	fmt.Println("response sent")
}

// Check client auth, check if keys are present, json2map, handle errors and respond, return nil if error and map on success
// Will respond and returns nil if something is wrong.
//  If returns map:
//  - json2map was successful,
//  - Client is authorised,
//  - All keys are present
func (s *Server) intro2handleFuncs(w http.ResponseWriter, r *http.Request, keys ...string) map[string]string {
	w.Header().Add("Content-Type", "application/json")
	m, err := json2map(r)
	// read data that will be used to create an account
	// Unmarshall error
	if err != nil {
		s.respond(w, msgFAIL, erc_UNMARSHAL_ERR, "json unmarshal error")
		fmt.Println("  intro: JSON UNMARSHAL ERROR")
		return nil
	}

	// auth client
	auth, msg := s.authClient(w, m)
	// Client not authorised
	if !auth {
		// authclient responds by itself
		fmt.Println("  intro: UNSUCCESSFUL CLIENT AUTHORISATION: ", msg)
		return nil
	}

	ok := checkKeys(m, keys...)
	if !ok {
		s.respond(w, msgFAIL, erc_JSON_KEYS_MISSING, fmt.Sprintf("required keys: %v", keys))
		fmt.Printf("  intro: MISSING KEYS. required keys: %v\n", keys)
		return nil
	}
	return m
}

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
	fmt.Println("\n ### Add Account Action ###")
	m := s.intro2handleFuncs(w, r, "login", "pass")
	if m == nil {
		return
	}

	succ, exists := s.m.loginExistsDB(m["login"])
	if !succ {
		s.respond(w, msgFAIL, erc_DB_ERR, msgERR_DB)
		return
	}

	if exists {
		// login is in use
		s.respond(w, msgFAIL, erc_LOGIN_IN_USE, "login is in use allready")
		return
	}

	// check if login and pass match requirements
	login_ok, login_msg := loginMatchesReq(m["login"])
	pass_ok, pass_msg := passMatchesReq(m["pass"])
	if !(login_ok && pass_ok) {
		// if one of them is wrong
		s.respond(w, msgFAIL, erc_CREDENTIALS_REQUIREMENTS_FAIL,
			"credentials dont match requirements. login: "+login_msg+";  pass: "+pass_msg)
		return
	}

	acc := makeAccount(m["login"], m["pass"])
	succ, msg := s.m.addAccount(acc)
	if !succ {
		fmt.Println("addAccountFailed(): ", msg)
		s.respond(w, msgFAIL, erc_CREATE_ACCOUNT_FAIL, "addAccount() failed. msg: "+msg)
		return
	}
	fmt.Println("Successfully added new account. Number of accounts in buffer: ", len(*s.m.accountBuffer))
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfully created account")
}

// POST remove account from accountBuffer
func (s *Server) handleRemoveAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n ### Remove Account Action ###")
	m := s.intro2handleFuncs(w, r, "login")
	if m == nil {
		return
	}

	ok, msg := s.m.removeAccount(m["login"])
	if !ok {
		s.respond(w, msgFAIL, erc_REMOVE_ACCOUNT_FAIL, "failed to remove account. "+msg)
		return
	}
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly removed account")
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
	fmt.Println("\n ### Request Auth Action ###")
	m := s.intro2handleFuncs(w, r, "login", "pass")
	if m == nil {
		return
	}
	succ, msg, acc := s.loginAccount(m["login"], m["pass"])
	if !succ {
		s.respond(w, msgFAIL, erc_REQUEST_AUTH_FAIL, "request auth failed. "+msg)
		return
	}

	resp := map[string]string{}
	resp["status"] = msgSUCCESS
	resp["err_code"] = erc_SUCCESS
	resp["msg"] = "successfully loged in"
	resp["token"] = acc.Current_token
	json.NewEncoder(w).Encode(resp)
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
	fmt.Println("\n ### prolong Auth Action ###")
	m := s.intro2handleFuncs(w, r, "login", "token")
	if m == nil {
		return
	}
	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.respond(w, msgFAIL, erc_NOT_AUTH_ACC, "account is not authenticated. "+msg)
		return
	}

	s.respond(w, msgSUCCESS, erc_SUCCESS, "account is authenticated.")
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
	fmt.Println("\n ### prolong Auth Action ###")
	m := s.intro2handleFuncs(w, r, "login", "token")
	if m == nil {
		return
	}
	ok, msg := s.prolongAuth(m["login"], m["token"])
	if !ok {
		s.respond(w, msgFAIL, erc_PROLONG_SESSION_FAIL, "failed to prolong session. "+msg)
		return
	}

	s.respond(w, msgSUCCESS, erc_SUCCESS, "succesfully prolonged session.")
}

// POST handle Logout request
func (s *Server) handleLogoutAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n ### Logout Action ###")
	m := s.intro2handleFuncs(w, r, "login", "token")
	if m == nil {
		return
	}
	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.respond(w, msgFAIL, erc_PROLONG_SESSION_FAIL, "need to be authenticated to logout. "+msg)
		return
	}

	succ, msg := s.m.updateLoggedIn(m["login"], false)
	if !succ {
		s.respond(w, msgFAIL, erc_DB_ERR, msg)
		return
	}
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successful logout")
}

////////////////////////////////
/// ~ CHANGE LOGIN / PASS ~~ ///
func (s *Server) handleChangeLogin(w http.ResponseWriter, r *http.Request) {
	/*
		required aditional json fields:
			login, token, new_login
	*/

	fmt.Println("\n ### Change Login Action ###")
	m := s.intro2handleFuncs(w, r, "login", "token", "new_login")
	if m == nil {
		return
	}

	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.respond(w, msgFAIL, erc_UNAUTHENTICATED, "need to be authenticated to change login. "+msg)
		return
	}
	succ, msg := s.m.updateLogin(m["login"], m["new_login"])
	if !succ {
		s.respond(w, msgFAIL, erc_CHANGE_LOGIN_FAILED, msg)
		return
	}
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly changed login to: "+m["new_login"])
}

func (s *Server) handleChangePass(w http.ResponseWriter, r *http.Request) {
	/*
		required aditional json fields:
			login, token, new_pass
	*/
	fmt.Println("\n ### Change Login Action ###")
	m := s.intro2handleFuncs(w, r, "login", "token", "new_pass")
	if m == nil {
		return
	}

	ok, msg := s.isAuthenticated(m["login"], m["token"])
	if !ok {
		s.respond(w, msgFAIL, erc_UNAUTHENTICATED, "need to be authenticated to change pass. "+msg)
		return
	}

	succ, msg := s.m.updatePassHash(m["login"], makeHash(m["new_pass"]))
	if !succ {
		s.respond(w, msgFAIL, erc_CHANGE_PASS_FAILED, msg)
		return
	}
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly changed pass")
}

////////////////////////////////
/////// ADMIN FUNCS GET ////////
// GET handle get one account //
func (s *Server) handleGet1Acc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n ### Get Account Action ###")
	m := s.intro2handleFuncs(w, r, "login")
	if m == nil {
		return
	}
	succ, msg, acc := s.m.getAcc(m["login"])
	if !succ {
		s.respond(w, msgFAIL, erc_GET_ACCOUNT_FAILED, msg)
		return
	}
	json.NewEncoder(w).Encode(*acc)
}

func (s *Server) handleGetAllLoginsDB(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n ### Get All Logins Action ###")
	m := s.intro2handleFuncs(w, r)
	if m == nil {
		return
	}

	succ, msg, logins_ls := s.m.getAllLogins()
	if !succ {
		s.respond(w, msgFAIL, erc_GET_ALL_LOGINS_ERR, msg)
		return
	}
	json.NewEncoder(w).Encode(logins_ls)
}

func (s *Server) handleGetAllAccountsDB(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n ### Get All Accounts Action ###")
	m := s.intro2handleFuncs(w, r)
	if m == nil {
		return
	}

	succ, msg, logins_ls := s.m.getAllAccounts()
	if !succ {
		s.respond(w, msgFAIL, erc_GET_ALL_ACCOUNTS_ERR, msg)
		return
	}
	json.NewEncoder(w).Encode(logins_ls)
}

func (s *Server) handleGetAllAccountsBuff(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n ### Get All Accounts from Buffer Action ###")
	m := s.intro2handleFuncs(w, r)
	if m == nil {
		return
	}

	var acc_ls []account
	for _, acc := range *s.m.accountBuffer {
		acc_ls = append(acc_ls, *acc)
	}

	json.NewEncoder(w).Encode(acc_ls)
}

func (s *Server) handleDELETE_ALL_RECORDS_FROM_DATABASE(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n ~~~~~~ DELETE ALL RECORDS FROM DATABASE ~~~~~~")
	m := s.intro2handleFuncs(w, r)
	if m == nil {
		return
	}
	succ := s.m.DELETE_ALL_RECORDS_IN_DATABASE()
	if !succ {
		s.respond(w, msgFAIL, erc_DELETE_ALL_ACCOUNTS_ERR, "fail during delete all accounts")
	}
	s.respond(w, msgSUCCESS, erc_SUCCESS, "succesfuly deleted all records")
}

//////////////////////////////////////

func (s *Server) ListenAndServe() {

	// client_id and password are to authenticate entity trying to create an account
	s.myr.HandleFunc(s.URL_ROOT+"addAccount", s.handleAddAccount).Methods("POST")
	s.myr.HandleFunc(s.URL_ROOT+"removeAccount", s.handleRemoveAccount).Methods("POST")
	s.myr.HandleFunc(s.URL_ROOT+"loginAccount", s.handleLoginAccount).Methods("POST")
	s.myr.HandleFunc(s.URL_ROOT+"prolongAuth", s.handleProlongAuth).Methods("POST")
	s.myr.HandleFunc(s.URL_ROOT+"isAuthenticated", s.handleIsAuthenticated).Methods("POST")
	s.myr.HandleFunc(s.URL_ROOT+"logoutAccount", s.handleLogoutAccount).Methods("POST")

	s.myr.HandleFunc(s.URL_ROOT+"changeLogin", s.handleChangeLogin).Methods("POST")
	s.myr.HandleFunc(s.URL_ROOT+"changePass", s.handleChangePass).Methods("POST")

	s.myr.HandleFunc(s.URL_ROOT+"getAccount", s.handleGet1Acc).Methods("GET")
	s.myr.HandleFunc(s.URL_ROOT+"getAllLoginsDB", s.handleGetAllLoginsDB).Methods("GET")
	s.myr.HandleFunc(s.URL_ROOT+"getAllAccountsDB", s.handleGetAllAccountsDB).Methods("GET")
	s.myr.HandleFunc(s.URL_ROOT+"getAllAccountsBuff", s.handleGetAllAccountsBuff).Methods("GET")

	s.myr.HandleFunc(s.URL_ROOT+"DeleteAllRecordsFromDatabase", s.handleDELETE_ALL_RECORDS_FROM_DATABASE).Methods("DELETE")

	s.m.TestDBConn()

	log.Fatal(http.ListenAndServe(":8888", s.myr))
}
