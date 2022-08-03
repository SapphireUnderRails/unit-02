package main

// Function to check if a key with a type of string is in a map.
func stringInKeys(key string, m map[string]int64) bool {
	for k := range m {
		if k == key {
			return true
		}
	}

	return false
}
