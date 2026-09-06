# SM Unlocker

Unlock generator for Scrap Mechanic. Reads your `unlock` file, detects the
Steam ID from the `User_<id>` folder, and rewrites the file with all 252
outfits. Creates an `unlock.bak` backup on first run.

Black-only minimal UI. Windows and Linux. Go + Fyne.

## Run

```sh
go run .
```

## Build

Linux:

```sh
go build -ldflags="-s -w" -o sm-unlocker-linux .
```

Windows (cross-compile from Linux, needs `x86_64-w64-mingw32-gcc`):

```sh
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
GOOS=windows GOARCH=amd64 \
go build -ldflags="-s -w -H=windowsgui" -o sm-unlocker-windows.exe .
```

Linux file picker is the native GTK dialog (needs `gtk3` dev packages to
build); Windows uses the native `GetOpenFileName` dialog.

## Test

```sh
go test ./... -count=1
```

## Layout

```text
main.go                 # app setup, black theme, opens UI
theme/black.go          # monochrome dark-only Fyne theme
ui/window.go            # minimal window: file row, Unlock button, status line
internal/unlock/        # unlock file codec + 252 outfit IDs
internal/steam/         # Scrap Mechanic folder lookup (Windows + Proton)
internal/service/       # inspect/unlock orchestration, backup, atomic write
```
