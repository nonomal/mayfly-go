ARG BASEIMAGES=m.daocloud.io/docker.io/alpine:3.20.2

FROM ${BASEIMAGES} AS builder
ARG TARGETARCH

ARG MAYFLY_GO_VERSION
ARG MAYFLY_GO_DIR_NAME=mayfly-go-linux-${TARGETARCH}
ARG MAYFLY_GO_URL=https://gitee.com/dromara/mayfly-go/releases/download/${MAYFLY_GO_VERSION}/${MAYFLY_GO_DIR_NAME}.zip

RUN apk add --no-cache wget unzip && \
    wget -cO mayfly-go.zip ${MAYFLY_GO_URL} && \
    unzip mayfly-go.zip && \
    cp -r ${MAYFLY_GO_DIR_NAME}/. /opt/ && \
    rm -rf mayfly-go.zip ${MAYFLY_GO_DIR_NAME}


FROM ${BASEIMAGES}

ARG TZ=Asia/Shanghai

# 安装必要的运行时依赖并创建非root用户
RUN apk add --no-cache ca-certificates tzdata && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone && \
    addgroup -g 1000 mayfly && \
    adduser -u 1000 -G mayfly -s /bin/sh -D mayfly

# 复制构建产物
COPY --from=builder /opt/ /mayfly-go/

# 设置工作目录和权限
WORKDIR /mayfly-go
RUN chown -R mayfly:mayfly /mayfly-go

# 切换到非root用户
USER mayfly

EXPOSE 18888

CMD ["./mayfly-go"]