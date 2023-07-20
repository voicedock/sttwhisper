# STT Whisper
Whisper.cpp based [VoiceDock STT](https://github.com/voicedock/voicedock-specs/blob/main/proto/voicedock/core/stt/v1/) implementation

> Provides gRPC API for high quality speech-to-text (from raw PCM stream) based on [Whisper.cpp](https://github.com/ggerganov/whisper.cpp) project.
> Provides download of new language packs via API.

# Usage
Run docker container on CPU:
```bash
docker run --rm \
  -v "$(pwd)/config:/data/config" \
  -v "$(pwd)/dataset:/data/dataset" \
  -p 9999:9999 \
  ghcr.io/voicedock/sttwhisper:latest sttwhisper
```
Run docker container on GPU (Nvidia CUDA):
```bash
docker run --rm \
  -v "$(pwd)/config:/data/config" \
  -v "$(pwd)/dataset:/data/dataset" \
  -p 9999:9999 \
  ghcr.io/voicedock/sttwhisper:gpu sttwhisper
```
Tested on NVIDIA GeForce RTX 3090.

Show more options:
```bash
docker run --rm ghcr.io/voicedock/sttwhisper sttwhisper -h
```
```
Usage: sttwhisper [--grpcaddr GRPCADDR] [--config CONFIG] [--datadir DATADIR] [--loglevel LOGLEVEL] [--logjson] [--whispermaxsegmentlen WHISPERMAXSEGMENTLEN] [--whispermaxtokens WHISPERMAXTOKENS] [--whisperthreads WHISPERTHREADS]

Options:
  --grpcaddr GRPCADDR    gRPC API host:port [default: 0.0.0.0:9999, env: GRPC_ADDR]
  --config CONFIG        configuration file for models [default: /data/config/sttwhisper.json, env: CONFIG]
  --datadir DATADIR      dataset directory [default: /data/dataset, env: DATA_DIR]
  --loglevel LOGLEVEL    log level: debug, info, warn, error, dpanic, panic, fatal [default: info, env: LOG_LEVEL]
  --logjson              set to true to use JSON format [env: LOG_JSON]
  --whispermaxsegmentlen WHISPERMAXSEGMENTLEN
                         maximum segment length in characters (0 = no limit) [env: WHISPER_MAX_SEGMENT_LEN]
  --whispermaxtokens WHISPERMAXTOKENS
                         maximum tokens per segment (0 = no limit) [env: WHISPER_MAX_TOKENS]
  --whisperthreads WHISPERTHREADS
                         number of threads to use during computation (0 = MAX) [env: WHISPER_THREADS]
  --help, -h             display this help and exit
```

## API
See implementation in [proto file](https://github.com/voicedock/voicedock-specs/blob/main/proto/voicedock/core/stt/v1/stt_api.proto).

## FAQ
### How to add a language pack?
1. Find model from [Hugging Face](https://huggingface.co/ggerganov/whisper.cpp/tree/main)
2. Copy link to download `.bin` model file
3. Add model to [sttwhisper.json](config%2Fsttwhisper.json) config:
   ```json
   {
     "name": "model_name",
     "languages": ["ru", "en"],
     "download_url": "download_url",
     "license": "license text to accept"
   }
    ```

### How to use preloaded model?
1. Add voice to [sttwhisper.json](config%2Fsttwhisper.json) config (leave "download_url" blank to disable downloads).
2. [Download](https://huggingface.co/ggerganov/whisper.cpp/tree/main)
3. Save model to directory `dataset/{model_name}/model.bin` (replace `{model_name}` to name from configuration file `sttwhisper.json`)


## CONTRIBUTING
Lint proto files:
```bash
docker run --rm -w "/work" -v "$(pwd):/work" bufbuild/buf:latest lint internal/api/grpc/proto
```
Generate grpc interface:
```bash
docker run --rm -w "/work" -v "$(pwd):/work" ghcr.io/voicedock/protobuilder:1.0.0 generate internal/api/grpc/proto --template internal/api/grpc/proto/buf.gen.yaml
```
Manual build CPU docker image:
```bash
docker build -t ghcr.io/voicedock/sttwhisper:latest .
```
Manual build GPU docker image:
```bash
docker build -t ghcr.io/voicedock/sttwhisper:gpu -f ./gpu.Dockerfile .
```

## Thanks
* [Georgi Gerganov](https://github.com/ggerganov) - STT Whisper uses go binding [whisper.cpp](https://github.com/ggerganov/whisper.cpp)