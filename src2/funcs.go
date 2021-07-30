package authserv2

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Take a string and return a hash of it
func MakeHash(pass string) string {
	// no err needed because cost is OK
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(hash)
}

func Log(str string, values ...interface{}) {
	if str[0] == '\n' {
		fmt.Printf("\n"+time.Now().Format("02/01/2006 – 15:04:05 ")+str[1:]+"\n", values...)
	} else {
		fmt.Printf(time.Now().Format("02/01/2006 – 15:04:05 ")+str+"\n", values...)
	}
}

func formatTime2TimeStamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
