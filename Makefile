bs:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-s --output type=local,dest=./build_output_s .

bd:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-d --output type=local,dest=./build_output_d .

b: DOCKER_BUILDKIT=1 docker build -f Dockerfile --output type=local,dest=./build .

// b: bs bd