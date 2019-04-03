# spinnaker-tools

Build

```bash
mkdir -p build
go get -u github.com/armory/spinnaker-tools
cd ~/home/go/src/github.com/armory/spinnaker-tools
GOOS=darwin GOARCH=amd64 go build -o build/spinnaker-tools-darwin
GOOS=linux GOARCH=amd64 go build -o build/spinnaker-tools-linux
```