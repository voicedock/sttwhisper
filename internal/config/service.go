package config

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type Service struct {
	confReader *Reader
	downloader *Downloader
	logger     *zap.Logger
	config     []*ModelWrap
	idxByLang  map[string]*ModelWrap
	idx        map[string]*ModelWrap
	dataDir    string
}

func NewService(
	confReader *Reader,
	downloader *Downloader,
	logger *zap.Logger,
	dataDir string,
) *Service {
	return &Service{
		confReader: confReader,
		downloader: downloader,
		logger:     logger,
		config:     []*ModelWrap{},
		idxByLang:  make(map[string]*ModelWrap),
		idx:        make(map[string]*ModelWrap),
		dataDir:    dataDir,
	}
}

func (s *Service) LoadConfig() error {
	items, err := s.confReader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	models := make([]*ModelWrap, 0, len(items))
	idx := map[string]*ModelWrap{}
	idxByLang := map[string]*ModelWrap{}
	for _, v := range items {
		if _, ok := idx[v.Name]; ok {
			s.logger.Warn("Model with same name skipped", zap.String("name", v.Name))
			continue
		}

		model := s.WrapModel(v)
		idx[v.Name] = model
		models = append(
			models,
			model,
		)

		idx[v.Name] = model
		for _, langCode := range v.Languages {
			idxByLang[langCode] = model
		}
	}

	s.config = models
	s.idx = idx
	s.idxByLang = idxByLang
	return nil
}

func (s *Service) WrapModel(model *Model) *ModelWrap {
	ret := &ModelWrap{
		Model:     model,
		ModelPath: filepath.Join(s.dataDir, model.Name, "model.bin"),
	}

	_, err := os.Stat(ret.ModelPath)
	ret.Downloaded = !os.IsNotExist(err)

	return ret
}

func (s *Service) FindAll() []*ModelWrap {
	return s.config
}

func (s *Service) Download(name string) error {
	langPack, ok := s.idx[name]
	if !ok {
		return errors.New("language pack not found")
	}

	if !langPack.Downloadable() {
		return errors.New("language pack is not downloadable")
	}

	err := s.downloader.Download(langPack.Model.DownloadUrl, langPack.ModelPath)
	if err != nil {
		return fmt.Errorf("failed to download language pack: %w", err)
	}

	return s.LoadConfig()
}

func (s *Service) FindDownloaded(lang string) *ModelWrap {
	ret := s.idxByLang[lang]
	if ret != nil && ret.Downloaded {
		return ret
	}

	return nil
}
