#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

IMG=af.hikvision.com.cn/docker-drpd/k8ss/prophet-webhook:3.6.1-VERSIONTAG.BUILDTIME
docker build . -t $IMG
docker push $IMG