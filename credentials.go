package mysocks

var credentials = map[string]string{}

func addCredential(username, password string) {
	credentials[username] = password
}

func authenticate(username, password string) bool {
	if storedPassword, ok := credentials[username]; ok {
		return storedPassword == password
	}
	return false
}
