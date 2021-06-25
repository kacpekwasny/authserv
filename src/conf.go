package authserv

import "time"

var (
	MAX_TOKEN_AGE   = time.Hour
	TOKEN_LENGTH    = 50
	CLIENT_PASSWORD = map[string]string{} // {client_name1:password1, client_name2:password2}
	PASS_MAX_LEN    = 30
	PASS_MIN_LEN    = 4
	LOGIN_MAX_LEN   = 40
	LOGIN_MIN_LEN   = 1
)

func passMatchesReq(pass string) (bool, string) {
	if !(PASS_MIN_LEN <= len(pass) && len(pass) <= PASS_MAX_LEN) {
		return false, "length not satisfied"
	}
	return true, ""
}

func loginMatchesReq(login string) (bool, string) {
	if !(LOGIN_MIN_LEN <= len(login) && len(login) <= LOGIN_MAX_LEN) {
		return false, "length not satisfied"
	}
	return true, ""
}
