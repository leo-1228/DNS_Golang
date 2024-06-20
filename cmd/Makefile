arm:
	GOOS=linux GOARCH=arm go build -o dnscheck.arm.out main.go

linux:
	GOOS=linux GOARCH=amd64 go build -o dnscheck.linux.x64.out main.go

mac:
	GOOS=darwin GOARCH=amd64 go build -o dnscheck.mac.x64.out main.go

all: arm linux mac