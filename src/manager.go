package authserv

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	cmt "github.com/kacpekwasny/commontools"
)

type Manager struct {
	db_username      string
	db_password      string
	db_address       string
	db_port          int
	db_database_name string
	db_tablename     string

	DB               *sql.DB
	db_ping_interval time.Duration
	db_ping_timeout  time.Duration
	db_ping_on_mtx   *sync.Mutex
	db_ping_on       bool
	connection       bool
	ping_err         error

	accountBuffer         *[]*account
	accBuffMTX            *sync.Mutex
	max_accountBuffer_len int
	observe_interval      time.Duration

	LOG_LEVEL int
}

func InitManager(db_username, db_password, db_address string, dp_port int, db_database_name, db_tablename string,
	ping_interval, ping_timeout time.Duration,
	max_accountBuffer_len int, observe_interval time.Duration) *Manager {
	m := &Manager{}

	m.db_username = db_username
	m.db_password = db_password
	m.db_address = db_address
	m.db_port = dp_port
	m.db_database_name = db_database_name
	m.db_tablename = db_tablename

	m.db_ping_interval = ping_interval
	m.db_ping_timeout = ping_timeout
	m.db_ping_on = false
	m.connection = false
	m.db_ping_on_mtx = &sync.Mutex{}

	m.accBuffMTX = &sync.Mutex{}
	m.accountBuffer = &[]*account{}
	m.max_accountBuffer_len = max_accountBuffer_len
	m.observe_interval = observe_interval

	m.LOG_LEVEL = 1

	m.DBConnection()
	return m
}

// Database funcs //
func (m *Manager) DBConnection() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?parseTime=True", m.db_username,
		m.db_password, m.db_address, m.db_port, m.db_database_name))
	cmt.Pc(err)
	m.DB = db
}

func (m *Manager) ControlConnectionDB(turn_on bool) {
	defer func() {
		m.db_ping_on_mtx.Lock()
		m.db_ping_on = turn_on
		m.db_ping_on_mtx.Unlock()
	}()
	// It is supposed to turn on if it is not allready turned on in another goroutine
	if turn_on && !m.db_ping_on {
		go func() {
			// set test ping as OFF when returning
			defer func() {
				m.db_ping_on_mtx.Lock()
				m.db_ping_on = false
				m.db_ping_on_mtx.Unlock()
			}()
			m.Log1(" TestDBConn() - START")
			ping_wg := &sync.WaitGroup{}
			ping_receieved_mtx := &sync.Mutex{}
			printed_connection_made := false
			for {
				ping_recieved := false
				// Start ping
				go func() {
					m.PingDB()
					ping_receieved_mtx.Lock()
					ping_recieved = (m.ping_err == nil)
					ping_receieved_mtx.Unlock()
				}()

				// check for ping timeout
				ping_wg.Add(1)
				go func() {
					defer ping_wg.Done()
					const maxtries = 10
					try := 0
					for {
						if !m.connection {
							m.Log1("goroutine TestDBConn, try: %v", try)
						}
						time.Sleep(m.db_ping_timeout / maxtries)
						if ping_recieved {
							m.connection = true
							return
						}
						if try > maxtries {
							ping_receieved_mtx.Lock()
							ping_recieved = false
							ping_receieved_mtx.Unlock()
							m.connection = false
							return
						}
						try++
					}
				}()
				// wait for ping timeout check gouroutine
				ping_wg.Wait()

				// info print
				if !m.connection {
					m.Log1("  NO CONNECTION TO DATABASE! PING TIMEOUT! PING NOT RETURNED IN  %v", m.db_ping_timeout)
					printed_connection_made = false
					// try to regain connection
					go m.DBConnection()
				}
				if !printed_connection_made && m.connection {
					m.Log1("Connection to db made. Ping is succesful.")
					printed_connection_made = true
				}
				if !m.db_ping_on {
					m.Log1(" TestDBConn() - END")
					return
				}
				time.Sleep(m.db_ping_interval)
			}
		}()
	}
}

func (m *Manager) PingDB() {
	m.ping_err = m.DB.Ping()
}

