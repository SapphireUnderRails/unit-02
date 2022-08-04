package main

// // Function to iterate over the array of users on cooldown
// // and to return a bool whether or not is in the list and
// // the index of the desired user if that user is in the array.
// func indexOfString(desiredKey string, arr []UserCooldown) (bool, int64) {
// 	// Iterate over the array.
// 	for index, key := range arr {
// 		// Check if the desired key matches the key at the current index.
// 		if key.userID == desiredKey {
// 			return true, int64(index)
// 		}
// 	}

// 	// We couldn't find a key that matches so we return -1
// 	// to signal that we could not find the desired key.
// 	return false, -1
// }

var (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)
