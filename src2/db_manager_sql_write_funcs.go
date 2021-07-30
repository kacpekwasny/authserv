package authserv2

import (
	"fmt"

	cmt "github.com/kacpekwasny/commontools"
)

func (m *Manager) UpdateValueDB(login, col_name string, value interface{}) {
	cmt.Pc(m.AppendQuery(fmt.Sprintf("UPDATE %s SET %s = ? WHERE login = ?", m.db_tablename, col_name), value, login))
}

func (m *Manager) DeleteAccountDB(login string) {
	cmt.Pc(m.AppendQuery(fmt.Sprintf("DELETE FROM %s WHERE login = ?", m.db_tablename), login))
}

func (m *Manager) InsertAccountDB(acc *Account) {
	sqltxt := fmt.Sprintf("INSERT INTO %v (login, pass_hash, last_login, token,logged_in) VALUES(?, ?, ?, ?, ?)", m.db_tablename)
	cmt.Pc(m.AppendQuery(sqltxt, acc.Login, acc.Pass_hash, formatTime2TimeStamp(acc.Last_login), acc.Current_token, acc.Logged_in))
}

func (m *Manager) DELETE_ALL_RECORDS_IN_DATABASE() (int64, error) {
	// return success
	//fmt.Println("DELETE_ALL_RECORDS_IN_DATABASE()")
	sqltxt := fmt.Sprintf("DELETE FROM %s;", m.db_tablename)
	res, err := m.DB.Exec(sqltxt)
	rows, _ := res.RowsAffected()
	m.Log1("DELETE ALL RECORDS, ROWS AFFECTED: %v", rows)
	if err != nil {
		return rows, err
	}
	return rows, err
}
