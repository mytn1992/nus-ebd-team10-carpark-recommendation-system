package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func OpenFile(path string) ([]byte, *os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error while openning the file (%v): %v", path, err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("error while reading the file (%v): %v", path, err)
	}

	return b, file, nil
}

func OpenJSONFile(path string, dest interface{}) error {
	b, _, err := OpenFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, dest)
	if err != nil {
		return fmt.Errorf("error while parsing the content (%v): %v", path, err)
	}
	return nil
}
