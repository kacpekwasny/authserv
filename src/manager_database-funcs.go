package authserv

import (
	"database/sql"
	"fmt"
	"time"

	cmt "github.com/kacpekwasny/commontools"
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

//// ADMIN FUNCS //////
func (m *Manager) getAllLogins() (bool, string, []string) {
	//fmt.Println("getAllLogins()")
	// return succ, msg
	//sqltxt has to have %v in place of table name
	sqltxt := "SELECT login FROM %s"
	rows, err := m.DB.Query(fmt.Sprintf(sqltxt, m.db_tablename))
	defer func() {
		cmt.Pc(rows.Close())
	}()

	if err != nil {
		//fmt.Println("  ERROR: ", err)
		return false, msgERR_WHEN_QUERY + " " + err.Error(), nil
	}
	var login_ls []string
	for rows.Next() {
		var login string
		err := rows.Scan(&login)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, msgERR_NO_ROWS_RETURNED, nil
			}
			return false, msgERR_WHEN_SCANNING_QUERY + " " + err.Error(), nil
		}
		login_ls = append(login_ls, login)
	}
	return true, msgSUCCESS, login_ls
}

func (m *Manager) getAllAccounts() (bool, string, []account) {
	//fmt.Println("getAllAccounts()")
	// return succ, msg
	//sqltxt has to have %v in place of table name
	sqltxt := "SELECT login, pass_hash, last_login, token, logged_in  FROM %s"
	rows, err := m.DB.Query(fmt.Sprintf(sqltxt, m.db_tablename))
	defer func() {
		cmt.Pc(rows.Close())
	}()

	if err != nil {
		//fmt.Println("  ERROR: ", err)
		return false, msgERR_WHEN_QUERY + " " + err.Error(), nil
	}
	var acc_ls []account
	for rows.Next() {
		var acc account
		err := rows.Scan(&acc.Login, &acc.Pass_hash, &acc.Last_login, &acc.Current_token, &acc.Logged_in)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, msgERR_NO_ROWS_RETURNED, nil
			}
			return false, msgERR_WHEN_SCANNING_QUERY + " " + err.Error(), nil
		}
		acc_ls = append(acc_ls, acc)
	}
	return true, msgSUCCESS, acc_ls
}
