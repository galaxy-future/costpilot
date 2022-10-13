FROM golang:1.17-alpine as builder

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.4/main/" > /etc/apk/repositories && \
    apk add --no-cache \
    wget \
    git

RUN mkdir -p /home/tiger/build && \
    mkdir -p /home/tiger/app

ARG build_dir=/home/tiger/build
ARG app_dir=/home/tiger/app

ENV ServiceName=costpilot

WORKDIR $build_dir

COPY . .

# Cache dependencies
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct

COPY go.mod go.mod
COPY go.sum go.sum
#RUN  go mod download

RUN mkdir -p output/conf output/bin

# detect mysql start

RUN find conf/ -type f ! -name "*local*" -print0 | xargs -0 -I{} cp {} output/conf/ && \
    cp -rf website/ output/ && \
    cp scripts/run.sh output/

RUN CGO_ENABLED=0 GO111MODULE=on go build -o output/bin/${ServiceName} ./

RUN cp -rf output/* $app_dir

# --------------------------------------------------------------------------------- #
# Executable image
FROM alpine:3.14

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.4/main/" > /etc/apk/repositories

RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
ENV TZ Asia/Shanghai

RUN apk add --no-cache bash
ARG app_dir=/home/tiger/app
ENV ENV=docker
COPY --from=builder $app_dir $app_dir
WORKDIR $app_dir

CMD ["/bin/sh","/home/tiger/app/run.sh"]