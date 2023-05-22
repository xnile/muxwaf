LDFLAGS := -s -w
VERSION := 0.0.1
GOOS ?= darwin
GOARCH ?= amd64
build-docker: build-ui build-docker-guard build-docker-apiserver build-docker-ui

build-apiserver:
	cd apiserver && env CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath -ldflags "$(LDFLAGS)" -o bin/muxwaf-apiserver .

build-apiserver-linux-amd64:
	cd apiserver && env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/muxwaf-apiserver .

build-ui:
	cd ui && yarn -i && yarn run build

build-docker-guard:
	docker buildx build -f docker/guard/Dockerfile -t xnile/muxwaf-guard:$(VERSION) ./

build-docker-apiserver:
	docker buildx build -f docker/apiserver/Dockerfile -t xnile/muxwaf-apiserver:$(VERSION) apiserver

build-docker-ui:
	docker buildx build -f docker/ui/Dockerfile -t xnile/muxwaf-ui:$(VERSION) ./

push-docker:
	docker push xnile/muxwaf-guard:$(VERSION)
	docker push xnile/muxwaf-apiserver:$(VERSION)
	docker push xnile/muxwaf-ui:$(VERSION)

run:
	docker-compose -f docker-compose.yml up
stop:
	docker-compose -f docker-compose.yml stop
rm:
	docker-compose -f docker-compose.yml rm -f

restart-guard:
	rm -rf /opt/apps/muxwaf/guard/logs/*.log
	brew services restart openresty

stop-guard:
	brew services stop openresty


.PHONY: clean
clean:
	rm -rf ./apiserver/bin
	rm -rf ./ui/dist