package tools

import "os"

func CreateOrOpenFile(fileName string) (bool, *os.File, error) {
	// check if file exists
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		// create file if it doesn't exist
		f, e := os.Create(fileName)
		return true, f, e
	} else if err != nil {
		// return error if any other error occurred
		return false, nil, err
	}

	// open file without truncating if it already exists
	f, e := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	return false, f, e
	// return os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
}
