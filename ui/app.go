package ui

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// App represents the main application
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window

	// Current image being worked on
	currentImage image.Image

	// Puzzle configuration
	rows    int
	columns int

	// UI components
	imagePreview *ImagePreview
	puzzleView   *PuzzleView
}

// NewApp creates a new CamShot application
func NewApp() *App {
	a := &App{
		fyneApp: app.New(),
		rows:    3,
		columns: 3,
	}

	a.mainWindow = a.fyneApp.NewWindow("CamShot Puzzle Maker")
	a.mainWindow.Resize(fyne.NewSize(900, 700))

	return a
}

// Run starts the application
func (a *App) Run() {
	a.setupUI()
	a.mainWindow.ShowAndRun()
}

// setupUI creates the main user interface
func (a *App) setupUI() {
	// Create the image preview component
	a.imagePreview = NewImagePreview()

	// Create the puzzle view component
	a.puzzleView = NewPuzzleView()

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Upload Photo", a.createUploadTab()),
		container.NewTabItem("Camera Capture", a.createCameraTab()),
		container.NewTabItem("Puzzle Settings", a.createPuzzleTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	a.mainWindow.SetContent(tabs)
}

// createUploadTab creates the photo upload interface
func (a *App) createUploadTab() fyne.CanvasObject {
	uploadView := NewUploadView(a)
	return uploadView.Container()
}

// createCameraTab creates the camera capture interface
func (a *App) createCameraTab() fyne.CanvasObject {
	cameraView := NewCameraView(a)
	return cameraView.Container()
}

// createPuzzleTab creates the puzzle configuration interface
func (a *App) createPuzzleTab() fyne.CanvasObject {
	// Rows slider
	rowsLabel := widget.NewLabel("Rows: 3")
	rowsSlider := widget.NewSlider(2, 10)
	rowsSlider.Value = 3
	rowsSlider.Step = 1
	rowsSlider.OnChanged = func(v float64) {
		a.rows = int(v)
		rowsLabel.SetText("Rows: " + intToStr(a.rows))
	}

	// Columns slider
	colsLabel := widget.NewLabel("Columns: 3")
	colsSlider := widget.NewSlider(2, 10)
	colsSlider.Value = 3
	colsSlider.Step = 1
	colsSlider.OnChanged = func(v float64) {
		a.columns = int(v)
		colsLabel.SetText("Columns: " + intToStr(a.columns))
	}

	// Generate puzzle button
	generateBtn := widget.NewButton("Generate Puzzle", func() {
		a.generatePuzzle()
	})
	generateBtn.Importance = widget.HighImportance

	// Export buttons
	exportPDFBtn := widget.NewButton("Export to PDF (A4)", func() {
		a.exportToPDF()
	})

	exportImagesBtn := widget.NewButton("Export as Images", func() {
		a.exportAsImages()
	})

	// Status label
	statusLabel := widget.NewLabel("Upload or capture a photo to begin")
	statusLabel.Wrapping = fyne.TextWrapWord

	// Layout
	configForm := container.NewVBox(
		widget.NewLabel("Puzzle Configuration"),
		widget.NewSeparator(),
		container.NewGridWithColumns(2, rowsLabel, rowsSlider),
		container.NewGridWithColumns(2, colsLabel, colsSlider),
		widget.NewSeparator(),
		generateBtn,
		widget.NewSeparator(),
		widget.NewLabel("Export Options"),
		container.NewGridWithColumns(2, exportPDFBtn, exportImagesBtn),
		widget.NewSeparator(),
		statusLabel,
	)

	// Puzzle preview on the right
	previewContainer := container.NewBorder(
		widget.NewLabel("Puzzle Preview"),
		nil, nil, nil,
		a.puzzleView.Container(),
	)

	return container.NewHSplit(configForm, previewContainer)
}

// SetImage sets the current working image
func (a *App) SetImage(img image.Image) {
	a.currentImage = img
	a.imagePreview.SetImage(img)
}

// GetImage returns the current working image
func (a *App) GetImage() image.Image {
	return a.currentImage
}

// GetRows returns the current row count
func (a *App) GetRows() int {
	return a.rows
}

// GetColumns returns the current column count
func (a *App) GetColumns() int {
	return a.columns
}

// GetWindow returns the main window
func (a *App) GetWindow() fyne.Window {
	return a.mainWindow
}

// generatePuzzle generates the puzzle from the current image
func (a *App) generatePuzzle() {
	if a.currentImage == nil {
		showError(a.mainWindow, "No image loaded. Please upload or capture a photo first.")
		return
	}

	a.puzzleView.GeneratePuzzle(a.currentImage, a.rows, a.columns)
	showInfo(a.mainWindow, "Puzzle generated successfully!")
}

// exportToPDF exports the puzzle to a PDF file
func (a *App) exportToPDF() {
	if a.puzzleView.GetPieces() == nil {
		showError(a.mainWindow, "Please generate a puzzle first.")
		return
	}

	a.puzzleView.ExportToPDF(a.mainWindow)
}

// exportAsImages exports puzzle pieces as individual images
func (a *App) exportAsImages() {
	if a.puzzleView.GetPieces() == nil {
		showError(a.mainWindow, "Please generate a puzzle first.")
		return
	}

	a.puzzleView.ExportAsImages(a.mainWindow)
}

// Helper function to convert int to string
func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}
