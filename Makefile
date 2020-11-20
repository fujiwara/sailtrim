sailtrim: cmd/*/* *.go go.*
	go build -o sailtrim ./cmd/sailtrim/

test:
	go test -race .
