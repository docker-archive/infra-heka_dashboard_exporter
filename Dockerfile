FROM       alpine:3.1
MAINTAINER Aaron 'Sweet River' Vinson <avinson@docker.com>
EXPOSE     9111
ENTRYPOINT [ "/bin/heka_exporter" ]

ENV  GOPATH  /go
ENV  APPPATH $GOPATH/src/github.com/docker-infra/heka_exporter
COPY . $APPPATH
RUN  apk add --update -t build-deps go git mercurial \
     && cd $APPPATH && go get -d && go build -o /bin/heka_exporter \
     && apk del --purge build-deps && rm -rf $GOPATH
