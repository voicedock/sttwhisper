package aimodel

import (
	"fmt"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/voicedock/sttwhisper/internal/config"
	"go.uber.org/zap"
)

type ModelManager struct {
	whisperMaxSegmentLen uint
	whisperMaxTokens     uint
	whisperThreads       uint
	model                whisper.Model
	modelPath            string
	logger               *zap.Logger
}

func NewModelManager(whisperMaxSegmentLen, whisperMaxTokens, whisperThreads uint, logger *zap.Logger) *ModelManager {
	return &ModelManager{
		whisperMaxSegmentLen: whisperMaxSegmentLen,
		whisperMaxTokens:     whisperMaxTokens,
		whisperThreads:       whisperThreads,
		logger:               logger,
	}
}

func (m *ModelManager) LoadModel(cfg *config.ModelWrap, lang string) (whisper.Context, error) {
	var err error
	if m.model != nil && m.modelPath != cfg.ModelPath {
		m.logger.Info("Unload old model")
		m.Unload()
	}

	if m.model == nil {
		m.logger.Info("Loading model", zap.String("path", cfg.ModelPath))
		m.model, err = whisper.New(cfg.ModelPath)
		if err != nil {
			m.logger.Error("Failed to load model", zap.Error(err))
			return nil, fmt.Errorf("failed to load model: %w", err)
		}

		m.logger.Info("Complete loading model", zap.Strings("languages", m.model.Languages()))
	}

	m.logger.Info("Create new model context", zap.String("lang", lang))
	mContext, err := m.model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed create model context: %w", err)
	}
	err = mContext.SetLanguage(lang)
	if err != nil {
		m.logger.Error("Failed to set language", zap.Error(err))
		return nil, fmt.Errorf("failed to set language: %w", err)
	}

	mContext.SetTranslate(false)
	mContext.SetMaxSegmentLength(m.whisperMaxSegmentLen)
	mContext.SetMaxTokensPerSegment(m.whisperMaxTokens)
	mContext.SetThreads(m.whisperThreads)

	return mContext, nil
}

func (m *ModelManager) Unload() {
	if m.model != nil {
		m.modelPath = ""
		if err := m.model.Close(); err != nil {
			m.logger.Warn("Failed to close model", zap.Error(err))
		}
		m.model = nil
	}
}
