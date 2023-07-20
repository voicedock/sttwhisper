package grpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/voicedock/audio"
	"github.com/voicedock/sttwhisper/internal/aimodel"
	sttv1 "github.com/voicedock/sttwhisper/internal/api/grpc/gen/voicedock/core/stt/v1"
	"github.com/voicedock/sttwhisper/internal/config"
	"go.uber.org/zap"
	"io"
)

func NewServerStt(configService *config.Service, mm *aimodel.ModelManager, logger *zap.Logger) *ServerStt {
	return &ServerStt{
		configService: configService,
		mm:            mm,
		logger:        logger,
	}
}

type ServerStt struct {
	configService *config.Service
	mm            *aimodel.ModelManager
	logger        *zap.Logger
	sttv1.UnimplementedSttAPIServer
}

func (s *ServerStt) SpeechToText(srv sttv1.SttAPI_SpeechToTextServer) error {
	var cfg *config.ModelWrap
	var sampleRate int32
	var channels int32
	var lang string
	var buf []float32

	for {
		req, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read data: %w", err)
		}

		// load model
		if cfg == nil {
			cfg = s.configService.FindDownloaded(req.GetLang())
			if cfg == nil {
				return fmt.Errorf("model not found for language `%s`", req.GetLang())
			}

			lang = req.GetLang()
			sampleRate = req.GetAudio().SampleRate
			channels = req.GetAudio().Channels

			s.logger.Info("SpeechToText: starting",
				zap.String("lang", lang),
				zap.Int32("sampleRate", sampleRate),
				zap.Int32("channels", channels),
			)
		}

		if channels != 1 {
			return errors.New("single channel audio required")
		}

		if sampleRate < 16000 {
			return fmt.Errorf("invalid sample rate: required greater than or equal to 16000, %d passed", sampleRate)
		}

		s.logger.Debug("SpeechToText: recv audio bytes", zap.Int("len", len(req.Audio.Data)))
		r := new(bytes.Buffer)
		r.Write(req.Audio.Data)
		rawPcm := make([]int16, len(req.Audio.Data)/2)
		err = binary.Read(r, binary.LittleEndian, &rawPcm)
		if err != nil {
			return fmt.Errorf("unable to read pcm int16 LE format data from bytes: %w", err)
		}

		buf = append(buf, audio.ConvertPcmIntToFloat[float32](16, rawPcm)...)
	}

	mContext, err := s.mm.LoadModel(cfg, lang)
	if err != nil {
		return fmt.Errorf("failed to load model for language %s: %w", lang, err)
	}

	s.logger.Debug("SpeechToText: system info", zap.String("info", mContext.SystemInfo()))
	s.logger.Info("SpeechToText: analyze buf", zap.Int("len", len(buf)))

	err = mContext.Process(
		audio.DownsamplePcm[float32](buf, int(sampleRate), 16000),
		nil,
	)
	if err != nil {
		return fmt.Errorf("model processing error: %w", err)
	}
	defer mContext.ResetTimings()

	for {
		segment, err := mContext.NextSegment()
		if err != nil && err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("failed get next segment: %w", err)
		}

		for _, token := range segment.Tokens {
			if !mContext.IsText(token) {
				continue
			}

			s.logger.Info("SpeechToText: token", zap.String("text", token.Text))

			err := srv.Send(&sttv1.SpeechToTextResponse{
				TokenText:        token.Text,
				TokenProbability: token.P,
			})
			if err != nil {
				return fmt.Errorf("failed get next segment: %w", err)
			}
		}

		s.logger.Debug("SpeechToText: segment", zap.String("text", segment.Text))
	}

	return nil
}

func (s *ServerStt) GetLanguagePacks(ctx context.Context, in *sttv1.GetLanguagePacksRequest) (*sttv1.GetLanguagePacksResponse, error) {
	var langPacks []*sttv1.LanguagePack
	for _, v := range s.configService.FindAll() {
		langPacks = append(langPacks, &sttv1.LanguagePack{
			Name:         v.Model.Name,
			Languages:    v.Model.Languages,
			Downloaded:   v.Downloaded,
			Downloadable: v.Downloadable(),
			License:      v.Model.License,
		})
	}

	return &sttv1.GetLanguagePacksResponse{
		Languages: langPacks,
	}, nil
}

func (s *ServerStt) DownloadLanguagePack(ctx context.Context, in *sttv1.DownloadLanguagePackRequest) (*sttv1.DownloadLanguagePackResponse, error) {
	s.logger.Info("DownloadLanguagePack: starting", zap.String("name", in.Name))
	defer s.logger.Info("DownloadLanguagePack: complete", zap.String("name", in.Name))
	err := s.configService.Download(in.Name)

	return &sttv1.DownloadLanguagePackResponse{}, err
}
