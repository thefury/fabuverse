
test:
	go test ./...

coverage:
	go test -coverpkg=./... ./...
