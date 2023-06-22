CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/server
cd bin
docker build . -t md-server
