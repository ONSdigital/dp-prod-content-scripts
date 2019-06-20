package content

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var LimitReached = errors.New("limit reached")

type PageType struct {
	Value string `json:"type"`
}

func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadJson(path string) (b []byte, err error) {
	b, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GetPageType(b []byte) (*PageType, error) {
	var page PageType
	if err := json.Unmarshal(b, &page); err != nil {
		return nil, err
	}
	return &page, nil

}
