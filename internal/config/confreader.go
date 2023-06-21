package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfReader struct {
	path string
}

func NewConfReader(path string) *ConfReader {
	return &ConfReader{
		path: path,
	}
}

func (r *ConfReader) ReadConfig() ([]*LanguagePack, error) {
	var ret []*LanguagePack
	f, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed open language pack config: %w", err)
	}

	err = json.NewDecoder(f).Decode(&ret)
	if err != nil {
		return nil, fmt.Errorf("failed read language pack config: %w", err)
	}

	return ret, nil
}
