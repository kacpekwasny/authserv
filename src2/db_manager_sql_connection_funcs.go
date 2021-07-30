package authserv2

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	cmt "github.com/kacpekwasny/commontools"
)

// connection funcs //
func (m *Manager) DBConnection() {
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?parseTime=True",
			m.db_username, m.db_password, m.db_address, m.db_port, m.db_database_name))
	cmt.Pc(err)
	m.DB = db
}

func (m *Manager) ControlConnectionDB(turn_on bool) {
	// block changing the ping_on
	// check if run only if off
	// set the ping_on
	// unlock possibility of changing
	m.db_ping_on_mtx.Lock()
	defer func() {
		m.db_ping_on = turn_on
		m.db_ping_on_mtx.Unlock()
	}()
	// Make sure only one go routine is Controling connection
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
			lost_connection := false
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
							m.Log3(" TestDBConn, try: %v", try)
						}
						time.Sleep(m.db_ping_timeout / maxtries)
						if ping_recieved {
							if lost_connection {
								m.Log1("Regained connection")
								lost_connection = false
							}
							m.connection = true
							return
						}
						if try > maxtries {
							ping_receieved_mtx.Lock()
							ping_recieved = false
							ping_receieved_mtx.Unlock()
							m.connection = false
							lost_connection = true
							return
						}
						try++
					}
				}()
				// wait for ping timeout check gouroutine
				ping_wg.Wait()

				// info print
				if !m.connection {
					m.Log2("  NO CONNECTION TO DATABASE! PING TIMEOUT! PING NOT RETURNED IN  %v", m.db_ping_timeout)
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

func (m *Manager) ConnectionIsLive() bool {
	return m.connection
}

func (m *Manager) OnRegainedConnection() {
	m.ClearCache()
	m.executing_unrelized_mtx.Lock()
	for len(m.query_unrelized) > 0 {
		q := m.query_unrelized[0]
		_, err := q.exec()
		if err != nil {
			m.connection = false
			break
		}
		m.query_unrelized = m.query_unrelized[1:]
	}
	m.executing_unrelized_mtx.Unlock()
}
