LDFLAGS := -s -w

.PHONY: build
build:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/muxwaf-apiserver .


docker:
	#docker buildx build -t xnile/muxwaf-apiserver:v0.0.1 ./ --push
	docker buildx build -t xnile/muxwaf-apiserver:v0.0.1 ./

.PHONY: clean
clean:
	rm -rf ./bin