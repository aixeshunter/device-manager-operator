FROM af.hikvision.com.cn/docker-drpd/library/golang:1.13.3-tools as builder

WORKDIR /go/src/hikvision.com/cloud/device-manager
COPY .git/ .git/
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY vendor/ vendor/
COPY Makefile Makefile

# Build
RUN make build

FROM af.hikvision.com.cn/docker-drpd/library/alpine:10.4-device

WORKDIR /
COPY --from=builder /go/src/hikvision.com/cloud/device-manager/bin/device-manager .

ENTRYPOINT ["/device-manager"]