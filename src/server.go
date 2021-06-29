package authserv

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cmt "github.com/kacpekwasny/commontools"
)

type Server struct {
	m        *Manager
	myr      *mux.Router
	url_root string
	Cnf      *Config
}

func InitServer(URL_ROOT, db_username, db_password, db_address string, dp_port int,
	db_database_name, db_tablename string, ping_interval, ping_timeout time.Duration, max_accountBuffer_len int,
	observe_interval time.Duration) *Server {

	s := &Server{}

	s.m = InitManager(db_username, db_password, db_address, dp_port,
		db_database_name, db_tablename, ping_interval, ping_timeout, max_accountBuffer_len,
		observe_interval)

	s.myr = &mux.Router{}
	s.Cnf = InitConfig()
	s.setURL_ROOT(URL_ROOT)
	return s
}

func (s *Server) setURL_ROOT(root string) {
	l := len(root)
	if l == 0 {
		s.url_root = "/"
		return
	}
	if root[l-1] != '/' {
		root = root + "/"
		s.url_root = root
		return
	}
	s.url_root = root
}

func (s *Server) SetLogLevel(level int) {
	s.Cnf.LOG_LEVEL = level
}

func (s *Server) SetManagerLogLevel(level int) {
	s.m.LOG_LEVEL = level
}

// Auth client entity (NOT USER/ACOUNT), return true for AUTHORISED and false for UNAUTH.
func (s *Server) authClient(w http.ResponseWriter, func_code int, m map[string]interface{}) (bool, string) {
	ok := checkKeys(m, "client_id", "client_password")
	if !ok {
		s.respond(w, msgFAIL, erc_JSON_KEYS_MISSING, "client_id or password missing from json", nil)
		return false, "client_id or password missing from json"
	}
	client_id := m["client_id"].(string)
	client_pass := m["client_password"].(string)

	// get good password assigned to client_id
	client, id_exists := s.Cnf.CLIENT_MAP[client_id]
	good_password := client.pass
	if !id_exists {
		// id not found in client database (not accounts DB)
		s.respond(w, msgFAIL, erc_CLIENT_ID_NOT_FOUND, "client failed to authenticate! client login not found, (credentials to authnticate request to this server)", nil)
		return false, "client failed to authenticate! client login not found, (credentials to authenticate request to this server)"
	}
	if good_password != client_pass {
		// check if password is correct
		s.respond(w, msgFAIL, erc_CLIENT_PASSWORD_MISSMATCH, "client failed to authenticate! client password doesnt match", nil)
		return false, fmt.Sprintf("client: '%s' failed to authenticate! client password doesnt match", client.id)
	}
	if !IntInSlice(func_code, client.func_id_access_list) {
		// client is not authorised for this action
		s.respond(w, msgFAIL, erc_CLIENT_PASSWORD_MISSMATCH, fmt.Sprintf("client '%s' is not authorised for this function! This func code: %v", client_id, func_code), nil)
		return false, fmt.Sprintf("client: '%s' is not authorised for this action. This func code: %v", client_id, func_code)
	}

	s.Cnf.Log2("OK authClient(...), '%v'.", client.id)
	return true, ""
}

// Shortcut for usual response
func (s *Server) respond(w http.ResponseWriter, status, err_code, msg string, get_resp interface{}) {
	w.Header().Add("Content-Type", "application/json")
	resp := map[string]interface{}{
		"status":   status,
		"err_code": err_code,
		"msg":      msg,
	}
	if get_resp != nil {
		resp["additional"] = get_resp
	}
	cmt.Pc(json.NewEncoder(w).Encode(resp))
}

