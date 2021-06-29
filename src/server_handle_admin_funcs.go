package authserv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

////////////////////////////////
/////// ADMIN FUNCS GET ////////
// GET handle get one account //
func (s *Server) handleGet1Acc(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle Get1Account")
	m := s.intro2handleFuncs(w, fncd_Get1Acc, r, "login")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	succ, msg, acc := s.m.getAcc(m["login"])
	if !succ {
		s.Cnf.Log2("getAcc( %v ), failed. "+msg, m["login"])
		s.respond(w, msgFAIL, erc_GET_ACCOUNT_FAILED, msg, nil)
		return
	}
	s.Cnf.Log3("Successfuly got account: " + acc.Login)
	s.respond(w, msgSUCCESS, erc_SUCCESS, "", *acc)
}

func (s *Server) handleGetAllLoginsDB(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle GetAllLoginsDB")
	m := s.intro2handleFuncs(w, fncd_GetAllLoginsDB, r)
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}

	succ, msg, logins_ls := s.m.getAllLogins()
	if !succ {
		s.Cnf.Log2("getAllLoginsDB(), failed. " + msg)
		s.respond(w, msgFAIL, erc_GET_ALL_LOGINS_ERR, msg, nil)
		return
	}
	s.Cnf.Log3("Successfuly got all logins")
	s.respond(w, msgSUCCESS, erc_SUCCESS, "", logins_ls)
}

func (s *Server) handleGetAllAccountsDB(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle GetAllAccountsDB")
	m := s.intro2handleFuncs(w, fncd_GetAllAccountsDB, r)
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}

	succ, msg, acccounts_ls := s.m.getAllAccounts()
	if !succ {
		s.Cnf.Log2("getAllAccounts(), failed. " + msg)
		s.respond(w, msgFAIL, erc_GET_ALL_ACCOUNTS_ERR, msg, nil)
		return
	}
	s.Cnf.Log3("Successfuly got all accounts")
	s.respond(w, msgSUCCESS, erc_SUCCESS, "", acccounts_ls)
}

func (s *Server) handleGetAllAccountsBuff(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle GetAllAccountsBuff")
	m := s.intro2handleFuncs(w, fncd_GetAllAccountsBuff, r)
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}

	var acc_ls []account
	for _, acc := range *s.m.accountBuffer {
		acc_ls = append(acc_ls, *acc)
	}
	s.Cnf.Log3("Successfuly got all accounts from buffer")
	s.respond(w, msgSUCCESS, erc_SUCCESS, "", acc_ls)
}

func (s *Server) handleDeleteAllRecordsFromDatabase(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle DeleteAllRecordsFromDatabase")
	m := s.intro2handleFuncs(w, fncd_DeleteAllRecordsFromDataBase, r)
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	succ := s.m.DELETE_ALL_RECORDS_IN_DATABASE()
	if !succ {
		s.Cnf.Log2("DELETE_ALL_RECORDS_IN_DATABASE(), failed.")
		s.m.accountBuffer = &[]*account{}
		s.respond(w, msgFAIL, erc_DELETE_ALL_ACCOUNTS_ERR, "fail during delete all accounts", nil)
	}
	s.Cnf.Log3("Successfuly deleted all records")
	s.respond(w, msgSUCCESS, erc_SUCCESS, "succesfuly deleted all records", nil)
}

////////////////////////////////
///////// CLIENT config ////////
func (s *Server) handleAddClient(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle GetAllAccountsDB")
	w.Header().Add("Content-Type", "application/json")
	reqBody, _ := ioutil.ReadAll(r.Body)
	m2 := &map[string]interface{}{}
	err := json.Unmarshal(reqBody, m2)
	if err != nil {
		s.respond(w, msgFAIL, erc_UNMARSHAL_ERR, "json unmarshal error", nil)
		s.Cnf.Log2("  removeClient: JSON UNMARSHAL ERROR")
		return
	}
	// auth client
	auth, msg := s.authClient(w, fncd_AddClient, *m2)
	// Client not authorised
	if !auth {
		// authclient responds by itself
		s.Cnf.Log2("  intro: UNSUCCESSFUL CLIENT AUTHORISATION: ", msg)
		return
	}

	keys := []string{"add_client_id", "add_client_password", "add_client_access_list"}
	ok := checkKeys(*m2, keys...)
	if !ok {
		s.respond(w, msgFAIL, erc_JSON_KEYS_MISSING, fmt.Sprintf("required keys: %v", keys), nil)
		s.Cnf.Log2("  intro: MISSING KEYS. required keys: %v\n", keys)
		return
	}
	var (
		add_client_id          = (*m2)["add_client_id"].(string)
		add_client_password    = (*m2)["add_client_password"].(string)
		add_client_access_list = sIntf2sInteger((*m2)["add_client_access_list"].([]interface{}))
	)

	succ := s.Cnf.AddClient(add_client_id, add_client_password, add_client_access_list)
	if !succ {
		s.Cnf.Log2("addClient(%v, ***, %v), failed. Client is present.", add_client_id, add_client_access_list)
		s.respond(w, msgFAIL, erc_ADD_CLIENT_ERR, "client_id is present", nil)
		return
	}
	s.Cnf.Log3("Successfuly added new client")
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly added new client", nil)
}

func (s *Server) handleRemoveClient(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle RemoveClient")
	m := s.intro2handleFuncs(w, fncd_RemoveClient, r, "remove_client_id")
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	succ := s.Cnf.RemoveClient(m["remove_client_id"])
	if !succ {
		s.Cnf.Log2("RemoveClient( %v ), failed. No client_id", m["remove_client_id"])
		s.respond(w, msgFAIL, erc_REMOVE_CLIENT_ERR, "client_id not found", nil)
		return
	}
	s.Cnf.Log3("Successfuly removed client")
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly removed client", nil)
}

func (s *Server) handleGetClientIds(w http.ResponseWriter, r *http.Request) {
	s.Cnf.Log1("\n – – – handle GetClientIds")
	m := s.intro2handleFuncs(w, fncd_GetClientIds, r)
	if m == nil {
		s.Cnf.Log2("intro2handleFuncs(...) stopped request")
		return
	}
	s.Cnf.Log3("Successfuly got all clients")
	s.respond(w, msgSUCCESS, erc_SUCCESS, "successfuly got clients", s.Cnf.GetClientIDs())
}
