test:
	go test ./...

build:
	go build ./cmd/tabled

docker:
	docker build -t filefrog/uta-tabled:latest .
	
	rm -rf docker/frontend/htdocs
	cp -a htdocs/ docker/frontend/htdocs
	docker build -t filefrog/uta-nginx:latest  -f docker/frontend/Dockerfile docker/frontend

release: docker
	docker push filefrog/uta-tabled:latest
	docker push filefrog/uta-nginx:latest

demo: docker
	docker-compose -p uta up -d

.PHONY: test build docker
