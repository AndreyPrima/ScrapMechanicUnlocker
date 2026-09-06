# SM Unlocker

![Release](https://img.shields.io/github/v/release/AndreyPrima/ScrapMechanicUnlocker?style=flat-square) ![Downloads](https://img.shields.io/github/downloads/AndreyPrima/ScrapMechanicUnlocker/total?style=flat-square) ![Go](https://img.shields.io/github/go-mod/go-version/AndreyPrima/ScrapMechanicUnlocker?style=flat-square) ![Platform](https://img.shields.io/badge/Windows_%7C_Linux-black?style=flat-square)

Point SM Unlocker at your Scrap Mechanic `unlock` file and unlock all 252 outfits in one click. The app takes your Steam ID from the `User_<id>` folder and stores your original file as `unlock.bak` on first run.

You get one black window on Windows and Linux. The code is Go + Fyne.

## Origin

This app shipped as a closed-source Windows `.exe`. I decompiled it, rebuilt it in Go + Fyne through vibecode, and put the source here. You get the same unlock format, byte for byte.

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

## Bugs

You may run into bugs. Open an issue and describe what broke, with your OS and the status line text.
