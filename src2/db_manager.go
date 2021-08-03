package authserv2

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

	// list of sql queries
	query_buff                chan *query
	query_buff_size           int
	query_buff_manager_on     bool
	query_buff_manager_on_mtx *sync.Mutex
	query_unrelized           []*query
	executing_unrelized_mtx   *sync.Mutex

	accs_cache       []*Account
	accs_cache_mtx   *sync.Mutex
	max_cached_accs  int
	observe_interval time.Duration

	LOG_LEVEL int
}

func InitManager(db_username, db_password, db_address string, dp_port int, db_database_name, db_tablename string,
	ping_interval, ping_timeout time.Duration,
	query_buff_size int,
	max_cached_accs int, observe_interval time.Duration) *Manager {
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
	//m.connection = false
	m.db_ping_on_mtx = &sync.Mutex{}

	m.query_buff = make(chan *query, query_buff_size)
	m.query_buff_size = query_buff_size
	m.query_buff_manager_on = false
	m.query_buff_manager_on_mtx = &sync.Mutex{}
	m.query_unrelized = []*query{}
	m.executing_unrelized_mtx = &sync.Mutex{}

	m.accs_cache_mtx = &sync.Mutex{}
	m.accs_cache = []*Account{}
	m.max_cached_accs = max_cached_accs
	m.observe_interval = observe_interval

	m.LOG_LEVEL = 1
	return m
}

func (m *Manager) Start() error {
	m.DBConnection()
	m.ControlConnectionDB(true)
	m.ManageQueryBuffer(true)
	err := m.LoadAccounts()
	if err != nil {
		fmt.Println("Manager.Start(): Manager.LoadAccounts(): err: ", err.Error())
		return err
	}
	return nil
}

func (c *Manager) Log1(str string, values ...interface{}) {
	fmt.Println(c)
	if c.LOG_LEVEL >= 1 {
		Log("Manager: "+str, values...)
	}
}

func (c *Manager) Log2(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 2 {
		Log("Manager: "+str, values...)
	}
}

func (c *Manager) Log3(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 3 {
		Log("Manager: "+str, values...)
	}
}
