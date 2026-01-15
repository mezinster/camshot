package ui

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// UploadView handles photo upload functionality
type UploadView struct {
	app           *App
	container     *fyne.Container
	previewImage  *canvas.Image
	statusLabel   *widget.Label
	loadedImage   image.Image
}

// NewUploadView creates a new upload view
func NewUploadView(app *App) *UploadView {
	v := &UploadView{
		app: app,
	}
	v.setup()
	return v
}

// setup initializes the upload view UI
func (v *UploadView) setup() {
	// Create placeholder image
	v.previewImage = canvas.NewImageFromImage(nil)
	v.previewImage.FillMode = canvas.ImageFillContain
	v.previewImage.SetMinSize(fyne.NewSize(400, 300))

	// Status label
	v.statusLabel = widget.NewLabel("No image loaded")
	v.statusLabel.Alignment = fyne.TextAlignCenter

	// Upload button
	uploadBtn := widget.NewButton("Select Image File", func() {
		v.openFileDialog()
	})
	uploadBtn.Importance = widget.HighImportance

	// Use image button
	useBtn := widget.NewButton("Use This Image", func() {
		v.useImage()
	})

	// Instructions
	instructions := widget.NewLabel(
		"Supported formats: JPEG, PNG\n" +
			"Select an image file to create a puzzle from it.",
	)
	instructions.Wrapping = fyne.TextWrapWord
	instructions.Alignment = fyne.TextAlignCenter

	// Preview container with border
	previewContainer := container.NewBorder(
		widget.NewLabel("Image Preview"),
		v.statusLabel,
		nil, nil,
		container.NewCenter(v.previewImage),
	)

	// Buttons row
	buttonsRow := container.NewHBox(
		uploadBtn,
		useBtn,
	)

	// Main layout
	v.container = container.NewBorder(
		instructions,
		container.NewCenter(buttonsRow),
		nil, nil,
		previewContainer,
	)
}

// openFileDialog opens a file selection dialog
func (v *UploadView) openFileDialog() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			showError(v.app.GetWindow(), "Error opening file: "+err.Error())
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer reader.Close()

		v.loadImage(reader.URI().Path())
	}, v.app.GetWindow())

	// Set file filter for images
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG"}))
	fd.Show()
}

// loadImage loads an image from the given path
func (v *UploadView) loadImage(path string) {
	file, err := os.Open(path)
	if err != nil {
		showError(v.app.GetWindow(), "Error opening file: "+err.Error())
		return
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		showError(v.app.GetWindow(), "Error decoding image: "+err.Error())
		return
	}

	v.loadedImage = img
	v.previewImage.Image = img
	v.previewImage.Refresh()

	bounds := img.Bounds()
	v.statusLabel.SetText("Loaded: " + format + " (" +
		intToStr(bounds.Dx()) + "x" + intToStr(bounds.Dy()) + ")")
}

// useImage sets the loaded image as the current working image
func (v *UploadView) useImage() {
	if v.loadedImage == nil {
		showError(v.app.GetWindow(), "No image loaded. Please select an image first.")
		return
	}

	v.app.SetImage(v.loadedImage)
	showInfo(v.app.GetWindow(), "Image loaded! Go to 'Puzzle Settings' tab to configure and generate your puzzle.")
}

// Container returns the view's container
func (v *UploadView) Container() fyne.CanvasObject {
	return v.container
}
