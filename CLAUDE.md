# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Native build (Linux/macOS)
go build -o camshot

# Windows cross-compilation from Linux (requires mingw-w64)
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o camshot.exe

# Run the application
./camshot
```

Cross-compilation requires `gcc-mingw-w64-x86-64` due to Fyne's OpenGL dependency (go-gl/gl requires CGO).

## Architecture

CamShot is a Fyne-based desktop application that converts photos into printable jigsaw puzzles.

### Package Structure

- **main.go** - Entry point, creates and runs the App
- **ui/** - Fyne GUI components (App, tabs, views)
- **puzzle/** - Jigsaw piece generation with interlocking edges
- **camera/** - Platform-specific webcam capture (Windows-only via VfW API)
- **print/** - PDF generation for A4 output using gofpdf

### Key Flows

**Image → Puzzle → Export:**
1. User loads image via upload or camera capture → stored in `App.currentImage`
2. `puzzle.Generator` creates pieces with interlocking tabs/slots using Bezier curves
3. `PuzzleView` displays the grid preview
4. Export options: PDF (via `print.A4Layout`) or individual PNGs

**Edge Generation System:**
- `EdgeType` (Flat/Tab/Slot) determines piece connectivity
- Generator pre-computes edge types ensuring adjacent pieces interlock (tab matches slot)
- `EdgeGenerator.DrawTab()` renders curved edges using scanline polygon fill

### Platform Considerations

Camera capture uses build tags:
- `capture_windows.go` - Windows VfW API via avicap32.dll
- `capture_other.go` - Stub returning "not available" for non-Windows

The camera implementation is a placeholder - it creates test patterns rather than actual captures. Production use would require gocv or escapi.
