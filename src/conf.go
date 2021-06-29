package authserv

import (
	"fmt"
	"time"
)

type Config struct {
	MAX_TOKEN_AGE time.Duration
	TOKEN_LENGTH  int
	CLIENT_MAP    map[string]client
	PASS_MAX_LEN  int
	PASS_MIN_LEN  int
	LOGIN_MAX_LEN int
	LOGIN_MIN_LEN int
	LOG_LEVEL     int
}

type client struct {
	id                  string
	pass                string
	func_id_access_list []int
}

func InitConfig() *Config {
	c := Config{}
	c.MAX_TOKEN_AGE = time.Hour
	c.TOKEN_LENGTH = 50
	c.CLIENT_MAP = map[string]client{} // {client_id1:client{}, client_id2:client{}}
	c.PASS_MAX_LEN = 30
	c.PASS_MIN_LEN = 4
	c.LOGIN_MAX_LEN = 40
	c.LOGIN_MIN_LEN = 1
	return &c
}

func (c *Config) passMatchesReq(pass string) (bool, string) {
	if !(c.PASS_MIN_LEN <= len(pass) && len(pass) <= c.PASS_MAX_LEN) {
		return false, "length not satisfied"
	}
	return true, ""
}

func (c *Config) loginMatchesReq(login string) (bool, string) {
	if !(c.LOGIN_MIN_LEN <= len(login) && len(login) <= c.LOGIN_MAX_LEN) {
		return false, "length not satisfied"
	}
	return true, ""
}

// Add client and return info it there wasnt one allready
func (c *Config) AddClient(id, pass string, func_id_access_list []int) bool {
	_, id_exists := c.CLIENT_MAP[id]
	if id_exists {
		return false
	}
	c.CLIENT_MAP[id] = client{id, pass, func_id_access_list}
	return true
}

// delete  client and return info if there was such id
func (c *Config) RemoveClient(id string) bool {
	_, id_exists := c.CLIENT_MAP[id]
	if id_exists {
		delete(c.CLIENT_MAP, id)
		return true
	}
	return false
}

func (c *Config) GetClientIDs() []string {
	ls := []string{}
	for id, _ := range c.CLIENT_MAP {
		ls = append(ls, id)
	}
	return ls
}

func (c *Config) Log1(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 1 {
		fmt.Printf(time.Now().Format("02/01/2006 – 15:04:05 SrvLog: ")+str+"\n", values...)
	}
}

func (c *Config) Log2(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 2 {
		fmt.Printf(time.Now().Format("02/01/2006 – 15:04:05 SrvLog: ")+str+"\n", values...)
	}
}

func (c *Config) Log3(str string, values ...interface{}) {
	if c.LOG_LEVEL >= 3 {
		fmt.Printf(time.Now().Format("02/01/2006 – 15:04:05 SrvLog: ")+str+"\n", values...)
	}
}
