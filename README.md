# moonpool - self-hosted media tagging server
moonpool is a small program intended to share and organize files between friends. It organizes media based on a tagging system similar to many other booru services.
## Building
### Windows
 1. Download repository
`git clone https://github.com/dtbead/moonpool && cd moonpool/cmd/moonpool`
 2. Download packages
 `go mod tidy`
 3. Build moonpool
 `go build`

### Linux
1. Download repository
`git clone https://github.com/dtbead/moonpool && cd moonpool/cmd/moonpool`
2. Download packages
`go mod tidy`
3. Build moonpool
`go build`

## Using moonpool
run `./moonpool --help` to see all commands. As a quick start, use `./moonpool launch` to run the webUI.

## Notes
moonpool is currently in alpha and thus provides no guarantees to data integrity, nor software stability.