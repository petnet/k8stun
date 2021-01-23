
.PHONY: build clean

dist:
	mkdir -p dist

dist/k8stun: dist
	go build -o $@ cmd/k8stun/main.go

build: dist/k8stun

clean:
	rm -fr dist

run:
	go run cmd/k8stun/main.go -config example/config.yaml

install: dist/k8stun
	mkdir -p ~/.local/bin
	cp dist/k8stun ~/.local/bin/k8stun