bs:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-s --output type=local,dest=./build_output_s .

bd:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.cross-d --output type=local,dest=./build_output_d .

b: bs bd