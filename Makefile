sailtrim: cmd/*/* *.go go.*
	go build -o sailtrim ./cmd/sailtrim/

test:
	go test -race ./...

install: sailtrim
	install sailtrim ~/bin

snapshot:
	goreleaser build --snapshot --rm-dist
