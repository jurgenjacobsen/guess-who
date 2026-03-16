# Guess Who?
Guess Who game with a React frontend and Go backend, featuring player and moderator modes with editable people cards and a single-command Windows build.

## Build the executable
The Go application embeds the production React build from `app/dist` into the executable.

Run this from the repo root to build the React app, regenerate the Windows icon resource, and package everything into `GuessWho-{version}.exe` (version from root `package.json`):

```powershell
.\build.cmd
```

The build pipeline itself lives in `tools/build.ps1`.
