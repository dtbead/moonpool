package cmd

func contains(set []string, value string) bool {
	for _, k := range set {
		if k == value {
			return true
		}
	}
	return false
}
