package authserv

func checkKeys(m map[string]string, keys ...string) bool {
	for _, k := range keys {
		_, ok := m[k]
		if !ok {
			return false
		}
	}
	return true
}
