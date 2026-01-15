package print

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/jung-kurt/gofpdf"
)

// A4 dimensions in millimeters
const (
	A4Width  = 210.0
	A4Height = 297.0
	Margin   = 10.0 // 10mm margins
)

// A4Layout handles PDF generation for A4 paper
type A4Layout struct {
	pieces  []image.Image
	rows    int
	columns int
}

// NewA4Layout creates a new A4 layout generator
func NewA4Layout(pieces []image.Image, rows, columns int) *A4Layout {
	return &A4Layout{
		pieces:  pieces,
		rows:    rows,
		columns: columns,
	}
}

// GeneratePDF creates a printable PDF file
func (l *A4Layout) GeneratePDF(outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("CamShot Puzzle", false)
	pdf.SetAuthor("CamShot Puzzle Maker", false)

	// Calculate available space
	availableW := A4Width - 2*Margin
	availableH := A4Height - 2*Margin - 20 // 20mm for header

	// Calculate piece size in mm
	// Try to fit all pieces on one page if possible
	totalPieces := len(l.pieces)
	piecesPerPage := l.calculatePiecesPerPage(availableW, availableH)

	// Calculate number of pages needed
	numPages := (totalPieces + piecesPerPage - 1) / piecesPerPage

	pieceIdx := 0
	for page := 0; page < numPages; page++ {
		pdf.AddPage()

		// Add header
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, fmt.Sprintf("Puzzle Pieces - Page %d of %d", page+1, numPages))
		pdf.Ln(15)

		// Add instructions on first page
		if page == 0 {
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(0, 5, fmt.Sprintf("Total pieces: %d (%d rows x %d columns)", totalPieces, l.rows, l.columns))
			pdf.Ln(8)
		}

		// Calculate grid layout for this page
		gridCols, gridRows, pieceW, pieceH := l.calculatePageLayout(availableW, availableH-15, piecesPerPage)

		// Add pieces to this page
		startY := Margin + 25 // After header
		startX := Margin

		for r := 0; r < gridRows && pieceIdx < totalPieces; r++ {
			for c := 0; c < gridCols && pieceIdx < totalPieces; c++ {
				x := startX + float64(c)*(pieceW+5) // 5mm gap between pieces
				y := startY + float64(r)*(pieceH+15) // 15mm gap for labels

				// Add the piece image
				err := l.addImageToPDF(pdf, l.pieces[pieceIdx], x, y, pieceW, pieceH)
				if err != nil {
					// Skip this piece but continue
					pdf.SetFont("Arial", "", 8)
					pdf.Text(x, y+pieceH/2, fmt.Sprintf("Piece %d", pieceIdx+1))
				}

				// Add piece number
				pdf.SetFont("Arial", "", 9)
				pdf.Text(x+pieceW/2-3, y+pieceH+5, fmt.Sprintf("%d", pieceIdx+1))

				pieceIdx++
			}
		}

		// Add cutting guide note
		pdf.SetFont("Arial", "I", 8)
		pdf.SetY(A4Height - 15)
		pdf.Cell(0, 5, "Cut along the piece edges. Numbers help with assembly.")
	}

	// Add assembly guide page
	l.addAssemblyGuidePage(pdf)

	return pdf.OutputFileAndClose(outputPath)
}

// calculatePiecesPerPage determines how many pieces fit on one page
func (l *A4Layout) calculatePiecesPerPage(availW, availH float64) int {
	// Aim for reasonable piece sizes (at least 30mm)
	minPieceSize := 30.0
	maxCols := int(availW / (minPieceSize + 5))
	maxRows := int(availH / (minPieceSize + 15))

	if maxCols < 1 {
		maxCols = 1
	}
	if maxRows < 1 {
		maxRows = 1
	}

	return maxCols * maxRows
}

// calculatePageLayout calculates the optimal grid layout for a page
func (l *A4Layout) calculatePageLayout(availW, availH float64, maxPieces int) (cols, rows int, pieceW, pieceH float64) {
	totalPieces := len(l.pieces)
	if totalPieces > maxPieces {
		totalPieces = maxPieces
	}

	// Try different configurations to find the best fit
	bestCols := 1
	bestRows := totalPieces
	bestSize := 0.0

	for c := 1; c <= totalPieces; c++ {
		r := (totalPieces + c - 1) / c

		w := (availW - float64(c-1)*5) / float64(c)
		h := (availH - float64(r-1)*15) / float64(r)

		// Keep aspect ratio square-ish
		size := min(w, h)

		if size > bestSize {
			bestSize = size
			bestCols = c
			bestRows = r
		}
	}

	cols = bestCols
	rows = bestRows
	pieceW = bestSize
	pieceH = bestSize

	return
}

// addImageToPDF adds an image to the PDF at the specified position
func (l *A4Layout) addImageToPDF(pdf *gofpdf.Fpdf, img image.Image, x, y, w, h float64) error {
	if img == nil {
		return fmt.Errorf("nil image")
	}

	// Convert image to PNG bytes
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return err
	}

	// Register the image
	imgName := fmt.Sprintf("piece_%d_%d", int(x), int(y))
	pdf.RegisterImageOptionsReader(imgName, gofpdf.ImageOptions{ImageType: "PNG"}, &buf)

	// Add to PDF
	pdf.ImageOptions(imgName, x, y, w, h, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	return nil
}

// addAssemblyGuidePage adds a page showing the puzzle layout
func (l *A4Layout) addAssemblyGuidePage(pdf *gofpdf.Fpdf) {
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Assembly Guide")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, "Arrange the pieces according to this grid:")
	pdf.Ln(10)

	// Draw the grid layout
	availW := A4Width - 2*Margin
	cellW := availW / float64(l.columns)
	cellH := cellW // Square cells

	if cellH*float64(l.rows) > 150 {
		cellH = 150 / float64(l.rows)
		cellW = cellH
	}

	startX := Margin + (availW-cellW*float64(l.columns))/2
	startY := 50.0

	pdf.SetDrawColor(100, 100, 100)
	pdf.SetFont("Arial", "", 8)

	for r := 0; r < l.rows; r++ {
		for c := 0; c < l.columns; c++ {
			x := startX + float64(c)*cellW
			y := startY + float64(r)*cellH

			// Draw cell border
			pdf.Rect(x, y, cellW, cellH, "D")

			// Add piece number
			pieceNum := r*l.columns + c + 1
			pdf.Text(x+cellW/2-3, y+cellH/2+2, fmt.Sprintf("%d", pieceNum))
		}
	}

	// Add tips
	pdf.SetY(startY + float64(l.rows)*cellH + 20)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Assembly Tips:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	tips := []string{
		"1. Start with corner pieces (they have two flat edges)",
		"2. Build the border next (pieces with one flat edge)",
		"3. Group pieces by color or pattern similarity",
		"4. Work on small sections at a time",
		"5. The tabs (bumps) fit into slots (holes) of adjacent pieces",
	}

	for _, tip := range tips {
		pdf.Cell(0, 6, tip)
		pdf.Ln(7)
	}
}

// SaveIndividualPieces saves each piece as a separate image file
func (l *A4Layout) SaveIndividualPieces(outputDir string) error {
	for i, piece := range l.pieces {
		if piece == nil {
			continue
		}

		filename := fmt.Sprintf("%s/piece_%02d.png", outputDir, i+1)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}

		err = png.Encode(file, piece)
		file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
