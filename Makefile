PLATFORMS=linux darwin

release: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$@ GOARCH=amd64 CGO_ENABLED=0 go build -o styx-server ./cmd/styx-server
	GOOS=$@ GOARCH=amd64 CGO_ENABLED=0 go build -o styx ./cmd/styx
	mkdir data
	tar czf styx-$@-amd64.tar.gz styx-server styx data config.toml LICENSE
	rm -rf styx-server styx data

clean:
	rm styx-*-amd64.tar.gz

.PHONY: release clean $(PLATFORMS)
