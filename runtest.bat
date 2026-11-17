@echo off
setlocal

echo Use CTRL C to cancel
go clean -testcache
rem go test -timeout 99999s
rem go test -v

set "FUZZER_COMMIT_ID="
for /f "tokens=*" %%i in ('git rev-parse HEAD 2^>nul') do set "FUZZER_COMMIT_ID=%%i"
if "%FUZZER_COMMIT_ID%"=="" (
    set "FUZZER_COMMIT_ID=unknown"
)

set "FUZZER_PORT=20260"
set "FUZZER_HOST=0.0.0.0"
set "FUZZER_WEBPORT=8080"

if /i "%~1"=="demo" goto run_demo
if /i "%~1"=="sim" goto run_sim_explicit

set "SIM_NAME=2207BMSO"
set "ARG1=%~1"
if "%ARG1%"=="" goto run_sim_default
if "%ARG1:~0,1%"=="-" goto run_sim_flags

set "TAGS=sim,web"
set "SIM_NAME=%~1"
set "FUZZER_SIM_NAME=%SIM_NAME%"
go test -tags="%TAGS%" -v -run TestSimFuzzer -timeout 105m -sim="%SIM_NAME%" %2 %3 %4 %5 %6 %7 %8 %9
goto done

:run_sim_flags
set "TAGS=sim,web"
set "FUZZER_SIM_NAME=%SIM_NAME%"
go test -tags="%TAGS%" -v -run TestSimFuzzer -timeout 105m -sim="%SIM_NAME%" %1 %2 %3 %4 %5 %6 %7 %8 %9
goto done

:run_sim_default
set "TAGS=sim,web"
set "FUZZER_SIM_NAME=%SIM_NAME%"
go test -tags="%TAGS%" -v -run TestSimFuzzer -timeout 105m -sim="%SIM_NAME%"
goto done

:run_sim_explicit
set "TAGS=sim,web"
set "SIM_NAME=%~2"
if "%SIM_NAME%"=="" set "SIM_NAME=2207BMSO"
if "%SIM_NAME:~0,1%"=="-" goto run_sim_explicit_flags
set "FUZZER_SIM_NAME=%SIM_NAME%"
go test -tags="%TAGS%" -v -run TestSimFuzzer -timeout 105m -sim="%SIM_NAME%" %3 %4 %5 %6 %7 %8 %9
goto done

:run_sim_explicit_flags
set "SIM_NAME=2207BMSO"
set "FUZZER_SIM_NAME=%SIM_NAME%"
go test -tags="%TAGS%" -v -run TestSimFuzzer -timeout 105m -sim="%SIM_NAME%" %2 %3 %4 %5 %6 %7 %8 %9
goto done

:run_demo
set "TAGS=demo,web"
set "FUZZER_SIM_NAME="
go test -tags="%TAGS%" -v -run Test0 -timeout 105m %2 %3 %4 %5 %6 %7 %8 %9

:done
rem go test -v -tags="sim,web" -run TestSimFuzzer -timeout 105m -sim=2207BMSO

endlocal
