drive digger : Scans all files, even the hidden  🤫  ones so that you can organize them xD

## Build for Win 
```
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export GOOS=windows
export GOARCH=amd64
go build -o yourprogram.exe
```

