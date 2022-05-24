package utils

import "log"

// checks if the error is not nil, if its not, fatal and log
func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
