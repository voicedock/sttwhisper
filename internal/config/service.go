package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

type Service struct {
	confReader      *ConfReader
	dataReader      *DataReader
	downloader      *Downloader
	config          []*LangPackData
	idxConfig       map[string]*LangPackData
	idxByNameConfig map[string]*LangPackData
	dataDir         string
}

func NewService(
	confReader *ConfReader,
	dataReader *DataReader,
	downloader *Downloader,
	dataDir string,
) *Service {
	return &Service{
		confReader:      confReader,
		dataReader:      dataReader,
		downloader:      downloader,
		config:          []*LangPackData{},
		idxConfig:       make(map[string]*LangPackData),
		idxByNameConfig: make(map[string]*LangPackData),
		dataDir:         dataDir,
	}
}

func (s *Service) LoadConfig() error {
	langConf, err := s.confReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed load config: %w", err)
	}

	var cfg []*LangPackData
	for _, v := range langConf {
		vData, _ := s.dataReader.ReadData(v)
		cfg = append(cfg, vData)
	}

	s.config = cfg

	s.RebuildIdx()
	return nil
}

func (s *Service) RebuildIdx() {
	for _, langData := range s.config {
		s.idxByNameConfig[langData.LangPack.Name] = langData
		for _, langCode := range langData.LangPack.Languages {
			s.idxConfig[langCode] = langData
		}
	}
}

func (s *Service) FindAll() []*LangPackData {
	return s.config
}

func (s *Service) Download(name string) error {
	langPack, ok := s.idxByNameConfig[name]
	if !ok {
		return errors.New("language pack not found")
	}

	if !langPack.Downloadable() {
		return errors.New("language pack is not downloadable")
	}

	dowloadUrl := filepath.Join(s.dataDir, langPack.LangPack.Name, "model.bin")
	err := s.downloader.Download(langPack.LangPack.DownloadUrl, dowloadUrl)
	if err != nil {
		return fmt.Errorf("faile download language pack: %w", err)
	}

	return s.LoadConfig()
}

func (s *Service) FindDownloaded(lang string) *LangPackData {
	ret := s.idxConfig[lang]
	if ret != nil && ret.Downloaded() {
		return ret
	}

	return nil
}
