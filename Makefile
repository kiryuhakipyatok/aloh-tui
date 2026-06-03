bs:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-s --output type=local,dest=./build_output_s .

bd:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-d --output type=local,dest=./build_output_d .

b: 
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-ds --output type=local,dest=./build .

lb:
	GOOS=linux CGO_ENABLED=1 GOARCH=amd64 \
	go build -tags pkgconfig,netgo,osusergo \
	-ldflags="-s -w -X 'main.version=1.0.0' -X 'main.env=prod' -extldflags '-L/usr/local/lib -Wl,-rpath=/usr/local/lib -lopusfile -lX11 -lxcb -lXau -lXdmcp -lspeexdsp -lrnnoise -lopus -logg -lm'" \
	-o aloh cmd/app/main.go