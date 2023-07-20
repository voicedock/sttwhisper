FROM golang:1.20 as builder

RUN apt update && apt install -y \
        g++ \
        wget \
        git && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build

ADD . /usr/src/app

RUN wget https://github.com/ggerganov/whisper.cpp/archive/refs/tags/v1.4.0.tar.gz && \
    tar -xvf "v1.4.0.tar.gz" --strip-components 1 -C "./" && \
    cd bindings/go && \
    make whisper && \
    cd /usr/src/app && \
    C_INCLUDE_PATH=/build LIBRARY_PATH=/build go build -o ./sttwhisper ./cmd/sttwhisper

FROM debian:12

RUN apt update && \
    apt install -y ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/src/app/sttwhisper /usr/local/bin/sttwhisper

CMD ["sttwhisper"]