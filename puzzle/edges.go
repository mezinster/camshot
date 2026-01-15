package puzzle

import (
	"image"
	"image/color"
	"math"
)

// EdgeType represents the type of edge on a puzzle piece
type EdgeType int

const (
	EdgeFlat EdgeType = iota // Flat edge (border pieces)
	EdgeTab                  // Outward protrusion
	EdgeSlot                 // Inward indentation
)

// Opposite returns the complementary edge type
func (e EdgeType) Opposite() EdgeType {
	switch e {
	case EdgeTab:
		return EdgeSlot
	case EdgeSlot:
		return EdgeTab
	default:
		return EdgeFlat
	}
}

// Direction represents the direction of an edge
type Direction int

const (
	DirectionUp Direction = iota
	DirectionRight
	DirectionDown
	DirectionLeft
)

// EdgeGenerator creates jigsaw puzzle edges using Bezier curves
type EdgeGenerator struct {
	tabSize float64
}

// NewEdgeGenerator creates a new edge generator
func NewEdgeGenerator(tabSize float64) *EdgeGenerator {
	return &EdgeGenerator{tabSize: tabSize}
}

// DrawTab draws a tab or slot on the mask at the specified position
func (e *EdgeGenerator) DrawTab(mask *image.RGBA, cx, cy int, dir Direction, isTab bool) {
	white := color.RGBA{255, 255, 255, 255}
	transparent := color.RGBA{0, 0, 0, 0}

	// Generate the tab shape points
	points := e.generateTabShape(float64(cx), float64(cy), dir, isTab)

	// Fill the shape
	if isTab {
		// Add pixels (draw white)
		e.fillShape(mask, points, white)
	} else {
		// Remove pixels (draw transparent)
		e.fillShape(mask, points, transparent)
	}
}

// Point represents a 2D point
type Point struct {
	X, Y float64
}

// generateTabShape generates the points defining the tab/slot shape
func (e *EdgeGenerator) generateTabShape(cx, cy float64, dir Direction, isTab bool) []Point {
	// Tab dimensions
	tabWidth := e.tabSize * 1.2
	tabHeight := e.tabSize
	neckWidth := e.tabSize * 0.6

	// Generate the shape based on direction
	// For tabs: shape protrudes outward from the piece
	// For slots: shape cuts inward into the piece (opposite direction)
	var points []Point

	switch dir {
	case DirectionUp:
		if isTab {
			// Tab going up (outward from piece)
			points = e.generateVerticalTab(cx, cy, tabWidth, tabHeight, neckWidth, true)
		} else {
			// Slot: cut downward into piece so neighbor's tab fits
			points = e.generateVerticalTab(cx, cy, tabWidth, tabHeight, neckWidth, false)
		}
	case DirectionDown:
		if isTab {
			// Tab going down (outward from piece)
			points = e.generateVerticalTab(cx, cy, tabWidth, tabHeight, neckWidth, false)
		} else {
			// Slot: cut upward into piece so neighbor's tab fits
			points = e.generateVerticalTab(cx, cy, tabWidth, tabHeight, neckWidth, true)
		}
	case DirectionLeft:
		if isTab {
			// Tab going left (outward from piece)
			points = e.generateHorizontalTab(cx, cy, tabWidth, tabHeight, neckWidth, true)
		} else {
			// Slot: cut rightward into piece so neighbor's tab fits
			points = e.generateHorizontalTab(cx, cy, tabWidth, tabHeight, neckWidth, false)
		}
	case DirectionRight:
		if isTab {
			// Tab going right (outward from piece)
			points = e.generateHorizontalTab(cx, cy, tabWidth, tabHeight, neckWidth, false)
		} else {
			// Slot: cut leftward into piece so neighbor's tab fits
			points = e.generateHorizontalTab(cx, cy, tabWidth, tabHeight, neckWidth, true)
		}
	}

	return points
}

// generateVerticalTab generates a tab shape going up or down with a rounded circular head
func (e *EdgeGenerator) generateVerticalTab(cx, cy, tabW, tabH, neckW float64, up bool) []Point {
	var points []Point

	// Direction multiplier
	d := 1.0
	if up {
		d = -1.0
	}

	steps := 40

	// The head will be a semicircle with radius based on tab dimensions
	headRadius := tabW * 0.45
	neckLength := tabH - headRadius // Height of the neck portion

	// Left side of neck at base
	points = append(points, Point{cx - neckW/2, cy})

	// Left S-curve: neck widens smoothly to head
	// Uses cubic Bezier approximation for smooth transition
	for i := 0; i <= steps/2; i++ {
		t := float64(i) / float64(steps/2)
		// Smooth S-curve using cubic Bezier
		x := cubicBezier(cx-neckW/2, cx-neckW/2, cx-headRadius, cx-headRadius, t)
		y := cubicBezier(cy, cy+d*neckLength*0.4, cy+d*neckLength*0.6, cy+d*neckLength, t)
		points = append(points, Point{x, y})
	}

	// Semicircular head - draw arc from left to right
	headCenterY := cy + d*neckLength
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		// Angle goes from 180° to 0° (or 0° to 180° depending on direction)
		var angle float64
		if up {
			angle = math.Pi - t*math.Pi // 180° to 0° (top half of circle)
		} else {
			angle = t * math.Pi // 0° to 180° (bottom half of circle)
		}
		x := cx + headRadius*math.Cos(angle)
		y := headCenterY + d*headRadius*math.Sin(angle)
		points = append(points, Point{x, y})
	}

	// Right S-curve: head narrows back to neck
	for i := 0; i <= steps/2; i++ {
		t := float64(i) / float64(steps/2)
		x := cubicBezier(cx+headRadius, cx+headRadius, cx+neckW/2, cx+neckW/2, t)
		y := cubicBezier(cy+d*neckLength, cy+d*neckLength*0.6, cy+d*neckLength*0.4, cy, t)
		points = append(points, Point{x, y})
	}

	// Right side of neck back to base
	points = append(points, Point{cx + neckW/2, cy})

	return points
}

