package main

import (
	"fmt"
	grpcapi "github.com/voicedock/sttwhisper/internal/api/grpc"
	sttv1 "github.com/voicedock/sttwhisper/internal/api/grpc/gen/voicedock/extensions/stt/v1"
	"github.com/voicedock/sttwhisper/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:9999")
	if err != nil {
		fmt.Printf("failed to listen GRPC server: %s\n", err)
	}

	dataDir := "/data/dataset"
	dl := config.NewDownloader()
	cr := config.NewConfReader("/data/config/sttwhisper.json")
	dr := config.NewDataReader(dataDir)
	cs := config.NewService(cr, dr, dl, dataDir)
	cs.LoadConfig()

	srv := grpcapi.NewServerStt(cs)

	s := grpc.NewServer()
	sttv1.RegisterSttAPIServer(s, srv)
	reflection.Register(s)
	s.Serve(lis)
}
