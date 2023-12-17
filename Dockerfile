FROM golang:1.21.5-alpine as builder
# ARG URL=https://github.com/crochee/${Project}.git
ARG GOSU_VERSION=1.17
WORKDIR /workspace
# 下载git 修改配置
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories &&\
    apk add --no-cache git tzdata
# 设置代理环境变量
RUN go env -w GOPROXY=https://goproxy.io,https://goproxy.cn,direct &&\
    go env -w GO111MODULE=on
# 代码拷贝
RUN git clone -b master https://github.com/crochee/iam.git
# 代码编译
RUN cd iam && go mod tidy &&\
    GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o iam -tags jsoniter ./cmd/iam &&\
    go install github.com/tianon/gosu@${GOSU_VERSION}
# 整理项目需要拷贝的资源
RUN mv ./iam/entrypoint.sh . &&\
    mv ./iam/config . &&\
    mkdir ./out &&\
    cp ./iam/iam ./out/ &&\
    cp ${GOPATH}/bin/gosu .

FROM alpine:latest as runner
ARG WorkDir=/opt/cloud
WORKDIR ${WorkDir}
# add our user and group first to make sure their IDs get assigned consistently, regardless of whatever dependencies get added
RUN addgroup -g 10000 cloud && adduser -g cloud dev -u 5000 -D -H
# 预创建文件夹
RUN mkdir -p ${WorkDir}/conf ${WorkDir}/log
# 资源拷贝
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /workspace/gosu /usr/local/bin/
COPY --from=builder /workspace/out/iam /usr/local/bin/
COPY --from=builder /workspace/entrypoint.sh /usr/local/bin/
COPY --from=builder /workspace/config ${WorkDir}/config
# 赋予执行权限
RUN chmod +x /usr/local/bin/iam /usr/local/bin/entrypoint.sh /usr/local/bin/gosu
# 将工作目录加入用户组
RUN chown -R  cloud:dev ${WorkDir}
# 日志文件夹0744
RUN chmod u=rwx,g=r,o=r ${WorkDir}/log
# 配置文件目录和文件0440,只有读权限
RUN chown -R root:root ${WorkDir}/config &&\
    chmod -R a+r,o-wx ${WorkDir}/config

EXPOSE 31000
STOPSIGNAL 2

ENTRYPOINT ["entrypoint.sh"]
CMD ["iam"]
