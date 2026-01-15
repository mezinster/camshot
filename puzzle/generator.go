package puzzle

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"
)

// Generator creates jigsaw puzzle pieces from an image
type Generator struct {
	source    image.Image
	rows      int
	columns   int
	pieceW    int
	pieceH    int
	tabSize   float64
	edgeTypes [][]EdgeSet // Edge types for each piece
	rng       *rand.Rand
}

// EdgeSet contains the edge types for all four sides of a piece
type EdgeSet struct {
	Top    EdgeType
	Right  EdgeType
	Bottom EdgeType
	Left   EdgeType
}

// NewGenerator creates a new puzzle generator
func NewGenerator(source image.Image, rows, columns int) *Generator {
	bounds := source.Bounds()

	g := &Generator{
		source:  source,
		rows:    rows,
		columns: columns,
		pieceW:  bounds.Dx() / columns,
		pieceH:  bounds.Dy() / rows,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Tab size is proportional to piece size
	g.tabSize = float64(min(g.pieceW, g.pieceH)) * 0.2

	// Generate edge types
	g.generateEdgeTypes()

	return g
}

// generateEdgeTypes determines which edges have tabs/slots
func (g *Generator) generateEdgeTypes() {
	g.edgeTypes = make([][]EdgeSet, g.rows)

	for row := 0; row < g.rows; row++ {
		g.edgeTypes[row] = make([]EdgeSet, g.columns)

		for col := 0; col < g.columns; col++ {
			edges := EdgeSet{}

			// Top edge
			if row == 0 {
				edges.Top = EdgeFlat
			} else {
				// Must match the bottom of the piece above
				edges.Top = g.edgeTypes[row-1][col].Bottom.Opposite()
			}

			// Left edge
			if col == 0 {
				edges.Left = EdgeFlat
			} else {
				// Must match the right of the piece to the left
				edges.Left = g.edgeTypes[row][col-1].Right.Opposite()
			}

			// Right edge
			if col == g.columns-1 {
				edges.Right = EdgeFlat
			} else {
				// Randomly choose tab or slot
				if g.rng.Float64() < 0.5 {
					edges.Right = EdgeTab
				} else {
					edges.Right = EdgeSlot
				}
			}

			// Bottom edge
			if row == g.rows-1 {
				edges.Bottom = EdgeFlat
			} else {
				// Randomly choose tab or slot
				if g.rng.Float64() < 0.5 {
					edges.Bottom = EdgeTab
				} else {
					edges.Bottom = EdgeSlot
				}
			}

			g.edgeTypes[row][col] = edges
		}
	}
}

// Generate creates all puzzle pieces
func (g *Generator) Generate() []image.Image {
	pieces := make([]image.Image, g.rows*g.columns)

	for row := 0; row < g.rows; row++ {
		for col := 0; col < g.columns; col++ {
			idx := row*g.columns + col
			pieces[idx] = g.generatePiece(row, col)
		}
	}

	return pieces
}

// generatePiece creates a single puzzle piece
func (g *Generator) generatePiece(row, col int) image.Image {
	edges := g.edgeTypes[row][col]

	// Calculate the extended bounds for this piece (including tabs)
	tabPx := int(g.tabSize)

	// Piece dimensions with extra space for tabs
	width := g.pieceW + tabPx*2
	height := g.pieceH + tabPx*2

	// Create the piece image with transparency
	piece := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create the mask for this piece
	mask := g.createPieceMask(edges, width, height, tabPx)

	// Source position in the original image
	srcX := col * g.pieceW
	srcY := row * g.pieceH

	// Copy pixels from source where mask allows
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Check if this pixel is inside the mask
			if mask.At(x, y).(color.RGBA).A > 0 {
				// Calculate source coordinates
				sx := srcX + x - tabPx
				sy := srcY + y - tabPx

				// Check bounds
				srcBounds := g.source.Bounds()
				if sx >= srcBounds.Min.X && sx < srcBounds.Max.X &&
					sy >= srcBounds.Min.Y && sy < srcBounds.Max.Y {
					piece.Set(x, y, g.source.At(sx, sy))
				}
			}
		}
	}

	// Add a subtle border for visibility
	g.addBorder(piece, mask)

	return piece
}

// createPieceMask creates the jigsaw-shaped mask for a piece
func (g *Generator) createPieceMask(edges EdgeSet, width, height, tabPx int) *image.RGBA {
	mask := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the base rectangle
	white := color.RGBA{255, 255, 255, 255}
	baseRect := image.Rect(tabPx, tabPx, tabPx+g.pieceW, tabPx+g.pieceH)
	draw.Draw(mask, baseRect, &image.Uniform{white}, image.Point{}, draw.Src)

	// Add/remove tabs on each edge
	tabGen := NewEdgeGenerator(g.tabSize)

	// Top edge
	if edges.Top != EdgeFlat {
		centerX := tabPx + g.pieceW/2
		tabGen.DrawTab(mask, centerX, tabPx, DirectionUp, edges.Top == EdgeTab)
	}

	// Bottom edge
	if edges.Bottom != EdgeFlat {
		centerX := tabPx + g.pieceW/2
		tabGen.DrawTab(mask, centerX, tabPx+g.pieceH, DirectionDown, edges.Bottom == EdgeTab)
	}

	// Left edge
	if edges.Left != EdgeFlat {
		centerY := tabPx + g.pieceH/2
		tabGen.DrawTab(mask, tabPx, centerY, DirectionLeft, edges.Left == EdgeTab)
	}

	// Right edge
	if edges.Right != EdgeFlat {
		centerY := tabPx + g.pieceH/2
		tabGen.DrawTab(mask, tabPx+g.pieceW, centerY, DirectionRight, edges.Right == EdgeTab)
	}

	return mask
}

// addBorder adds a subtle border around the piece edges
func (g *Generator) addBorder(piece, mask *image.RGBA) {
	borderColor := color.RGBA{80, 80, 80, 200}
	bounds := piece.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Check if this is an edge pixel
			if mask.At(x, y).(color.RGBA).A > 0 {
				// Check neighbors
				isEdge := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						nx, ny := x+dx, y+dy
						if nx < bounds.Min.X || nx >= bounds.Max.X ||
							ny < bounds.Min.Y || ny >= bounds.Max.Y {
							isEdge = true
							break
						}
						if mask.At(nx, ny).(color.RGBA).A == 0 {
							isEdge = true
							break
						}
					}
					if isEdge {
						break
					}
				}

				if isEdge {
					// Blend border color with existing color
					existing := piece.At(x, y).(color.RGBA)
					blended := blendColors(existing, borderColor, 0.5)
					piece.Set(x, y, blended)
				}
			}
		}
	}
}

// blendColors blends two colors with the given ratio
func blendColors(c1, c2 color.RGBA, ratio float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c1.R)*(1-ratio) + float64(c2.R)*ratio),
		G: uint8(float64(c1.G)*(1-ratio) + float64(c2.G)*ratio),
		B: uint8(float64(c1.B)*(1-ratio) + float64(c2.B)*ratio),
		A: c1.A,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
