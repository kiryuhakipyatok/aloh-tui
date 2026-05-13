bs:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-s --output type=local,dest=./build_output_s .

bd:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-d --output type=local,dest=./build_output_d .

b: DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-ds --output type=local,dest=./build .