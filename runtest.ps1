param(
    [string]$Mode = "sim",
    [string]$SimName = "2207BMSO",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ExtraArgs
)

Write-Host "Use CTRL C to cancel"
go clean -testcache
# go test -timeout 99999s
# go test -v

$commitId = (git rev-parse HEAD 2>$null)
if (-not $commitId) {
    $commitId = "unknown"
}

$env:FUZZER_PORT = "20260"
$env:FUZZER_HOST = "0.0.0.0"
$env:FUZZER_WEBPORT = "8080"
$env:FUZZER_COMMIT_ID = "$commitId"

if ($Mode -eq "demo") {
    $TAGS = "demo,web"
    $env:FUZZER_SIM_NAME = ""
    $testArgs = @("test", "-tags=$TAGS", "-v", "-run", "Test0", "-timeout", "105m")
} else {
    $TAGS = "sim,web"
    if ($Mode -like "-*") {
        $ExtraArgs = @($Mode) + $ExtraArgs
        $Mode = "sim"
    }
    if ($Mode -ne "sim") {
        $SimName = $Mode
    }
    if ($SimName -like "-*") {
        $ExtraArgs = @($SimName) + $ExtraArgs
        $SimName = "2207BMSO"
    }
    if (-not $SimName) {
        $SimName = "2207BMSO"
    }
    $env:FUZZER_SIM_NAME = $SimName
    $testArgs = @("test", "-tags=$TAGS", "-v", "-run", "TestSimFuzzer", "-timeout", "105m", "-sim=$SimName")
}
if ($ExtraArgs) {
    $testArgs += $ExtraArgs
}

go @testArgs
