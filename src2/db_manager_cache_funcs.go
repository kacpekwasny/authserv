package authserv2

import (
	"time"
)

func (m *Manager) ObserveAccountCache() {
	//fmt.Println("ObserverAccBuff()")
	for {
		if len(m.accs_cache) > m.max_cached_accs {
			acc := m.PopLastAccountCache()
			if acc != nil {
				m.InsertAccountDB(acc)
			}
		}
		time.Sleep(m.observe_interval)
	}
}

func (m *Manager) ClearCache() {
	m.Log3("accounts cache was cleared")
	m.accs_cache = []*Account{}
}

// return nil if cache is empty
func (m *Manager) PopLastAccountCache() *Account {
	m.accs_cache_mtx.Lock()
	defer m.accs_cache_mtx.Unlock()
	acc_len := len(m.accs_cache)
	if acc_len > 0 {
		acc := m.accs_cache[acc_len-1]
		m.accs_cache = m.accs_cache[:acc_len-1]
		return acc
	}
	return nil
}

func (m *Manager) GetAccountCache(login string) *Account {
	//fmt.Println("getAccBuff()")
	m.accs_cache_mtx.Lock()
	defer m.accs_cache_mtx.Unlock()
	for _, acc := range m.accs_cache {
		if acc.Login == login {
			return acc
		}
	}
	return nil
}

func (m *Manager) AppendAccountCache(acc *Account) {
	if acc == nil {
		return
	}
	m.accs_cache_mtx.Lock()
	defer m.accs_cache_mtx.Unlock()
	m.accs_cache = append(m.accs_cache, acc)
}

func (m *Manager) AccountInCache(login string) bool {
	//fmt.Println("accIsInBuff()")
	for _, acc := range m.accs_cache {
		if acc.Login == login {
			return true
		}
	}
	return false
}

func (m *Manager) LoadAccountToCache(login string) error {
	if m.AccountInCache(login) {
		return nil
	}
	acc, err := m.GetAccountDB(login)
	if err != nil {
		return err
	}
	m.AppendAccountCache(acc)
	return nil
}

func (m *Manager) DeleteAccountCache(login string) {
	var (
		index  int
		found  bool
		length int = len(m.accs_cache)
	)
	m.accs_cache_mtx.Lock()
	for i, acc := range m.accs_cache {
		if acc.Login == login {
			found = true
			index = i
			break
		}
	}
	if !found {
		m.accs_cache_mtx.Unlock()
		return
	}
	if index < length-1 {
		m.accs_cache = append(m.accs_cache[:index], m.accs_cache[index+1:]...)
		m.accs_cache_mtx.Unlock()
		return
	}
	m.PopLastAccountCache()
	m.accs_cache_mtx.Unlock()
}
