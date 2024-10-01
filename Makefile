build-image:
	docker build -t pgs3:latest .

build:
	CGO_ENABLED=0 go build -o pgs3 .