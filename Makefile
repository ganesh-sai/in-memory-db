build:
	GOOS=linux go build -o bin/im-memory-db-linux main.go
	GOOS=darwin go build -o bin/im-memory-db-darwin main.go
	GOOS=windows go build -o bin/in-memory-db-windows.exe main.go
run:
	go run main.go


clean:
	rm -rf bin

