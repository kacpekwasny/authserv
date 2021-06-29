package authserv

import "errors"

const (
	// generic
	msgSUCCESS = "success"
	msgFAIL    = "fail"

	// messages - has to start with msg

	msgERR_WHEN_DBEXEC         = "error when db exec"
	msgERR_WHEN_QUERY          = "error when query"
	msgERR_WHEN_SCANNING_QUERY = "error when scanning query"
	msgERR_NO_ROWS_AFFECTED    = "no rows affected"
	msgERR_NO_ROWS_RETURNED    = "no rows were returned"
	msgLOGIN_NOT_FOUND_DB      = "login is not present in db"
	msgROWS_AFFECTED_Sprintf   = "rows affected: %v"

	msgTOKEN_MISSMATCH    = "token missmatch"
	msgACCOUNT_LOGGED_OUT = "account is logged out"
	msgTOKEN_TIMEOUT      = "token timeout"
	msgPASSWORD_MISSMATCH = "password missmatch"

	// error codes
	erc_SUCCESS                       = "0"  // success,
	erc_CLIENT_ID_NOT_FOUND           = "1"  // client entity id not found
	erc_CLIENT_PASSWORD_MISSMATCH     = "2"  // client entity password missmatch
	erc_UNMARSHAL_ERR                 = "3"  // unmarshal error
	erc_LOGIN_IN_USE                  = "4"  // login in use
	erc_CREDENTIALS_REQUIREMENTS_FAIL = "5"  // credentials dont match requirements
	erc_PROLONG_SESSION_FAIL          = "6"  // failed to prolong session
	erc_REQUEST_AUTH_FAIL             = "7"  // request auth fail
	erc_LOGIN_NOT_FOUND               = "8"  // login not found
	erc_JSON_KEYS_MISSING             = "9"  // not all needed keys are present
	erc_REMOVE_ACCOUNT_FAIL           = "10" // failed to remove account
	erc_CREATE_ACCOUNT_FAIL           = "11" // failed to create account
	erc_DB_ERR                        = "12" // connection to db failed
	erc_GET_ACCOUNT_FAILED            = "13" // getAcc() didnt go successfuly
	erc_CHANGE_LOGIN_FAILED           = "14" // changing login failed
	erc_CHANGE_PASS_FAILED            = "15" // fail during change of pass
	erc_UNAUTHENTICATED               = "16" // account is unauthenticated
	erc_NOT_AUTH_ACC                  = "17" // account is not authenticated
	erc_GET_ALL_LOGINS_ERR            = "18" // get all logins error
	erc_GET_ALL_ACCOUNTS_ERR          = "19" // get all accounts error
	erc_DELETE_ALL_ACCOUNTS_ERR       = "20" // error during DELETE ALL
	erc_ADD_CLIENT_ERR                = "21" // error during adding client
	erc_REMOVE_CLIENT_ERR             = "22" // error during removing client

	// function codes used for authorisation
	// server_handle_admin_funcs.go
	fncd_DeleteAllRecordsFromDataBase = 0
	fncd_GetAllAccountsDB             = 1
	fncd_GetAllAccountsBuff           = 2

	fncd_Get1Acc        = 3
	fncd_GetAllLoginsDB = 4
	fncd_AddClient      = 4
	fncd_RemoveClient   = 5
	fncd_GetClientIds   = 6

	// server_hadnel_funcs.go
	fncd_AddAccount      = 7
	fncd_RemoveAccount   = 8
	fncd_LoginAccount    = 9
	fncd_IsAuthenticated = 10
	fncd_ProlongAuth     = 11
	fncd_LogoutAccount   = 12

	// server.go
	fncd_ChangeLogin = 13
	fncd_ChangePass  = 14
)

var (
	// errors - has to start with ERR
	ERR_TOO_MANY_ROWS_AFFECTED = errors.New("too many rows affected")
)