func (m *Manager) dbExec(sqltxt string, vars ...interface{}) (bool, string) {
	// return succ, msg
	res, err := m.DB.Exec(sqltxt, vars...)
	if err != nil {
		//fmt.Println(err)
		return false, err.Error()
	}
	rows, _ := res.RowsAffected()
	if rows < 1 {
		return false, msgERR_NO_ROWS_AFFECTED
	}
	return true, fmt.Sprintf(msgROWS_AFFECTED_Sprintf, rows)
}

func (m *Manager) dbQueryRow(sqltxt string, vars []interface{}, pointers ...interface{}) (bool, string) {
	//fmt.Printf("dbQueryRow(%v, %v)\n", sqltxt, vars)
	// return succ, msg
	//sqltxt has to have %v in place of table name
	row := m.DB.QueryRow(fmt.Sprintf(sqltxt, m.db_tablename), vars...)

	//var intface_ls []interface{}
	//for _ = range pyfuncs.Range(values - 1) {
	//	intface := &interface{}
	//	intface_ls = append(intface_ls, intface)
	//}
	err := row.Scan(pointers...)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, msgERR_NO_ROWS_RETURNED
		}
		return false, msgERR_WHEN_SCANNING_QUERY + " " + err.Error()
	}
	return true, msgSUCCESS
}

func (m *Manager) MakeTable(table_name string) (bool, string) {
	_, err := m.DB.Exec("CREATE TABLE IF NOT EXISTS " + table_name + " (login VARCHAR(50) PRIMARY KEY, pass_hash TINYTEXT, last_login TIMESTAMP, token TINYTEXT, logged_in BOOLEAN)")
	if err != nil {
		m.Log1(" MakeTable(%v) FAILED : false, %v", table_name, err.Error())
		return false, err.Error()
	}
	m.Log1(" MakeTable(%v) SUCCESS : true, ''", table_name)
	return true, ""
}

// ~~~~~~~~~~~~~~~~ //

// BUFFER FUNCS ~~~~ START //
func (m *Manager) ObserveAccBuff() {
	//fmt.Println("ObserverAccBuff()")
	for {
		if len(*m.accountBuffer) > m.max_accountBuffer_len {
			acc := m.popLastAccfromBuff()
			if acc != nil {
				m.insertAccDB(acc)
			}
		}
		time.Sleep(m.observe_interval)
	}
}

func (m *Manager) popLastAccfromBuff() *account {
	//fmt.Println("popLastAccfromBuff()")
	m.accBuffMTX.Lock()
	defer m.accBuffMTX.Unlock()
	accLsLen := len(*m.accountBuffer)
	if accLsLen > 0 {
		acc := (*(m.accountBuffer))[accLsLen-1]
		*m.accountBuffer = (*m.accountBuffer)[:accLsLen-1]
		return acc
	}
	return nil
}

func (m *Manager) getAccBuff(login string) (bool, *account) {
	//fmt.Println("getAccBuff()")
	m.accBuffMTX.Lock()
	defer m.accBuffMTX.Unlock()
	for _, acc := range *m.accountBuffer {
		if acc.Login == login {
			return true, acc
		}
	}
	return false, nil
}

func (m *Manager) appendAccToBuff(acc *account) {
	//fmt.Println("appendAccToBuff()")
	if acc == nil {
		return
	}
	m.accBuffMTX.Lock()
	defer m.accBuffMTX.Unlock()
	*m.accountBuffer = append(*m.accountBuffer, acc)
}

func (m *Manager) accIsInBuff(login string) bool {
	//fmt.Println("accIsInBuff()")
	for _, acc := range *m.accountBuffer {
		if acc.Login == login {
			return true
		}
	}
	return false
}

func (m *Manager) loadAccToBuff(login string) (bool, string) {
	//fmt.Printf("loadAccToBuff(%v)\n", login)
	// return success - false when no login found and
	if m.accIsInBuff(login) {
		return true, msgSUCCESS
	} else {
		succ, msg, acc := m.getAccDB(login)
		if succ {
			m.appendAccToBuff(acc)
			return true, msgSUCCESS
		}
		if msg == msgERR_NO_ROWS_RETURNED {
			return false, msgLOGIN_NOT_FOUND_DB
		}
		return false, msg
	}
}

