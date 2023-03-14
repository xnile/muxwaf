FROM golang:1.18 AS build

ENV GOPROXY=https://goproxy.cn,direct
COPY . /build
WORKDIR /build

RUN make build

FROM alpine:3

#RUN mkdir /etc/muxwaf/

COPY --from=build /build/bin/muxwaf-apiserver /usr/bin/muxwaf-apiserver

ENTRYPOINT ["/usr/bin/muxwaf-apiserver", "-c", "/etc/muxwaf/config.yaml"]
