
APP = bin/server_exporter

docker-build:
	@docker run -it --rm \
	  -e GO111MODULE=on \
	  --entrypoint "/bin/bash" \
	  -v ${PWD}:/src \
	  techknowlogick/xgo \
	  -c "cd /src && make clean && make build-all"

build-amd64:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
	 go build \
	   -trimpath \
	   -mod vendor \
	   -ldflags '-s -w ' \
	   -o ${APP}_amd64 server_exporter.go

build-arm64:
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc-6\
	 go build \
	   -trimpath \
	   -mod vendor \
	   -ldflags '-s -w ' \
	   -o ${APP}_arm64 server_exporter.go

build-all: build-amd64 build-arm64

clean:
	rm -f bin/server_exporter_*

debug: build-amd64
	@sudo SE_LOG=DEBUG ${APP}_amd64 --collector.sensor.state.path="bin/sensor.statelog"