// Check client auth, check if keys are present, json2map, handle errors and respond, return nil if error and map on success
// Will respond and returns nil if something is wrong.
//  If returns map:
//  - json2map was successful,
//  - Client is authenticated,
//  - Client is authorised,
//  - All keys are present
func (s *Server) intro2handleFuncs(w http.ResponseWriter, func_code int, r *http.Request, keys ...string) map[string]string {
	w.Header().Add("Content-Type", "application/json")
	m, err := json2map(r)
	// read data that will be used to create an account
	// Unmarshall error
	if err != nil {
		s.respond(w, msgFAIL, erc_UNMARSHAL_ERR, "json unmarshal error", nil)
		s.Cnf.Log2("  intro: JSON UNMARSHAL ERROR")
		return nil
	}

	// auth client
	auth, msg := s.authClient(w, func_code, mS2mI(m))
	// Client not authenticated or authorised
	if !auth {
		// authclient responds by itself so no need for responding
		s.Cnf.Log2("  intro: UNSUCCESSFUL authClient: ", msg)
		return nil
	}

	ok := checkKeys(mS2mI(m), keys...)
	if !ok {
		s.respond(w, msgFAIL, erc_JSON_KEYS_MISSING, fmt.Sprintf("required keys: %v", keys), nil)
		s.Cnf.Log2("  intro: MISSING KEYS. required keys: %v\n", keys)
		return nil
	}
	return m
}

func (s *Server) DisplayStartInfoFromDB(wait_for_connection bool, success *bool) {
	go func() {
		tries := 0
		if wait_for_connection {
			for !s.m.connection {
				// connection is not ready yet
				if tries > 1000 {
					s.Cnf.Log1("DisplayInfoFromDB() tried %v times, with no success", tries)
					*success = false
					return
				}
				time.Sleep(time.Millisecond * 50)
				tries++
			}
		}

		succ, msg, accs := s.m.getAllLogins()
		if !succ {
			s.Cnf.Log1("Startup info: s.m.getAllLogins() unsuccesful. msg: " + msg)
			*success = false
			return
		}
		s.Cnf.Log1(" <<< Startup info: accounts in database: %v >>> ", len(accs))
		*success = true
	}()
}

//////////////////////////////////////

func (s *Server) ListenAndServe() {

	// client_id and password are to authenticate entity trying to create an account
	s.myr.HandleFunc(s.url_root+"addAccount", s.handleAddAccount).Methods("POST")
	s.myr.HandleFunc(s.url_root+"removeAccount", s.handleRemoveAccount).Methods("DELETE")
	s.myr.HandleFunc(s.url_root+"loginAccount", s.handleLoginAccount).Methods("POST")
	s.myr.HandleFunc(s.url_root+"prolongAuth", s.handleProlongAuth).Methods("POST")
	s.myr.HandleFunc(s.url_root+"isAuthenticated", s.handleIsAuthenticated).Methods("POST")
	s.myr.HandleFunc(s.url_root+"logoutAccount", s.handleLogoutAccount).Methods("POST")

	s.myr.HandleFunc(s.url_root+"changeLogin", s.handleChangeLogin).Methods("POST")
	s.myr.HandleFunc(s.url_root+"changePass", s.handleChangePass).Methods("POST")

	// admin funcs
	// get accounts funcs
	s.myr.HandleFunc(s.url_root+"getAccount", s.handleGet1Acc).Methods("GET")
	s.myr.HandleFunc(s.url_root+"getAllLoginsDB", s.handleGetAllLoginsDB).Methods("GET")
	s.myr.HandleFunc(s.url_root+"getAllAccountsDB", s.handleGetAllAccountsDB).Methods("GET")
	s.myr.HandleFunc(s.url_root+"getAllAccountsBuff", s.handleGetAllAccountsBuff).Methods("GET")

	// delete all accounts func
	s.myr.HandleFunc(s.url_root+"DeleteAllRecordsFromDatabase", s.handleDeleteAllRecordsFromDatabase).Methods("DELETE")

	// client funcs
	s.myr.HandleFunc(s.url_root+"addClient", s.handleAddClient).Methods("POST")
	s.myr.HandleFunc(s.url_root+"removeClient", s.handleRemoveClient).Methods("DELETE")
	s.myr.HandleFunc(s.url_root+"getClientIDs", s.handleGetClientIds).Methods("GET")

	i := false
	s.DisplayStartInfoFromDB(true, &i)
	s.m.ControlConnectionDB(true)
	log.Fatal(http.ListenAndServe(":8888", s.myr))
}
