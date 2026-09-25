param(
    [string]$Mode = "demo",
    [string]$SimName = "",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ExtraArgs
)

Write-Host "Use CTRL C to cancel"
go clean -testcache
# go test -timeout 99999s
# go test -v

if ($Mode -eq "sim") {
    $TAGS = "sim,web"
} else {
    $TAGS = "demo,web"
}

$commitId = (git rev-parse HEAD 2>$null)
if (-not $commitId) {
    $commitId = "unknown"
}

$env:FUZZER_PORT = "20260"
$env:FUZZER_HOST = "0.0.0.0"
$env:FUZZER_SIM_NAME = $SimName
$env:FUZZER_WEBPORT = "8080"
$env:FUZZER_COMMIT_ID = "$commitId"

$testArgs = @("test", "-tags=$TAGS", "-v", "-run", "Test0", "-timeout", "105m")
if ($ExtraArgs) {
    $testArgs += $ExtraArgs
}

go @testArgs
