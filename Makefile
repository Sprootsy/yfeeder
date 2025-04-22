
build:
	CGO_ENABLED=0 go build -v -o .build/yclient ./yclient/cmd/yclient.go

test:
	go test -v ./yclient/...

serve:
	caddy file-server --listen :2015 --root ./website
