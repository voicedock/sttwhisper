ARG UBUNTU_VERSION=22.04
# This needs to generally match the container host's environment.
ARG CUDA_VERSION=11.7.1
# Target the CUDA build image
ARG BASE_CUDA_DEV_CONTAINER=nvidia/cuda:${CUDA_VERSION}-devel-ubuntu${UBUNTU_VERSION}
# Target the CUDA runtime image
ARG BASE_CUDA_RUN_CONTAINER=nvidia/cuda:${CUDA_VERSION}-runtime-ubuntu${UBUNTU_VERSION}

FROM ${BASE_CUDA_DEV_CONTAINER} as builder

RUN apt update && apt install -y \
        g++ \
        wget \
        git && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build

ADD . /usr/src/app

ENV PATH $PATH:/usr/local/go/bin
ENV CUDA_DOCKER_ARCH=all

RUN wget https://go.dev/dl/go1.19.10.linux-amd64.tar.gz && \
    tar -xvf "go1.19.10.linux-amd64.tar.gz" -C /usr/local && \
    wget https://github.com/ggerganov/whisper.cpp/archive/refs/tags/v1.4.2.tar.gz && \
    tar -xvf "v1.4.2.tar.gz" --strip-components 1 -C "./"

RUN WHISPER_CUBLAS=1 make libwhisper.a

RUN cd /usr/src/app && \
    CGO_CFLAGS="-I/build -DGGML_USE_CUBLAS -I/usr/local/cuda/include -I/opt/cuda/include -I/targets/x86_64-linux/include" \
    CGO_CXXFLAGS="-I/build -I/build/examples -DGGML_USE_CUBLAS -I/usr/local/cuda/include -I/opt/cuda/include -I/targets/x86_64-linux/include" \
    CGO_LDFLAGS="-lwhisper -lcublas -lculibos -lcudart -lcublasLt -L/build -L/usr/local/cuda/lib64 -L/opt/cuda/lib64 -L/targets/x86_64-linux/lib" \
    go build -o ./sttwhisper ./cmd/sttwhisper

FROM ${BASE_CUDA_RUN_CONTAINER}

RUN apt update && \
    apt install -y ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/src/app/sttwhisper /usr/local/bin/sttwhisper

CMD ["sttwhisper"]