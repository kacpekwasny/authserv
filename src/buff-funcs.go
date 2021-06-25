package authserv

import "time"

// ALL Account atributes BUFFER UPDATE //
func (m *Manager) updateLoginBuff(login, newlogin string) bool {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false
	}
	acc.setLogin(newlogin)
	return true
}

func (m *Manager) updatePassHashBuff(login, newPass_hash string) bool {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false
	}
	acc.updatePassHash(newPass_hash)
	return true
}

func (m *Manager) updateLastLogin2NowBuff(login string) bool {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false
	}
	acc.uLogTime()
	return true
}

func (m *Manager) updateTokenBuff(login, newtoken string) bool {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false
	}
	acc.Current_token = newtoken
	return true
}

func (m *Manager) updateLoggedInBuff(login string, logged_in bool) bool {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false
	}
	acc.Logged_in = logged_in
	return true
}

// ALL Account atributes DATABASE GET //
func (m *Manager) getPassHashBuff(login string) (bool, string) {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false, ""
	}
	return true, acc.Pass_hash
}

func (m *Manager) getLastLoginBuff(login string) (bool, time.Time) {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false, time.Time{}
	}
	return true, acc.Last_login
}

func (m *Manager) getTokenBuff(login string) (bool, string) {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false, ""
	}
	return true, acc.Current_token
}

func (m *Manager) getLoggedInBuff(login string) (bool, bool) {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false, false
	}
	return true, acc.Logged_in
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ //
