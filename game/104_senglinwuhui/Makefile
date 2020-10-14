#/bin/bash
# This is how we want to name the binary output
# 101 龙虎斗
101_LONGHUDOU_OUTPUT=../../bin/qp_104_slwh
101_LONGHUDOU_SRC=main.go


# These are the values we want to pass for Version and BuildTime
GITTAG=1.0.0
BUILD_TIME=`date +%Y%m%d%H%M%S`



# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.Version=${GITTAG} -X main.Build_Time=${BUILD_TIME} -s -w"

local: clean
	go build ${LDFLAGS} -o ${101_LONGHUDOU_OUTPUT}_mac ${101_LONGHUDOU_SRC}

.PHONY: proto
proto:
	protoc --proto_path=${GOPATH}/src:. --go_out=. msg/*.proto

.PHONY: linux
linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags=jsoniter ${LDFLAGS} -o ${101_LONGHUDOU_OUTPUT} ${101_LONGHUDOU_SRC}

.PHONY: clean
clean:
	rm -f ../../bin/qp_104_*
