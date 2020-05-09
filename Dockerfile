FROM golang:1.13.1 as builder
# https://askubuntu.com/questions/420784/what-do-the-disabled-login-and-gecos-options-of-adduser-command-stand
RUN adduser --disabled-login --gecos "" citrixuser
COPY . $GOPATH/src/labels-db
WORKDIR $GOPATH/src/labels-db
RUN GOPROXY=direct GOSUMDB=off GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go install -ldflags "-extldflags -static -s -w" labels-db

FROM alpine

COPY --from=builder /go/bin/labels-db /go/bin/labels-db
COPY --from=builder /etc/passwd /etc/passwd

USER citrixuser
ENTRYPOINT ["/go/bin/labels-db"]