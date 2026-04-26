@echo off
set GOOS=windows
set GOARCH=amd64

if "%1"=="--recompile" (
    echo Recompiling...
    go build -gcflags="all=-N -l" -o _output/es-cli.exe ./cmd/es-cli/
)

dlv exec _output/es-cli.exe --headless --listen=:2345 --api-version=2 --accept-multiclient
