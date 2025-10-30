```
iv: 5030511e51bbf22fb651936cbeeed884
key: 2814a9ddc9529f174206d5c030229e610cdac08da698003ab0a8f5fcad93a095
```

```console
$ CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -ldflags "-extldflags=-static" -o hello.dll ./hello.go
```

example powershell calls
```
[abc.abcabc]::NetcatSendTCP("127.0.0.1", 12345, "hello from pwsh", 200, 1) | Out-Null
[abc.abcabc]::PortScan("127.0.0.1", 1, 65535, 200, 1, "C:\scans\127.0.0.1_scan.txt")
```

## References
1. <https://github.com/shantanu561993/SharpChisel/tree/master/Go/chisel-64>
