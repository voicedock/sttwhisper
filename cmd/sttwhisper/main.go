package main

import (
	"github.com/alexflint/go-arg"
	"github.com/voicedock/sttwhisper/internal/aimodel"
	grpcapi "github.com/voicedock/sttwhisper/internal/api/grpc"
	sttv1 "github.com/voicedock/sttwhisper/internal/api/grpc/gen/voicedock/core/stt/v1"
	"github.com/voicedock/sttwhisper/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"runtime"
)

var cfg AppConfig
var logger *zap.Logger

func init() {
	arg.MustParse(&cfg)
	logger = initLogger(cfg.LogLevel, cfg.LogJson)
}

func main() {
	defer logger.Sync()

	if cfg.WhisperThreads == 0 {
		cfg.WhisperThreads = uint(runtime.NumCPU())
	}

	logger.Info(
		"Starting STT Whisper",
		zap.String("data_dir", cfg.DataDir),
		zap.String("config", cfg.Config),
		zap.Uint("whisper_threads", cfg.WhisperThreads),
		zap.Uint("whisper_max_tokens", cfg.WhisperMaxTokens),
		zap.Uint("whisper_max_segment_len", cfg.WhisperMaxSegmentLen),
	)

	lis, err := net.Listen("tcp", cfg.GrpcAddr)
	if err != nil {
		logger.Fatal("Failed to listen gRPC server", zap.Error(err))
	}

	dl := config.NewDownloader()
	cr := config.NewReader(cfg.Config)
	cs := config.NewService(cr, dl, logger, cfg.DataDir)
	mm := aimodel.NewModelManager(cfg.WhisperMaxSegmentLen, cfg.WhisperMaxTokens, cfg.WhisperThreads, logger)
	defer mm.Unload()
	err = cs.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	srv := grpcapi.NewServerStt(cs, mm, logger)

	s := grpc.NewServer()
	sttv1.RegisterSttAPIServer(s, srv)
	reflection.Register(s)

	logger.Info("gRPC server listen", zap.String("addr", cfg.GrpcAddr))
	err = s.Serve(lis)
	if err != nil {
		logger.Fatal("gRPC server error", zap.Error(err))
	}
}
