# CamShot Python MVP

A Python/Tkinter rewrite of CamShot - converts photos into printable jigsaw puzzles.

## Setup

### 1. Install System Dependencies

**Ubuntu/Debian:**
```bash
sudo apt-get install python3-tk
```

**Fedora:**
```bash
sudo dnf install python3-tkinter
```

**macOS** (with Homebrew):
```bash
brew install python-tk
```

**Windows:**
Tkinter is included with the standard Python installer.

### 2. Install Python Dependencies

```bash
cd python_version
pip install -r requirements.txt
```

Or install individually:
```bash
pip install Pillow opencv-python reportlab numpy
```

### 3. Run the Application

```bash
python main.py
```

## Features

- **Image Loading**: Load images from files (PNG, JPG, etc.) or capture from webcam
- **Puzzle Generation**: Create jigsaw pieces with interlocking tabs and slots
- **Grid Configuration**: Choose puzzle size (2-10 columns/rows)
- **Export Options**:
  - Individual PNG files (with transparency)
  - Multi-page A4 PDF for printing

## Project Structure

```
python_version/
├── main.py              # Entry point
├── requirements.txt     # Python dependencies
├── ui/
│   ├── app.py           # Main application window
│   ├── image_tab.py     # Image loading/camera capture
│   ├── puzzle_tab.py    # Grid config and generation
│   └── export_tab.py    # PNG/PDF export
├── puzzle/
│   └── generator.py     # Piece generation with interlocking edges
├── camera/
│   └── capture.py       # OpenCV webcam capture
└── export/
    └── pdf.py           # A4 PDF generation
```

## Comparison with Go Version

| Feature | Go/Fyne | Python/Tkinter |
|---------|---------|----------------|
| Camera Support | Windows-only (VfW) | Cross-platform (OpenCV) |
| Distribution | Single binary | Requires Python + deps |
| UI Look | Modern (Fyne) | Classic (ttk themed) |
| Performance | Faster | Adequate for this use |
