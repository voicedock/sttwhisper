# STT Whisper
Whisper.cpp based [VoiceDock STT](https://github.com/voicedock/voicedock-specs/blob/main/proto/voicedock/extensions/stt/v1/) implementation

> Provides gRPC API for high quality speech-to-text (from raw PCM stream) based on [Whisper.cpp](https://github.com/ggerganov/whisper.cpp) project.
> Provides download of new language packs via API.

# Usage
Build docker image:
```bash
docker build -t sttwhisper .
```
Run docker container:
```bash
docker run --rm \
  -v "$(pwd)/config:/data/config" \
  -v "$(pwd)/dataset:/data/dataset" \
  -p 9999:9999 \
  sttwhisper sttwhisper
```
## API
See implementation in [proto file](https://github.com/voicedock/voicedock-specs/blob/main/proto/voicedock/extensions/stt/v1/stt_api.proto).

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
Create protobuilder docker image:
```bash
cd ci/protobuilder && \
docker build -t protobuilder .
```
Lint proto files:
```bash
docker run --rm -w "/work" -v "$(pwd):/work" bufbuild/buf:latest lint internal/api/grpc/proto
```
Generate grpc interface:
```bash
docker run --rm -w "/work" -v "$(pwd):/work" protobuilder generate internal/api/grpc/proto --template internal/api/grpc/proto/buf.gen.yaml
```

## Thanks
* [Georgi Gerganov](https://github.com/ggerganov) - STT Whisper uses go binding [whisper.cpp](https://github.com/ggerganov/whisper.cpp)