package authserv2

import (
	"database/sql"
	"fmt"
	"time"
)

type query struct {
	stmt   *sql.Stmt
	params []interface{}
}

func (q *query) exec() (sql.Result, error) {
	return q.stmt.Exec(q.params...)
}

func (m *Manager) InitQuery(sql_txt string, params ...interface{}) (*query, error) {
	stmt, err := m.DB.Prepare(sql_txt)
	if err != nil {
		return nil, err
	}
	return &query{
		stmt:   stmt,
		params: params,
	}, nil
}

func (m *Manager) AppendQuery(sql_txt string, params ...interface{}) error {
	q, err := m.InitQuery(sql_txt, params...)
	if err != nil {
		return err
	}
	m.query_buff <- q
	return nil
}

func (m *Manager) ManageQueryBuffer(turn_on bool) {
	m.query_buff_manager_on_mtx.Lock()
	defer func() {
		m.query_buff_manager_on = turn_on
		m.query_buff_manager_on_mtx.Unlock()
	}()
	// Only one managing go routine running at a time
	if turn_on && !m.query_buff_manager_on {
		go func() {
			for m.query_buff_manager_on {

				// wait for unrealized to finish
				m.executing_unrelized_mtx.Lock()
				m.executing_unrelized_mtx.Unlock()

				q := <-m.query_buff
				r, err := q.exec()
				defer q.stmt.Close()
				if err != nil {
					m.execReturnedError(q, err)
				}
				i, _ := r.RowsAffected()
				if i == 0 {
					m.Log3("m.ManagerQueryBuffer(), ZERO RowsAffected() for query: %v", q)
				}
				// untill connection restored or manager turned off
				for !m.connection && m.query_buff_manager_on {
					time.Sleep(time.Second / 10)
				}
			}
		}()
	}
}

func (m *Manager) execReturnedError(q *query, err error) {
	fmt.Println("authserv2: execReturnedError():")
	fmt.Println(q.stmt)
	fmt.Println(q.params...)
	//m.connection = false
	m.executing_unrelized_mtx.Lock()
	m.query_unrelized = append(m.query_unrelized, q)
	m.executing_unrelized_mtx.Unlock()
	panic(err)
}
