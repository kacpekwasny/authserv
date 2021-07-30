package authserv2

import (
	"database/sql"
	"time"
)

func (m *Manager) AddAccount(acc *Account) {
	m.InsertAccountDB(acc)
	m.Log2("Manger.addAccountDB(%v)", acc.Login)
	m.AppendAccountCache(acc)
}

func (m *Manager) RemoveAccount(login string) {
	m.DeleteAccountDB(login)
	m.DeleteAccountCache(login)
	m.Log2("Manger.RemoveAccount(%v)", login)
}

func (m *Manager) UpdateLogin(login, new_login string) error {
	err := m.LoadAccountToCache(login)
	switch err {
	case nil:
		m.UpdateValueDB(login, "login", new_login)
		m.GetAccountCache(login).Login = login
		return nil
	case sql.ErrNoRows:
		return ErrLoginNotFound
	default:
		m.Log1("m.UpdateLogin() err:", err.Error())
		if !m.connection {
			return ErrDBoffline
		}
		return err
	}
}

func (m *Manager) UpdatePassHash(login, new_pass_hash string) error {
	err := m.LoadAccountToCache(login)
	switch err {
	case nil:
		m.UpdateValueDB(login, "pass_hash", new_pass_hash)
		m.Log2("Manger.UpdatePassHash( %v )", login)
		m.GetAccountCache(login).Pass_hash = new_pass_hash
		return nil
	case sql.ErrNoRows:
		return ErrLoginNotFound
	default:
		m.Log1("m.UpdatePassHash() err:", err.Error())
		if !m.connection {
			return ErrDBoffline
		}
		return err
	}
}

func (m *Manager) UpdateToken(login, token string) error {
	err := m.LoadAccountToCache(login)
	switch err {
	case nil:
		m.UpdateValueDB(login, "token", token)
		m.Log2("Manger.UpdateToken( %v, xxxx... )", login)
		m.GetAccountCache(login).Current_token = token
		return nil
	case sql.ErrNoRows:
		return ErrLoginNotFound
	default:
		m.Log1("m.UpdateToken() err:", err.Error())
		if !m.connection {
			return ErrDBoffline
		}
		return err
	}
}

func (m *Manager) UpdateLastLogin2Now(login string) error {
	err := m.LoadAccountToCache(login)
	switch err {
	case nil:
		m.UpdateValueDB(login, "last_login", formatTime2TimeStamp(time.Now()))
		m.GetAccountCache(login).UpdateLogTime()
		m.Log2("Manger.UpdateLastLogin2NowDB( %v )", login)
		return nil
	case sql.ErrNoRows:
		return ErrLoginNotFound
	default:
		m.Log1("m.UpdateLastLogin2Now() err:", err.Error())
		if !m.connection {
			return ErrDBoffline
		}
		return err
	}
}

func (m *Manager) UpdateLoggedIn(login string, logged_in bool) error {
	err := m.LoadAccountToCache(login)
	switch err {
	case nil:
		m.UpdateValueDB(login, "logged_in", logged_in)
		m.GetAccountCache(login).Logged_in = logged_in
		m.Log2("Manger.UpdateLoggedIn( %v )", login)
		return nil
	case sql.ErrNoRows:
		return ErrLoginNotFound
	default:
		m.Log1("m.UpdateLoggedIn() err:", err.Error())
		if !m.connection {
			return ErrDBoffline
		}
		return err
	}
}

func (m *Manager) GetAccount(login string) (*Account, error) {
	acc := m.GetAccountCache(login)
	if acc != nil {
		return acc, nil
	}
	acc, err := m.GetAccountDB(login)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLoginNotFound
		}
		return nil, err
	}
	return acc, nil
}
