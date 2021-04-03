
###################################
#Build stage
FROM golang:1.15-alpine3.13 AS build-env

#Build deps
RUN apk --no-cache add git  

#Setup repo
COPY . /app/go-trader
WORKDIR /app/go-trader

#Checkout version if set
# RUN if [ -n "${GITEA_VERSION}" ]; then git checkout "${GITEA_VERSION}"; fi \
#  && make clean-all build
# RUN go mod download
RUN go build .

#FROM alpine:3.13
# FROM unity-registry.cn-shanghai.cr.aliyuncs.com/plasticscm/alpine:latest
# LABEL maintainer="maintainers@gitea.io"

EXPOSE 3001

#RUN apk --no-cache add \
#    bash \
#    curl \
#    gettext \
#    su-exec \
#    gnupg

# ENV GITEA_CUSTOM /data/gitea
# VOLUME ["/data"]

CMD ["/app/go-trader/go-trader", "test"]
