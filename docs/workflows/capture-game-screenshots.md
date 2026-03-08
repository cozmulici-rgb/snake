# Capture Game Screenshots

This is the project memory for capturing browser-game screenshots reliably in this repo.

## When to use this

Use this after changing the web UI in:

- `internal/ui/web/static/index.html`
- `internal/ui/web/static/app.js`

## Key constraints

- The web frontend is embedded by Go. After changing static files, restart `go run ./cmd/web` or you will capture stale assets.
- PowerShell script execution may block `npx.ps1`. Use `npx.cmd` and `npm.cmd` instead.
- The Playwright client script imports `playwright` as a module from the skill directory. If it is missing, install it into the skill path, not the repo.

## One-time Playwright fix

```powershell
npm.cmd install --prefix C:\Users\cozmu\.codex\skills\develop-web-game --no-save --no-package-lock playwright
```

## Start or restart the web server

Kill the current listener on `127.0.0.1:8080`, then relaunch the embedded web UI:

```powershell
$conn = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
if ($conn) { Stop-Process -Id $conn.OwningProcess -Force }
cmd /c "start /b go run ./cmd/web > output\web-server.out.log 2> output\web-server.err.log"
Start-Sleep -Seconds 5
Invoke-WebRequest -UseBasicParsing http://127.0.0.1:8080/api/state | Select-Object -ExpandProperty StatusCode
```

## Create simple action payloads

```powershell
@'
{
  "steps": [
    { "buttons": [], "frames": 2 }
  ]
}
'@ | Set-Content output/start-actions.json

@'
{
  "steps": [
    { "buttons": ["right"], "frames": 4 },
    { "buttons": [], "frames": 8 }
  ]
}
'@ | Set-Content output/live-actions.json
```

## Capture screenshots

Start screen:

```powershell
node C:/Users/cozmu/.codex/skills/develop-web-game/scripts/web_game_playwright_client.js --url http://127.0.0.1:8080 --actions-file output/start-actions.json --iterations 1 --pause-ms 120 --screenshot-dir output/web-game-start
```

Ready screen with board:

```powershell
Invoke-WebRequest -UseBasicParsing -Method Post -ContentType 'application/json' -Body '{"preset":0}' http://127.0.0.1:8080/api/start | Out-Null
node C:/Users/cozmu/.codex/skills/develop-web-game/scripts/web_game_playwright_client.js --url http://127.0.0.1:8080 --actions-file output/start-actions.json --iterations 1 --pause-ms 120 --screenshot-dir output/web-game-ready
```

Game-over or terminal state:

```powershell
node C:/Users/cozmu/.codex/skills/develop-web-game/scripts/web_game_playwright_client.js --url http://127.0.0.1:8080 --actions-file output/live-actions.json --iterations 2 --pause-ms 150 --screenshot-dir output/web-game-terminal
```

## Freeze state through the API when needed

The Go server keeps ticking in real time, so live captures can race into game-over before the screenshot lands. For overlays or stable board states, set the state first and then capture:

```powershell
Invoke-WebRequest -UseBasicParsing -Method Post -ContentType 'application/json' -Body '{"preset":0}' http://127.0.0.1:8080/api/start | Out-Null
Invoke-WebRequest -UseBasicParsing -Method Post -ContentType 'application/json' -Body '{}' http://127.0.0.1:8080/api/pause | Out-Null
node C:/Users/cozmu/.codex/skills/develop-web-game/scripts/web_game_playwright_client.js --url http://127.0.0.1:8080 --actions-file output/start-actions.json --iterations 1 --pause-ms 100 --screenshot-dir output/web-game-paused
```

## Review artifacts

- Screenshots land in the `output/web-game-*` folder passed to `--screenshot-dir`.
- State dumps land beside them as `state-*.json`.
- Console errors, if any, land as `errors-*.json`.

Open a screenshot locally with the image viewer tool using a full path, for example:

- `c:\Users\cozmu\Projects\snake\output\web-game-ready\shot-0.png`

## Known gotchas

- If the screenshot still shows the old HUD, the server was not restarted after the static-file edit.
- If `render_game_to_text` says the game is paused or over but the overlay copy looks wrong, trust the screenshot first and then inspect `state-*.json`.
- The Playwright client can click selectors like `#start-btn`, but it cannot send `P`; use the pause API when you need a deterministic paused capture.
