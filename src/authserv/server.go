package authserv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Auth client entity, return true for AUTHORISED and false for UNAUTH.
func authClient(w http.ResponseWriter, r *http.Request) (bool, string) {
	vars := mux.Vars(r)
	client_id := vars["client_id"]
	pass := vars["password"]

	// get good password assigned to client_id
	good_password, id_exists := CLIENT_PASSWORD[client_id]
	if !id_exists {
		// id not found in client database (not accounts DB)
		respond(w, "fail", "1", "client failed to authorise! client login not found, (credentials to authorise request to this server)")
		return false, "id doesnt exist"
	}
	if good_password != pass {
		respond(w, "fail", "2", "client failed to authorise! client password doesnt match")
		return false, "password missmatch"
	}

	return true, ""
}

// json from Request to map[string]string
func json2map(r *http.Request) (map[string]string, error) {
	// read data that will be used to create an account
	reqBody, _ := ioutil.ReadAll(r.Body)
	m := &map[string]string{}
	err := json.Unmarshal(reqBody, m)

	return *m, err
}

// Shortcut for usual response
func respond(w http.ResponseWriter, status, err_code, msg string) {
	w.Header().Add("Content-Type", "application/json")
	resp := map[string]string{
		"status":   status,
		"err_code": err_code,
		"msg":      msg,
	}
	json.NewEncoder(w).Encode(resp)
}

// POST handle add account
func handleAddAccount(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Add("Content-Type", "application/json")

	// auth client
	auth, _ := authClient(w, r)
	// Client not authorised
	if !auth {
		return
	}

	// read data that will be used to create an account
	m, err := json2map(r)
	// Unmarshall error
	if err != nil {
		respond(w, "fail", "3", "json unmarshal error")
		return
	}

	// check if all keys are present
	ok := checkKeys(m, "login", "pass")
	if !ok {
		respond(w, "fail", "9", "not all needed keys are present")
		return
	}

	_, login_in_use := getAccByLogin(m["login"])
	// Login is used
	if login_in_use == nil {
		// login is in use
		respond(w, "fail", "4", "login is in use allready")
		return
	}

	// check if login and pass match requirements
	login_ok, login_msg := loginMatchesReq(m["login"])
	pass_ok, pass_msg := passMatchesReq(m["pass"])
	if !(login_ok && pass_ok) {
		respond(w, "fail", "5",
			"credentials dont match requirements. login: "+login_msg+";  pass: "+pass_msg)
		return
	}

	acc := makeAccount(m["login"], m["pass"])
	addAccount(acc)
	fmt.Println("Successfully added new account. Number of accounts: ", len(*accountList))
	respond(w, "success", "0", "successfully created account")
}

// POST remove account from accountList
func handleRemAccount(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")

	// auth client
	auth, _ := authClient(w, r)
	// Client not authorised
	if !auth {
		return
	}

	m, err := json2map(r)
	if err != nil {
		respond(w, "fail", "3", "json unmarshal error")
		return
	}

	// check if all keys are present
	ok := checkKeys(m, "login")
	if !ok {
		respond(w, "fail", "9", "not all needed keys are present")
		return
	}

	ok, msg := removeAccount(m["login"])
	if !ok {
		respond(w, "fail", "10", "failed to remove account. "+msg)
		return
	}
	respond(w, "success", "0", "successfuly removed account")
}

// POST handle request for login
func handleRequestAuth(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Add("Content-Type", "application/json")

	// auth client
	auth, _ := authClient(w, r)
	// Client not authorised
	if !auth {
		return
	}

	m, err := json2map(r)
	if err != nil {
		respond(w, "fail", "3", "json unmarshal error")
		return
	}

	// check if all keys are present
	ok := checkKeys(m, "login", "pass")
	if !ok {
		respond(w, "fail", "9", "not all needed keys are present")
		return
	}

	ok, acc, msg := requestAuth(m["login"], m["pass"])
	if !ok {
		respond(w, "fail", "7", "request auth failed. "+msg)
		return
	}

	resp := map[string]string{}
	resp["status"] = "success"
	resp["err_code"] = "0"
	resp["msg"] = "successfully loged in"
	resp["token"] = acc.Current_token
	json.NewEncoder(w).Encode(resp)
}

// POST handle request to prelong token life
func handlePrelongAuth(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Add("Content-Type", "application/json")

	// auth client (and respond if fail)
	auth, _ := authClient(w, r)
	// Client not authorised
	if !auth {
		return
	}

	m, err := json2map(r)
	if err != nil {
		respond(w, "fail", "3", "json unmarshal error")
		return
	}

	// check if all keys are present
	ok := checkKeys(m, "login", "token")
	if !ok {
		respond(w, "fail", "9", "not all needed keys are present")
		return
	}

	ok, msg := prelongAuth(m["login"], m["token"])
	if !ok {
		respond(w, "fail", "6", "failed to prelong session. "+msg)
		return
	}

	respond(w, "success", "0", "succesfully prelonged session.")
}

// POST handle Logout request
func handleLogout(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")

	// auth client (and respond if fail)
	auth, _ := authClient(w, r)
	// Client not authorised
	if !auth {
		return
	}

	m, err := json2map(r)
	if err != nil {
		respond(w, "fail", "3", "json unmarshal error")
		return
	}

	ok, msg := prelongAuth(m["login"], m["token"])
	if !ok {
		respond(w, "fail", "6", "need to be able to prelong session to logout. "+msg)
		return
	}

	acc, err := getAccByLogin(m["login"])
	if err != nil {
		respond(w, "fail", "8", "login not found")
		return
	}

	acc.logout()
	respond(w, "success", "0", "successful logout")
}

// GET handle getaccounts and respond with json list of accounts
func handleGetAccounts(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")

	// auth client (and respond if fail)
	auth, _ := authClient(w, r)
	// Client not authorised
	if !auth {
		return
	}

	acclist := []account{}
	for _, acc := range *accountList {
		acclist = append(acclist, *acc)
	}
	json.NewEncoder(w).Encode(acclist)
}

// GET handle get one account
func handleGet1Acc(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")

	// auth client (and respond if fail)
	auth, _ := authClient(w, r)
	// Client not authorised
	if !auth {
		return
	}

	vars := mux.Vars(r)
	login := vars["login"]
	acc, err := getAccByLogin(login)
	if err != nil {
		respond(w, "fail", "8", "login not found")
		return
	}

	json.NewEncoder(w).Encode(*acc)

}
func HandleRequests() {
	// myRouter
	myr := mux.NewRouter().StrictSlash(true)

	// client_id and password are to authenticate entity trying to create an account
	myr.HandleFunc("/authserv/addaccount/{client_id}/{password}", handleAddAccount).Methods("POST")
	myr.HandleFunc("/authserv/remaccount/{client_id}/{password}", handleRemAccount).Methods("POST")
	myr.HandleFunc("/authserv/reqauth/{client_id}/{password}", handleRequestAuth).Methods("POST")
	myr.HandleFunc("/authserv/prelongauth/{client_id}/{password}", handlePrelongAuth).Methods("POST")
	myr.HandleFunc("/authserv/logout/{client_id}/{password}", handleLogout).Methods("POST")

	myr.HandleFunc("/authserv/getaccounts/{client_id}/{password}", handleGetAccounts)
	myr.HandleFunc("/authserv/getaccount/{login}/{client_id}/{password}", handleGet1Acc)

	log.Fatal(http.ListenAndServe(":8888", myr))

}
