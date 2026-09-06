# SM Unlocker

Point SM Unlocker at your Scrap Mechanic `unlock` file and unlock all 252 outfits in one click. The app takes your Steam ID from the `User_<id>` folder and stores your original file as `unlock.bak` on first run.

You get one black window on Windows and Linux. The code is Go + Fyne.

## Run

```sh
go run .
```

## Build

Build for Linux:

```sh
go build -ldflags="-s -w" -o sm-unlocker-linux .
```

Cross-compile for Windows from Linux with `x86_64-w64-mingw32-gcc` installed:

```sh
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
GOOS=windows GOARCH=amd64 \
go build -ldflags="-s -w -H=windowsgui" -o sm-unlocker-windows.exe .
```

You pick files through the native GTK dialog on Linux (install the `gtk3` dev packages to build it) and through `GetOpenFileName` on Windows.

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
