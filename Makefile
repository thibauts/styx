PLATFORMS := linux darwin
VERSION := $(shell git describe --tags --abbrev=0)

release: $(PLATFORMS) docker-build

$(PLATFORMS):
	mkdir -p release
	GOOS=$@ GOARCH=amd64 CGO_ENABLED=0 go build -o release/styx-server ./cmd/styx-server
	GOOS=$@ GOARCH=amd64 CGO_ENABLED=0 go build -o release/styx ./cmd/styx
	mkdir -p release/data
	tar czf release/styx-$(VERSION)-$@-amd64.tar.gz release/styx-server release/styx release/data config.toml LICENSE
	rm -rf release/styx-server release/styx release/data

docker-build:
	docker build -t dataptive/styx:$(VERSION) -t dataptive/styx:latest .

docker-push:
	docker push dataptive/styx:$(VERSION)
	docker push dataptive/styx:latest

clean:
	rm -rf release

.PHONY: release clean $(PLATFORMS) docker-build docker-push
