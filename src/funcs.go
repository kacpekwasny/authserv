package authserv

import (
	"time"

	cmt "github.com/kacpekwasny/commontools"
)

func checkKeys(m map[string]string, keys ...string) bool {
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

func genToken() string {
	return cmt.RandString(TOKEN_LENGTH)
}
