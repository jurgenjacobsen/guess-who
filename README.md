# Guess Who?

## Build the executable
The Go application embeds the production React build from `app/dist` into the executable.

Run this from the repo root to build the React app, regenerate the Windows icon resource, and package everything into `GuessWho-{version}.exe` (version from root `package.json`):

```powershell
.\build.cmd
```

The build pipeline itself lives in `tools/build.ps1`.
