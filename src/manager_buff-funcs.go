package authserv

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
	acc.UpdatePassHash(newPass_hash)
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

func (m *Manager) updateLoggedInBuff(login string, logged_in bool) bool {
	succ, acc := m.getAccBuff(login)
	if !succ {
		return false
	}
	acc.Logged_in = logged_in
	return true
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ //
