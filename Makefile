export GO15VENDOREXPERIMENT=1

exe = ./cmd/cast
buildargs = -ldflags '-w -s -X github.com/barnybug/go-cast.Version=${TRAVIS_TAG}'

.PHONY: all build install test coverage release upx

all: install

test:
	go test . ./api/... ./cmd/... ./controllers/... ./discovery/... ./events/... ./log/... ./net/... 

build:
	go build -i -v $(exe)

install:
	go install $(exe)

release:
	GOOS=linux GOARCH=amd64 go build $(buildargs) -o release/cast-linux-amd64 $(exe)
	GOOS=linux GOARCH=386 go build $(buildargs) -o release/cast-linux-386 $(exe)
	GOOS=linux GOARCH=arm go build $(buildargs) -o release/cast-linux-arm $(exe)
	GOOS=darwin GOARCH=amd64 go build $(buildargs) -o release/cast-mac-amd64 $(exe)
	GOOS=windows GOARCH=386 go build $(buildargs) -o release/cast-windows-386.exe $(exe)
	GOOS=windows GOARCH=amd64 go build $(buildargs) -o release/cast-windows-amd64.exe $(exe)
	goupx release/cast-linux-amd64
	upx release/cast-linux-386 release/cast-linux-arm release/cast-windows-386.exe

upx:
	upx dist/go-cast-linux-386/cast-linux-386 dist/go-cast-linux-amd64/cast-linux-amd64 dist/go-cast-linux-arm/cast-linux-arm dist/go-cast-windows-386/cast-windows-386.exe
