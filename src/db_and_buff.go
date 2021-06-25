package authserv

import (
	"fmt"
	"time"
)

func (m *Manager) addAccount(acc *account) (bool, string) {
	fmt.Printf("addAccount(%v)\n", acc)
	succ, msg := m.insertAccDB(acc)
	if !succ {
		return false, msg
	}
	m.appendAccToBuff(acc)
	return true, msgSUCCESS
}

func (m *Manager) removeAccount(login string) (bool, string) {
	fmt.Printf("removeAccount(%v)\n", login)
	succ, msg := m.deleteAccDB(login)
	if !succ {
		return false, msg
	}
	m.deleteAccBuff(login)
	return true, msgSUCCESS
}

// Update in buff and in //
///////////////////////////

func (m *Manager) updateLogin(login, new_login string) (bool, string) {
	// return err, success, msg
	fmt.Printf("updateLogin(%v, %v)\n", login, new_login)
	succ, msg := m.updateLoginDB(login, new_login)
	if !succ {
		return false, msg
	}
	succ, msg = m.loadAccToBuff(login)
	if !succ {
		panic(msg)
	}
	m.updateLoginBuff(login, new_login)
	return true, msgSUCCESS
}

func (m *Manager) updatePassHash(login, new_pass_hash string) (bool, string) {
	fmt.Printf("updatePassHash(%v, %v)\n", login, new_pass_hash)
	succ, msg := m.updatePassHashDB(login, new_pass_hash)
	if !succ {
		return false, msg
	}
	m.updatePassHashBuff(login, new_pass_hash)
	return true, msgSUCCESS
}

func (m *Manager) updateLastLogin2Now(login string) (bool, string) {
	fmt.Printf("updateLastLogin2Now(%v)\n", login)
	succ, msg := m.updateLastLogin2NowDB(login)
	if !succ {
		return false, msg
	}
	succ, msg = m.loadAccToBuff(login)
	if !succ {
		panic(msg)
	}
	m.updateLastLogin2NowBuff(login)
	return true, msgSUCCESS
}

func (m *Manager) updateToken(login, new_token string) (bool, string) {
	fmt.Printf("updateToken(%v, %v...)\n", login, new_token[:10])
	succ, msg := m.updateTokenDB(login, new_token)
	if !succ {
		return false, msg
	}
	succ, msg = m.loadAccToBuff(login)
	if !succ {
		panic(msg)
	}
	m.updateTokenBuff(login, new_token)
	return true, msgSUCCESS
}

func (m *Manager) updateLoggedIn(login string, logged_in bool) (bool, string) {
	fmt.Printf("updateLoggedIn(%v, %v)\n", login, logged_in)
	succ, msg := m.updateLoggedInDB(login, logged_in)
	if !succ {
		return false, msg
	}
	succ, msg = m.loadAccToBuff(login)
	if !succ {
		panic(msg)
	}
	m.updateLoggedInBuff(login, logged_in)
	return true, msgSUCCESS
}

// Get from buff or from DB //
//////////////////////////////

func (m *Manager) getAcc(login string) (bool, string, *account) {
	fmt.Printf("getAcc(%v)\n", login)
	// success, msg, pass_hash
	succ, acc := m.getAccBuff(login)
	if succ {
		// jest w buff
		return true, msgSUCCESS, acc
	}
	// nie ma w buff
	succ, msg, acc := m.getAccDB(login)
	if succ {
		// jest w db
		succ, msg = m.loadAccToBuff(login)
		if !succ {
			// tym razem nie udalo sie zaladowac z buff
			panic(msg)
		}
		// juz jest w buff
		return true, msgSUCCESS, acc
	}
	// nie udalo sie: nie ma LUB byl blad
	return false, msg, nil
}

func (m *Manager) getPassHash(login string) (bool, string, string) {
	fmt.Printf("getPassHash(%v)\n", login)
	// success, msg, pass_hash
	succ, pass_hash := m.getPassHashBuff(login)
	if succ {
		// jest w buff
		return true, msgSUCCESS, pass_hash
	}
	// nie ma w buff
	succ, msg, pass_hash := m.getPassHashDB(login)
	if succ {
		// jest w db
		succ, msg = m.loadAccToBuff(login)
		if !succ {
			// nie udalo sie pobrac z db
			panic(msg)
		}
		//udalo sie pobrac z db
		return true, msgSUCCESS, pass_hash
	}
	// nie ma
	return false, msg, ""
}

func (m *Manager) getLastLogin(login string) (bool, string, time.Time) {
	fmt.Printf("getLastLogin(%v)\n", login)
	// success, msg, pass_hash
	succ, last_login := m.getLastLoginBuff(login)
	if succ {
		// jest w buff
		return true, msgSUCCESS, last_login
	}
	// nie ma w buff
	succ, msg, last_login := m.getLastLoginDB(login)
	if succ {
		// jest w db
		succ, msg = m.loadAccToBuff(login)
		if !succ {
			// nie udalo sie pobrac z db
			panic(msg)
		}
		// jest juz w buff
		return true, msgSUCCESS, last_login
	}
	// nie ma albo blad
	return false, msg, time.Time{}
}

func (m *Manager) getToken(login string) (bool, string, string) {
	fmt.Printf("getToken(%v)\n", login)
	// success, msg, pass_hash
	succ, token := m.getTokenBuff(login)
	if succ {
		// jest w buff
		return true, msgSUCCESS, token
	}
	// nie ma w buff
	succ, msg, token := m.getTokenDB(login)
	if succ {
		// jest w db
		succ, msg = m.loadAccToBuff(login)
		if !succ {
			// nie udalo sie pobrac z db
			panic(msg)
		}
		// udalo sie pobrac z db
		return true, msgSUCCESS, token
	}
	// nie ma albo blad
	return false, msg, ""
}

func (m *Manager) getLoggedIn(login string) (bool, string, bool) {
	fmt.Printf("getLoggedIn(%v)\n", login)
	// success, msg, pass_hash
	succ, logged_in := m.getLoggedInBuff(login)
	if succ {
		// jest w buff
		return true, msgSUCCESS, logged_in
	}
	// nie ma w buff
	succ, msg, logged_in := m.getLoggedInDB(login)
	if succ {
		// jest w db
		succ, msg = m.loadAccToBuff(login)
		if !succ {
			panic(msg)
		}
		// udalo sie pobrac
		return true, msgSUCCESS, logged_in
	}
	// nie ma albo blad
	return false, msg, false
}
