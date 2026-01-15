package ui

import (
	"camshot/print"
	"camshot/puzzle"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// PuzzleView displays generated puzzle pieces
type PuzzleView struct {
	container    *fyne.Container
	gridContainer *fyne.Container
	pieces       []image.Image
	rows         int
	columns      int
}

// NewPuzzleView creates a new puzzle view
func NewPuzzleView() *PuzzleView {
	v := &PuzzleView{}
	v.setup()
	return v
}

// setup initializes the puzzle view UI
func (v *PuzzleView) setup() {
	placeholder := widget.NewLabel("Generate a puzzle to see preview")
	placeholder.Alignment = fyne.TextAlignCenter

	v.gridContainer = container.NewCenter(placeholder)

	v.container = container.NewBorder(
		nil, nil, nil, nil,
		v.gridContainer,
	)
}

// GeneratePuzzle creates puzzle pieces from the source image
func (v *PuzzleView) GeneratePuzzle(source image.Image, rows, columns int) {
	v.rows = rows
	v.columns = columns

	// Generate the puzzle pieces
	generator := puzzle.NewGenerator(source, rows, columns)
	v.pieces = generator.Generate()

	// Update the preview
	v.updatePreview()
}

// updatePreview updates the puzzle preview grid
func (v *PuzzleView) updatePreview() {
	if len(v.pieces) == 0 {
		return
	}

	// Create a grid of puzzle piece images
	grid := container.NewGridWithColumns(v.columns)

	for i, piece := range v.pieces {
		if piece == nil {
			continue
		}

		img := canvas.NewImageFromImage(piece)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(80, 80))

		// Add piece number label
		pieceNum := widget.NewLabel(fmt.Sprintf("%d", i+1))
		pieceNum.Alignment = fyne.TextAlignCenter

		pieceContainer := container.NewBorder(
			nil,
			pieceNum,
			nil, nil,
			img,
		)

		grid.Add(pieceContainer)
	}

	// Wrap in scroll container for many pieces
	scroll := container.NewScroll(grid)

	v.gridContainer.Objects = []fyne.CanvasObject{scroll}
	v.gridContainer.Refresh()
}

// GetPieces returns the generated puzzle pieces
func (v *PuzzleView) GetPieces() []image.Image {
	return v.pieces
}

// ExportToPDF exports the puzzle to a PDF file
func (v *PuzzleView) ExportToPDF(parent fyne.Window) {
	if len(v.pieces) == 0 {
		showError(parent, "No puzzle generated")
		return
	}

	// Ask user for save location
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			showError(parent, "Error: "+err.Error())
			return
		}
		if writer == nil {
			return // User cancelled
		}
		defer writer.Close()

		// Generate PDF
		pdfGen := print.NewA4Layout(v.pieces, v.rows, v.columns)
		err = pdfGen.GeneratePDF(writer.URI().Path())
		if err != nil {
			showError(parent, "Error generating PDF: "+err.Error())
			return
		}

		showInfo(parent, "PDF exported successfully!")
	}, parent)
}

// ExportAsImages exports puzzle pieces as individual image files
func (v *PuzzleView) ExportAsImages(parent fyne.Window) {
	if len(v.pieces) == 0 {
		showError(parent, "No puzzle generated")
		return
	}

	// Ask user for folder selection
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			showError(parent, "Error: "+err.Error())
			return
		}
		if uri == nil {
			return // User cancelled
		}

		folderPath := uri.Path()

		// Save each piece
		for i, piece := range v.pieces {
			if piece == nil {
				continue
			}

			filename := filepath.Join(folderPath, fmt.Sprintf("piece_%02d.png", i+1))
			file, err := os.Create(filename)
			if err != nil {
				showError(parent, "Error creating file: "+err.Error())
				return
			}

			err = png.Encode(file, piece)
			file.Close()
			if err != nil {
				showError(parent, "Error saving image: "+err.Error())
				return
			}
		}

		showInfo(parent, fmt.Sprintf("Exported %d puzzle pieces to folder!", len(v.pieces)))
	}, parent)
}

// Container returns the view's container
func (v *PuzzleView) Container() fyne.CanvasObject {
	return v.container
}
