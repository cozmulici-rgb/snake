param(
    [ValidateSet("console", "graphic", "graphic-fullscreen", "test", "build", "fmt")]
    [string]$Task = "console"
)

switch ($Task) {
    "console" {
        & go run .
    }
    "graphic" {
        & go run ./cmd/graphic
    }
    "graphic-fullscreen" {
        & go run ./cmd/graphic --fullscreen
    }
    "test" {
        & go test ./...
    }
    "build" {
        & go build ./...
    }
    "fmt" {
        & go fmt ./...
    }
}

exit $LASTEXITCODE
