@echo off
setlocal

echo Use CTRL C to cancel
go clean -testcache
rem go test -timeout 99999s
rem go test -v

if /i "%~1"=="sim" (
    set "TAGS=sim,web"
) else (
    set "TAGS=demo,web"
)

set "FUZZER_COMMIT_ID="
for /f "tokens=*" %%i in ('git rev-parse HEAD 2^>nul') do set "FUZZER_COMMIT_ID=%%i"
if "%FUZZER_COMMIT_ID%"=="" (
    set "FUZZER_COMMIT_ID=unknown"
)

set "FUZZER_PORT=20260"
set "FUZZER_HOST=0.0.0.0"
set "FUZZER_SIM_NAME=%~2"
set "FUZZER_WEBPORT=8080"

go test -tags="%TAGS%" -v -run Test0 -timeout 105m %3 %4 %5 %6 %7 %8 %9
rem go test -v -run Test0 -timeout 105m
rem go test -v -tags="sim,web" -run Test1 -timeout 120m

endlocal
