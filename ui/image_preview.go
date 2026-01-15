package ui

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// ImagePreview displays an image with proper scaling
type ImagePreview struct {
	container    *fyne.Container
	canvasImage  *canvas.Image
	currentImage image.Image
}

// NewImagePreview creates a new image preview component
func NewImagePreview() *ImagePreview {
	p := &ImagePreview{}
	p.setup()
	return p
}

// setup initializes the preview UI
func (p *ImagePreview) setup() {
	p.canvasImage = canvas.NewImageFromImage(nil)
	p.canvasImage.FillMode = canvas.ImageFillContain
	p.canvasImage.SetMinSize(fyne.NewSize(300, 200))

	p.container = container.NewCenter(p.canvasImage)
}

// SetImage sets the image to display
func (p *ImagePreview) SetImage(img image.Image) {
	p.currentImage = img
	p.canvasImage.Image = img
	p.canvasImage.Refresh()
}

// GetImage returns the current image
func (p *ImagePreview) GetImage() image.Image {
	return p.currentImage
}

// Container returns the preview's container
func (p *ImagePreview) Container() fyne.CanvasObject {
	return p.container
}
