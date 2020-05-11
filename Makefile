docker:
	DOCKER_BUILDKIT=1 docker build -t dheerajng/labels-db .

docker-run:
	(docker rm -f podlabels-db) || true
	docker run --name labels-db -d \
	-p 8080:8080 \
	-e DEBUG=true \
	citrix/labels-db

run:
	DEBUG=true GOPROXY=direct GOSUMDB=off go run main.go
