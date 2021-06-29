package authserv

func (m *Manager) addAccount(acc *account) (bool, string) {
	//fmt.Printf("addAccount(%v)\n", acc)
	succ, msg := m.insertAccDB(acc)
	m.Log2("Manger.addAccountDB(%v) -> succ: %v, msg: %v", acc.Login, succ, msg)
	if !succ {
		return false, msg
	}
	m.appendAccToBuff(acc)
	return true, msgSUCCESS
}

func (m *Manager) removeAccount(login string) (bool, string) {
	//fmt.Printf("removeAccount(%v)\n", login)
	succ, msg := m.deleteAccDB(login)
	m.Log2("Manger.removeAccountDB(%v) -> succ: %v, msg: %v", login, succ, msg)
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
	succ, msg := m.updateLoginDB(login, new_login)
	m.Log2("Manger.updateLoginDB( %v ,  %v ) -> succ: %v, msg: %v", login, new_login, succ, msg)
	if !succ {
		return false, msg
	}
	succ, msg = m.loadAccToBuff(login)
	if !succ {
		return false, msg
	}
	m.updateLoginBuff(login, new_login)
	return true, msgSUCCESS
}

func (m *Manager) updatePassHash(login, new_pass_hash string) (bool, string) {
	//fmt.Printf("updatePassHash(%v, %v)\n", login, new_pass_hash)
	succ, msg := m.updatePassHashDB(login, new_pass_hash)
	m.Log2("Manger.updatePassHashDB( %v ,  %v ) -> succ: %v, msg: %v", login, new_pass_hash, succ, msg)
	if !succ {
		return false, msg
	}
	m.updatePassHashBuff(login, new_pass_hash)
	return true, msgSUCCESS
}

func (m *Manager) updateLastLogin2Now(login string) (bool, string) {
	//fmt.Printf("updateLastLogin2Now(%v)\n", login)
	succ, msg := m.updateLastLogin2NowDB(login)
	m.Log2("Manger.updateLastLogin2NowDB( %v ) -> succ: %v, msg: %v", login, succ, msg)
	if !succ {
		return false, msg
	}
	succ, msg = m.loadAccToBuff(login)
	if !succ {
		return false, msg
	}
	m.updateLastLogin2NowBuff(login)
	return true, msgSUCCESS
}

func (m *Manager) updateLoggedIn(login string, logged_in bool) (bool, string) {
	//fmt.Printf("updateLoggedIn(%v, %v)\n", login, logged_in)
	succ, msg := m.updateLoggedInDB(login, logged_in)
	m.Log2("Manger.updateLoggedInDB( %v ,  %v ) -> succ: %v, msg: %v", login, logged_in, succ, msg)
	if !succ {
		return false, msg
	}
	succ, msg = m.loadAccToBuff(login)
	if !succ {
		return false, msg
	}
	m.updateLoggedInBuff(login, logged_in)
	return true, msgSUCCESS
}

// Get from buff or from DB //
//////////////////////////////

func (m *Manager) getAcc(login string) (bool, string, *account) {
	//fmt.Printf("getAcc(%v)\n", login)
	// success, msg, pass_hash
	succ, acc := m.getAccBuff(login)
	if succ {
		// jest w buff
		return true, msgSUCCESS, acc
	}
	// nie ma w buff
	succ, msg, acc := m.getAccDB(login)
	m.Log2("Manger.getAccDB( %v ) -> succ: %v, msg: %v", login, succ, msg)
	if succ {
		// jest w db
		succ, msg = m.loadAccToBuff(login)
		if !succ {
			// tym razem nie udalo sie zaladowac z buff
			return false, msg, nil
		}
		// juz jest w buff
		return true, msgSUCCESS, acc
	}
	// nie udalo sie: nie ma LUB byl blad
	return false, msg, nil
}
