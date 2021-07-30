package authserv2

import (
	"database/sql"
	"fmt"

	cmt "github.com/kacpekwasny/commontools"
)

func (m *Manager) DBQueryRow(sqltxt string, vars []interface{}, pointers ...interface{}) error {
	row := m.DB.QueryRow(fmt.Sprintf(sqltxt, m.db_tablename), vars...)
	err := row.Scan(pointers...)
	switch err {
	case nil:
		return nil
	case sql.ErrNoRows:
		return ErrLoginNotFound
	default:
		m.Log1("DBQueryRow(): row.Scan(): error: %v", err)
		//m.connection = false
		return ErrDBoffline
	}
}

func (m *Manager) GetAccountDB(login string) (*Account, error) {
	acc := &Account{}
	acc.Login = login
	return acc, m.DBQueryRow("SELECT pass_hash, token, last_login, logged_in FROM %s WHERE login=?", []interface{}{login},
		&acc.Pass_hash, &acc.Current_token, &acc.Last_login, &acc.Logged_in)
}

func (m *Manager) GetAllAccounts() ([]*Account, error) {
	sqltxt := "SELECT login, pass_hash, last_login, token, logged_in  FROM %s"
	rows, err := m.DB.Query(fmt.Sprintf(sqltxt, m.db_tablename))
	if err != nil {
		if err == sql.ErrNoRows {
			return []*Account{}, nil
		}
		fmt.Println("db_manager_sql_read_funcs.go: GetAllAccounts(), err =", err.Error())
		return nil, ErrDBoffline
	}
	if err := rows.Close(); err != nil {
		fmt.Println("db_manager_sql_read_funcs.go: GetAllAccounts(), rows.Close(): set //m.connection = false, err =", err)
		//m.connection = false
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("db_manager_sql_read_funcs.go: GetAllAccounts(), recover(): set //m.connection = false, err =", err.Error())
			//m.connection = false
		}
	}()
	if err != nil {
		return nil, err
	}
	var acc_ls []*Account
	for rows.Next() {
		var acc *Account
		err := rows.Scan(&acc.Login, &acc.Pass_hash, &acc.Last_login, &acc.Current_token, &acc.Logged_in)
		if err != nil {
			return nil, err
		}
		acc_ls = append(acc_ls, acc)
	}
	return acc_ls, nil
}

func (m *Manager) GetLogins() ([]string, error) {
	sqltxt := "SELECT login FROM %s"
	rows, err := m.DB.Query(fmt.Sprintf(sqltxt, m.db_tablename))
	defer func() {
		cmt.Pc(rows.Close())
	}()
	if err != nil {
		return nil, err
	}
	var logins []string
	for rows.Next() {
		var login string
		err := rows.Scan(&login)
		if err != nil {
			return nil, err
		}
		logins = append(logins, login)
	}
	return logins, nil
}

func (m *Manager) LoadAccounts() error {
	accs, err := m.GetAllAccounts()
	if err != nil {
		return err
	}
	m.accs_cache_mtx.Lock()
	m.accs_cache = accs
	m.accs_cache_mtx.Unlock()
	return nil
}

// return nil if login exists
func (m *Manager) LoginExistsDB(login string) error {
	sqltxt := fmt.Sprintf("SELECT login FROM %s WHERE login =?", m.db_tablename)
	err := m.DB.QueryRow(sqltxt, login).Scan(&login)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrLoginNotFound
		}
		return err
	}
	return nil
}
