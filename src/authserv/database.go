package authserv

import (
	"errors"
	"sync"
)

var (
	accountList = &[]*account{}
	mtx         = &sync.Mutex{}
)

// error if no account with a matching login
func getAccByLogin(login string) (*account, error) {
	for _, acc := range *accountList {
		if acc.Login == login {
			return acc, nil
		}
	}
	return nil, errors.New("no matching login")
}

func idIsUnique(id string) bool {
	for _, acc := range *accountList {
		if acc.Id == id {
			return false
		}
	}
	return true
}

func addAccount(acc *account) {
	mtx.Lock()
	*accountList = append(*accountList, acc)
	mtx.Unlock()
}

func removeAccount(login string) (bool, string) {
	acc, err := getAccByLogin(login)
	if err != nil {
		return false, "No account with such login"
	}
	mtx.Lock()
	var index int
	for i, acc_ := range *accountList {
		if acc.Id == acc_.Id {
			index = i
			break
		}
	}
	*accountList = append((*accountList)[:index], (*accountList)[index+1:]...)
	return true, ""
}
