package authserv

import "time"

// Checks if time since last logon is not too long
func (acc *account) durationOK(max_token_age time.Duration) bool {

	return time.Duration(time.Since(acc.Last_login)) < time.Duration(max_token_age)
}

// ret true and prelong acc if login, token and time ok
// if smth wrong return false and msg
func prelongAuth(login, token string) (bool, string) {

	// check if there is an account with such login
	acc, err := getAccByLogin(login)
	if err != nil {
		return false, "Login not found"
	}

	// check if tokens match
	if acc.Current_token != token {
		return false, "Token missmatch"
	}

	// check if user loged out
	if !acc.Loged_in {
		return false, "Account is logged out"
	}

	// check if this account's last login wasnt to long ago
	if !acc.durationOK(MAX_TOKEN_AGE) {
		acc.logout()
		return false, "Token timeout"
	}

	acc.uLogTime()
	acc.login()
	return true, ""

}

// return false and message if failure
// return true and "" if success
func requestAuth(login, pass string) (bool, *account, string) {

	// check if there is an account with such a login
	acc, err := getAccByLogin(login)
	if err != nil {
		return false, nil, "Login not found"
	}

	// check if hash matches
	if !passIsCorrect(pass, acc.Pass_hash) {
		return false, acc, "Password missmatch"
	}

	acc.uLogTime()
	acc.newToken()
	acc.login()

	return true, acc, ""
}
