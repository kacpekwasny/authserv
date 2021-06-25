package authserv

import (
	"database/sql"
	"fmt"
	"time"
)

// ALL Account atributes DATABASE UPDATE
func (m *Manager) updateLoginDB(login, new_login string) (bool, string) {
	// return: error; SUCCESS (true/false); message
	return m.updateValueDB(login, "login", new_login)
}

func (m *Manager) updatePassHashDB(login, pass_hash string) (bool, string) {
	// return: error; SUCCESS (true/false); message
	return m.updateValueDB(login, "pass_hash", pass_hash)
}

func (m *Manager) updateLastLogin2NowDB(login string) (bool, string) {
	// return: error; SUCCESS (true/false); message
	// set last_login time to the current
	return m.updateValueDB(login, "last_login", formatTime2TimeStamp(time.Now()))
}

func (m *Manager) updateTokenDB(login, token string) (bool, string) {
	// return: error; SUCCESS (true/false); message
	return m.updateValueDB(login, "token", token)
}

func (m *Manager) updateLoggedInDB(login string, logged_in bool) (bool, string) {
	// return: error; SUCCESS (true/false); message
	return m.updateValueDB(login, "logged_in", logged_in)
}

//

// ALL Account atributes DATABASE GET
func (m *Manager) getPassHashDB(login string) (bool, string, string) {
	succ, msg, intface := m.getValueDB(login, "pass_hash")
	if succ {
		return succ, msg, intface.(string)
	}
	return succ, msg, ""
}

func (m *Manager) getLastLoginDB(login string) (bool, string, time.Time) {
	succ, msg, intface := m.getValueDB(login, "last_login")
	if succ {
		return succ, msg, intface.(time.Time)
	}
	return succ, msg, time.Time{}
}

func (m *Manager) getTokenDB(login string) (bool, string, string) {
	succ, msg, intface := m.getValueDB(login, "token")
	if succ {
		return succ, msg, intface.(string)
	}
	return succ, msg, ""
}

func (m *Manager) getLoggedInDB(login string) (bool, string, bool) {
	succ, msg, intface := m.getValueDB(login, "logged_in")
	if succ {
		return succ, msg, intface.(bool)
	}
	return succ, msg, false
}

//// ADMIN FUNCS //////
func (m *Manager) getAllLogins() (bool, string, []string) {
	fmt.Println("getAllLogins()")
	// return succ, msg
	//sqltxt has to have %v in place of table name
	sqltxt := "SELECT login FROM %s"
	rows, err := m.DB.Query(fmt.Sprintf(sqltxt, m.db_tablename))
	defer rows.Close()

	if err != nil {
		fmt.Println("  ERROR: ", err)
		return false, msgERR_WHEN_QUERY, nil
	}
	var login_ls []string
	for rows.Next() {
		var login string
		err := rows.Scan(&login)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, msgERR_NO_ROWS_RETURNED, nil
			}
			return false, msgERR_WHEN_SCANNING_QUERY, nil
		}
		login_ls = append(login_ls, login)
	}
	return true, msgSUCCESS, login_ls
}

func (m *Manager) getAllAccounts() (bool, string, []account) {
	fmt.Println("getAllAccounts()")
	// return succ, msg
	//sqltxt has to have %v in place of table name
	sqltxt := "SELECT login, pass_hash, last_login, token, logged_in  FROM %s"
	rows, err := m.DB.Query(fmt.Sprintf(sqltxt, m.db_tablename))
	defer rows.Close()

	if err != nil {
		fmt.Println("  ERROR: ", err)
		return false, msgERR_WHEN_QUERY, nil
	}
	var acc_ls []account
	for rows.Next() {
		var acc account
		err := rows.Scan(&acc.Login, &acc.Pass_hash, &acc.Last_login, &acc.Current_token, &acc.Logged_in)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, msgERR_NO_ROWS_RETURNED, nil
			}
			return false, msgERR_WHEN_SCANNING_QUERY, nil
		}
		acc_ls = append(acc_ls, acc)
	}
	return true, msgSUCCESS, acc_ls
}
