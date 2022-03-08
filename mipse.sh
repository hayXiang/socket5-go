GOOS=linux
GOARCH=mipsle 
GOMIPS=softfloat
output=xsocks5-$GOARCH
go build -ldflags "-s -w" -o $output
upx -9 $output
