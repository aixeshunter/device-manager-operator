FROM af.hikvision.com.cn/docker-drpd/library/alpine:3.10.4


RUN apk add --no-cache \
    util-linux \
    xfsprogs \
    e2fsprogs \
    openssh \
    rm -rf /var/cache/apk/*


# build command
# docker build --network=host -t af.hikvision.com.cn/docker-drpd/library/alpine:10.4-device .