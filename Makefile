
build:
	CGO_ENABLED=0 go build -v -o .build/yclient ./yclient/cmd/yclient.go

test:
	go test -v ./yclient/...

serve:
	caddy fmt --overwrite
	caddy adapt
	caddy run
