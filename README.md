
# moonpool - self-hosted media tagging server

moonpool is a small media server intended to share and organize files between friends.

It organizes media based on a tagging system similar to many other booru services, though moonpool aims to be a higher performant alternative to [Hydrus](https://github.com/hydrusnetwork/hydrus/),
as well as a more easily accessible and self-hosted alternative to booru's such as [danbooru](https://github.com/danbooru/danbooru).

## Building
### Windows
1. Install MSYS2 and run the following commands in MINGW64  
 `pacman -S --needed base-devel mingw-w64-ucrt-x86_64-gcc`   
 `pacman -S mingw-w64-ucrt-x86_64-libwebp`  
 *you may need to add MSYS2 to your enviroment variables path (eg `C:\msys64\ucrt64\bin`)*
 1. Download repository
`git clone https://github.com/dtbead/moonpool && cd moonpool/cmd/moonpool`
 2. Download packages
 `go mod tidy`
 3. Enable CGO
 `go env -w CGO_ENABLED=1`
 4. Build Moonpool
 `go build`

### Linux
1. Install `libwebp` from your package manager
2. Download repository
`git clone https://github.com/dtbead/moonpool && cd moonpool/cmd/moonpool`
3. Download packages
`go mod tidy`
4. Build Moonpool
`go build`