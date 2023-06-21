package grpc

import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	commonv1 "github.com/voicedock/sttwhisper/internal/api/grpc/gen/voicedock/extensions/common/v1"
	sttv1 "github.com/voicedock/sttwhisper/internal/api/grpc/gen/voicedock/extensions/stt/v1"
	"github.com/voicedock/sttwhisper/internal/config"
	"io"
	"math"
)

func NewServerStt(configService *config.Service) *ServerStt {
	return &ServerStt{
		configService: configService,
	}
}

type ServerStt struct {
	configService *config.Service
	sttv1.UnimplementedSttAPIServer
}

func (s *ServerStt) SpeechToText(srv sttv1.SttAPI_SpeechToTextServer) error {
	var modelPath string
	var sampleRate int32
	var codec commonv1.AudioCodec
	var channels int32
	var lang string

	var buf []float32
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed read data: %w", err)
		}

		// load model
		if modelPath == "" {
			langPack := s.configService.FindDownloaded(req.GetLang())
			if langPack == nil {
				return errors.New("lang not loaded")
			}

			lang = req.GetLang()
			modelPath = langPack.ModelPath
			sampleRate = req.GetData().SampleRate
			codec = req.GetData().Codec
			channels = req.GetData().Channels
			if channels != 1 {
				return errors.New("require one channels")
			}
			if codec == commonv1.AudioCodec_AUDIO_CODEC_INVALID {
				return errors.New("audio codec is not supported")
			}
		}

		if sampleRate < 16000 {
			return errors.New("incorrect sample rate: require 16000")
		}

		rawPcm := downsample(ConvertInts[float32](req.Data.Data), int(sampleRate))
		buf = append(buf, rawPcm...)
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		return errors.New("failed load language model")
	}

	defer model.Close()

	// Process samples
	context, err := model.NewContext()
	if err != nil {
		return errors.New("failed create model context")
	}
	context.SetLanguage(lang)
	context.SetTranslate(false)

	if err := context.Process(buf, nil); err != nil {
		return err
	}

	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}

		for _, token := range segment.Tokens {
			srv.Send(&sttv1.SpeechToTextResponse{
				TokenText:        token.Text,
				TokenProbability: token.P,
			})
		}
	}

	return nil
}

func (s *ServerStt) GetLanguagePacks(ctx context.Context, in *sttv1.GetLanguagePacksRequest) (*sttv1.GetLanguagePacksResponse, error) {
	var langPacks []*sttv1.LanguagePack
	for _, v := range s.configService.FindAll() {
		langPacks = append(langPacks, &sttv1.LanguagePack{
			Name:         v.LangPack.Name,
			Languages:    v.LangPack.Languages,
			Downloaded:   v.Downloaded(),
			Downloadable: v.Downloadable(),
			License:      v.LangPack.License,
		})
	}

	return &sttv1.GetLanguagePacksResponse{
		Languages: langPacks,
	}, nil
}

func (s *ServerStt) DownloadLanguagePack(ctx context.Context, in *sttv1.DownloadLanguagePackRequest) (*sttv1.DownloadLanguagePackResponse, error) {
	err := s.configService.Download(in.Name)

	return &sttv1.DownloadLanguagePackResponse{}, err
}

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

func ConvertInts[U, T Int](s []T) (out []U) {
	out = make([]U, len(s))
	for i := range s {
		out[i] = U(s[i])
	}
	return out
}

func downsample(pcmData []float32, sampleRate int) []float32 {
	// Downsample (required sampleRate equal 16000)
	sampleRateRatio := sampleRate / 16000
	if sampleRateRatio > 1 {
		newPcmData := make([]float32, len(pcmData)/sampleRateRatio)
		var offsetResult = 0
		var offsetBuffer = 0
		for offsetResult < len(newPcmData) {
			var nextOffsetBuffer = int(math.Round(float64(offsetResult+1) * float64(sampleRateRatio)))
			// Use average value of skipped samples
			var accum float32
			var count float32
			for i := offsetBuffer; i < nextOffsetBuffer && i < len(pcmData); i++ {
				accum += pcmData[i]
				count++
			}
			newPcmData[offsetResult] = accum / count
			// Or you can simply get rid of the skipped samples:
			// result[offsetResult] = buffer[nextOffsetBuffer];
			offsetResult++
			offsetBuffer = nextOffsetBuffer
		}

		return newPcmData
	}

	return pcmData
}
