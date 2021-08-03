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
			m.Log1(" ControlConnectionDB() - START")
			pch := make(chan error, 1)
			mu := sync.Mutex{}
			counter := 0
			for {
				counter = counter % 5
				go func() {
					pch <- m.PingDB()
				}()

				select {
				case <-pch:
					mu.Lock()
					if !m.connection {
						m.Log1("Regained connection")
					}
					m.connection = true
					mu.Unlock()
				case <-time.After(m.db_ping_timeout):
					mu.Lock()
					if m.connection {
						m.Log1("Lost connection")
					} else {
						// dont spam "No connection to database"
						// print "db disconnected" every couple tries
						if counter == 0 {
							m.Log1("No connection to database")
						}
					}
					m.connection = false
					mu.Unlock()
				}
				time.Sleep(m.db_ping_interval)
				counter++
			}
		}()
	}
}

func (m *Manager) PingDB() error {
	return m.DB.Ping()
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
