# spinnaker-tools
* TODO: Support non-admin creds
* TODO: Switch to k8s bindings
* TODO: Lots of other tests and bugfixes and improvements

This is currently very rough and is essentially an MVP.  Needs a decent amount of refactoring, and we need to add more parameters so this can be fully automated.

Build

```bash
mkdir -p build
go get -u github.com/armory/spinnaker-tools
cd ~/go/src/github.com/armory/spinnaker-tools
GOOS=darwin GOARCH=amd64 go build -o build/spinnaker-tools-darwin
GOOS=linux GOARCH=amd64 go build -o build/spinnaker-tools-linux
```

[![asciicast](https://asciinema.org/a/5w3Tpygafe2cF8pB7R4OgtuBT.svg)](https://asciinema.org/a/5w3Tpygafe2cF8pB7R4OgtuBT)