// generateHorizontalTab generates a tab shape going left or right with a rounded circular head
func (e *EdgeGenerator) generateHorizontalTab(cx, cy, tabW, tabH, neckW float64, left bool) []Point {
	var points []Point

	// Direction multiplier
	d := 1.0
	if left {
		d = -1.0
	}

	steps := 40

	// The head will be a semicircle with radius based on tab dimensions
	headRadius := tabW * 0.45
	neckLength := tabH - headRadius // Width of the neck portion

	// Top of neck at base
	points = append(points, Point{cx, cy - neckW/2})

	// Top S-curve: neck widens smoothly to head
	for i := 0; i <= steps/2; i++ {
		t := float64(i) / float64(steps/2)
		x := cubicBezier(cx, cx+d*neckLength*0.4, cx+d*neckLength*0.6, cx+d*neckLength, t)
		y := cubicBezier(cy-neckW/2, cy-neckW/2, cy-headRadius, cy-headRadius, t)
		points = append(points, Point{x, y})
	}

	// Semicircular head - draw arc from top to bottom
	headCenterX := cx + d*neckLength
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		// Angle goes around the semicircle
		var angle float64
		if left {
			angle = -math.Pi/2 - t*math.Pi // -90° to -270° (left half of circle)
		} else {
			angle = -math.Pi/2 + t*math.Pi // -90° to 90° (right half of circle)
		}
		x := headCenterX + d*headRadius*math.Cos(angle)
		y := cy + headRadius*math.Sin(angle)
		points = append(points, Point{x, y})
	}

	// Bottom S-curve: head narrows back to neck
	for i := 0; i <= steps/2; i++ {
		t := float64(i) / float64(steps/2)
		x := cubicBezier(cx+d*neckLength, cx+d*neckLength*0.6, cx+d*neckLength*0.4, cx, t)
		y := cubicBezier(cy+headRadius, cy+headRadius, cy+neckW/2, cy+neckW/2, t)
		points = append(points, Point{x, y})
	}

	// Bottom of neck back to base
	points = append(points, Point{cx, cy + neckW/2})

	return points
}

// bezierPoint calculates a point on a quadratic Bezier curve
func bezierPoint(p0, p1, p2, t float64) float64 {
	return (1-t)*(1-t)*p0 + 2*(1-t)*t*p1 + t*t*p2
}

// cubicBezier calculates a point on a cubic Bezier curve for smoother transitions
func cubicBezier(p0, p1, p2, p3, t float64) float64 {
	mt := 1 - t
	return mt*mt*mt*p0 + 3*mt*mt*t*p1 + 3*mt*t*t*p2 + t*t*t*p3
}

// fillShape fills a polygon defined by points with the given color
func (e *EdgeGenerator) fillShape(mask *image.RGBA, points []Point, c color.RGBA) {
	if len(points) < 3 {
		return
	}

	// Find bounding box
	minX, minY := points[0].X, points[0].Y
	maxX, maxY := points[0].X, points[0].Y

	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	// Scanline fill algorithm
	for y := int(minY); y <= int(maxY); y++ {
		// Find intersections with polygon edges
		var intersections []float64

		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			p1, p2 := points[i], points[j]

			if (p1.Y <= float64(y) && p2.Y > float64(y)) ||
				(p2.Y <= float64(y) && p1.Y > float64(y)) {
				// Calculate x intersection
				x := p1.X + (float64(y)-p1.Y)/(p2.Y-p1.Y)*(p2.X-p1.X)
				intersections = append(intersections, x)
			}
		}

		// Sort intersections
		for i := 0; i < len(intersections)-1; i++ {
			for j := i + 1; j < len(intersections); j++ {
				if intersections[j] < intersections[i] {
					intersections[i], intersections[j] = intersections[j], intersections[i]
				}
			}
		}

		// Fill between pairs of intersections
		for i := 0; i+1 < len(intersections); i += 2 {
			for x := int(intersections[i]); x <= int(intersections[i+1]); x++ {
				mask.Set(x, y, c)
			}
		}
	}
}
