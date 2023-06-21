package config

import (
	"os"
	"path/filepath"
)

type DataReader struct {
	dataDir string
}

func NewDataReader(dataDir string) *DataReader {
	return &DataReader{
		dataDir: dataDir,
	}
}

func (d *DataReader) ReadData(langPack *LanguagePack) (*LangPackData, error) {
	ret := &LangPackData{
		LangPack: langPack,
	}
	dataPath := filepath.Join(d.dataDir, langPack.Name, "model.bin")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return ret, nil
	}

	ret.ModelPath = dataPath

	return ret, nil
}
