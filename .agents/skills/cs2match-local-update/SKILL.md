---
name: cs2match-local-update
description: Rebuild and reload this project's local Docker frontend, Nakama Go backend, or both. Use for natural-language requests such as “帮我更新前端”, “帮我更新后端”, or “帮我更新前后端”.
---

# Update CS2Match locally

Run all commands from the repository root. This workflow updates local build artifacts and Docker services; it does not mean editing application source code.

## Select the target

- “更新前端” means frontend only.
- “更新后端” means backend only.
- “更新前后端”, “更新前端和后端”, or “更新全部” means backend first, then frontend.

Do not rebuild the unrequested side. Do not run Luban generation merely because generated config files exist. If the user explicitly asks to publish changed Excel/Luban configuration, use `scripts/update-local-config.ps1`, which performs generation plus both updates.

## Update the backend

1. On Windows PowerShell or CMD, run `cmd.exe /d /c server\build.bat`. On Linux, macOS, Git Bash, or WSL2, run `bash server/build.sh`.
2. Stop immediately if compilation fails. Do not restart Nakama with an old artifact and do not claim success.
3. Run `docker compose up -d db nakama`, then `docker compose restart nakama` so the running process loads the new `server/build/backend.so`.
4. Check `docker compose ps nakama` and `docker compose logs --tail 100 nakama`. Treat plugin-loading errors or an unhealthy container as failure and report the relevant log lines.

Use the repository build scripts rather than reconstructing the `docker run` command. They pin `heroiclabs/nakama-pluginbuilder:3.30.0` and the required Go build flags.

## Update the frontend

1. Run `docker compose build --no-cache frontend`.
2. Stop immediately if the image build fails.
3. Run `docker compose up -d --no-deps --force-recreate frontend`.
4. Check `docker compose ps frontend`. When an HTTP check is available, also verify that `http://localhost:3000` responds.

The `--no-cache` build is intentional for an explicit update request: it avoids serving an image made from a stale cached source layer.

## Update both

Complete and verify the backend workflow first, then complete and verify the frontend workflow. Stop at the first failed step and state which side was not updated.

Finish with a concise report naming the updated targets, whether each verification passed, and any command the user must retry. Never report success based only on a zero-exit build if the corresponding service failed to start or reload.