func (m *Manager) deleteAccBuff(login string) bool {
	//fmt.Printf("deleteAccBuff(%v)\n", login)
	m.accBuffMTX.Lock()
	var (
		index  int
		found  bool
		length int = len(*m.accountBuffer)
	)
	for i, acc := range *m.accountBuffer {
		if acc.Login == login {
			found = true
			index = i
			break
		}
	}
	if !found {
		m.accBuffMTX.Unlock() //
		return false
	}
	if index < length-1 {
		*m.accountBuffer = append((*m.accountBuffer)[:index], (*m.accountBuffer)[index+1:]...)
		m.accBuffMTX.Unlock() //
		return true
	}
	//
	m.accBuffMTX.Unlock() //
	m.popLastAccfromBuff()
	return true
}

// END //

// ~~~~ CRUD ~~~~ FUNCS ~~~~ //
// DATABASE
func (m *Manager) insertAccDB(acc *account) (bool, string) {
	// return error; SUCCESS (true/false); message

	//fmt.Printf("newAccDB(%v)\n", *acc)

	sqltxt := fmt.Sprintf("INSERT INTO %v (login, pass_hash, last_login, token,logged_in) VALUES(?, ?, ?, ?, ?)", m.db_tablename)
	res, err := m.DB.Exec(sqltxt, acc.Login, acc.Pass_hash, formatTime2TimeStamp(acc.Last_login), acc.Current_token, acc.Logged_in)
	if err != nil {
		return false, msgERR_WHEN_DBEXEC
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return false, msgERR_NO_ROWS_AFFECTED
	}
	if rows == 1 {
		return true, msgSUCCESS
	}
	return true, fmt.Sprintf("Too many rows affected. Affected: %v", rows)
}

func (m *Manager) getAccDB(login string) (bool, string, *account) {
	//fmt.Printf("getAccDB(%v)\n", login)
	acc := &account{}
	acc.Login = login
	succ, msg := m.dbQueryRow("SELECT pass_hash, token, last_login, logged_in FROM %s WHERE login=?", []interface{}{login},
		&acc.Pass_hash, &acc.Current_token, &acc.Last_login, &acc.Logged_in)
	if !succ {
		return false, msg, nil
	}
	return true, msgSUCCESS, acc
}

func (m *Manager) updateValueDB(login, col_name string, value interface{}) (bool, string) {
	// return error; SUCCESS (true/false); message
	//fmt.Printf("updateValueDB('%v', '%v', %v) \n", login, col_name, value)
	return m.dbExec(fmt.Sprintf("UPDATE %s SET %s = ? WHERE login = ?", m.db_tablename, col_name), value, login)
}

func (m *Manager) deleteAccDB(login string) (bool, string) {
	//fmt.Printf("deleteAccDB('%v')\n", login)
	return m.dbExec(fmt.Sprintf("DELETE FROM %s WHERE login = ?", m.db_tablename), login)
}

func (m *Manager) loginExistsDB(login string) (bool, bool, string) {
	// success, exists, msg
	//fmt.Printf("loginExistsDB(%v)\n", login)
	sqltxt := fmt.Sprintf("SELECT login FROM %s WHERE login =?", m.db_tablename)
	err := m.DB.QueryRow(sqltxt, login).Scan(&login)
	if err != nil {
		if err == sql.ErrNoRows {
			// 	   succ, doesnt exist
			return true, false, msgSUCCESS
		}
		return false, false, msgERR_WHEN_SCANNING_QUERY
	}
	//     succ, exists
	return true, true, msgSUCCESS
}

func (m *Manager) DELETE_ALL_RECORDS_IN_DATABASE() bool {
	// return success
	//fmt.Println("DELETE_ALL_RECORDS_IN_DATABASE()")
	sqltxt := fmt.Sprintf("DELETE FROM %s;", m.db_tablename)
	res, err := m.DB.Exec(sqltxt)
	rows, _ := res.RowsAffected()
	fmt.Println("   DELETE ALL RECORDS, ROWS AFFECTED:", rows)
	if err != nil {
		return err == sql.ErrNoRows
	}
	return true
}

// ~~~~ END ~~~~ CRUD ~~~~~ //
//
//
//
///////////////////////////////////////
//// LOG FUNCS, same as for Config ////

func (c *Manager) Log1(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 1 {
		Log(str, values...)
	}
}

func (c *Manager) Log2(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 2 {
		Log(str, values...)
	}
}

func (c *Manager) Log3(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 3 {
		Log(str, values...)
	}
}