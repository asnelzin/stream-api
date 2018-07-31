docker:
	docker build -t asnelzin/stream-api .

tests:
	go test ./...
