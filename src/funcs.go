package authserv

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func checkKeys(m map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		_, ok := m[k]
		if !ok {
			return false
		}
	}
	return true
}

func formatTime2TimeStamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// Take a string and return a hash of it
func makeHash(pass string) string {

	// no err needed because cost is OK
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	return string(hash)
}

// pass_hash is the hash from database and
// pass is the one sent by user
func passIsCorrect(pass, pass_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(pass_hash), []byte(pass))
	return err == nil
}

// json from Request to map[string]string
func json2map(r *http.Request) (map[string]string, error) {
	// read data that will be used to create an account
	reqBody, _ := ioutil.ReadAll(r.Body)
	m := &map[string]string{}
	err := json.Unmarshal(reqBody, m)
	return *m, err
}

// convert slice float to slice integer
func sIntf2sInteger(intf []interface{}) []int {
	si := []int{}
	for _, val := range intf {
		si = append(si, int(val.(float64)))
	}
	return si
}

// convert map[string]string to map[string]interface
// but only these entries that are string:string
func mS2mI(ms map[string]string) map[string]interface{} {
	mi := map[string]interface{}{}
	for k, v := range ms {
		mi[k] = v
	}
	return mi
}

func IntInSlice(i int, s []int) bool {
	for _, num := range s {
		if i == num {
			return true
		}
	}
	return false
}
