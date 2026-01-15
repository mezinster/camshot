"""
Puzzle piece generator - cuts image into interlocking pieces.
"""

from dataclasses import dataclass
from enum import Enum
from PIL import Image, ImageDraw
import random


class EdgeType(Enum):
    """Type of edge for a puzzle piece side."""
    FLAT = 0   # Border edge - straight
    TAB = 1    # Protrusion (convex)
    SLOT = 2   # Indentation (concave)


@dataclass
class PuzzlePiece:
    """A single puzzle piece with its image and metadata."""
    image: Image.Image
    row: int
    col: int
    edges: tuple[EdgeType, EdgeType, EdgeType, EdgeType]  # top, right, bottom, left


class PuzzleGenerator:
    """Generates jigsaw puzzle pieces from an image."""

    # Tab size as fraction of piece dimension
    TAB_SIZE_RATIO = 0.20

    def __init__(self, image: Image.Image, cols: int = 4, rows: int = 3):
        self.source = image
        self.cols = cols
        self.rows = rows

        # Calculate piece dimensions
        self.piece_width = image.width // cols
        self.piece_height = image.height // rows

        # Tab/slot size
        self.tab_width = int(self.piece_width * self.TAB_SIZE_RATIO)
        self.tab_height = int(self.piece_height * self.TAB_SIZE_RATIO)

        # Pre-generate edge types for consistency
        self._edge_map = self._generate_edge_map()

    def _generate_edge_map(self) -> dict:
        """
        Pre-compute edge types for all pieces.
        Ensures adjacent pieces have matching tab/slot pairs.
        """
        edges = {}
        random.seed(42)  # Reproducible puzzles

        # Horizontal edges (between rows)
        for row in range(self.rows + 1):
            for col in range(self.cols):
                key = ('h', row, col)
                if row == 0 or row == self.rows:
                    edges[key] = EdgeType.FLAT
                else:
                    edges[key] = random.choice([EdgeType.TAB, EdgeType.SLOT])

        # Vertical edges (between columns)
        for row in range(self.rows):
            for col in range(self.cols + 1):
                key = ('v', row, col)
                if col == 0 or col == self.cols:
                    edges[key] = EdgeType.FLAT
                else:
                    edges[key] = random.choice([EdgeType.TAB, EdgeType.SLOT])

        return edges

    def _get_piece_edges(self, row: int, col: int) -> tuple[EdgeType, EdgeType, EdgeType, EdgeType]:
        """Get edge types for a specific piece (top, right, bottom, left)."""
        top = self._edge_map[('h', row, col)]
        bottom = self._edge_map[('h', row + 1, col)]
        left = self._edge_map[('v', row, col)]
        right = self._edge_map[('v', row, col + 1)]

        # Invert edges to match adjacent pieces
        # If the edge above has a TAB going down, we need a SLOT going up
        if top == EdgeType.TAB:
            top = EdgeType.SLOT
        elif top == EdgeType.SLOT:
            top = EdgeType.TAB

        if left == EdgeType.TAB:
            left = EdgeType.SLOT
        elif left == EdgeType.SLOT:
            left = EdgeType.TAB

        return (top, right, bottom, left)

    def generate(self) -> list[PuzzlePiece]:
        """Generate all puzzle pieces."""
        pieces = []

        for row in range(self.rows):
            for col in range(self.cols):
                piece = self._create_piece(row, col)
                pieces.append(piece)

        return pieces

    def _create_piece(self, row: int, col: int) -> PuzzlePiece:
        """Create a single puzzle piece with interlocking edges."""
        edges = self._get_piece_edges(row, col)

        # Padding for tabs - consistent for ALL pieces
        pad = max(self.tab_width, self.tab_height)

        # Consistent canvas size for all pieces
        canvas_width = self.piece_width + 2 * pad
        canvas_height = self.piece_height + 2 * pad

        # Source coordinates (where this piece's content starts in source image)
        src_x = col * self.piece_width
        src_y = row * self.piece_height

        # Calculate crop region from source (with padding where available)
        crop_x1 = src_x - pad
        crop_y1 = src_y - pad
        crop_x2 = src_x + self.piece_width + pad
        crop_y2 = src_y + self.piece_height + pad

        # Calculate where to paste on canvas (handles edge cases)
        paste_x = 0 if crop_x1 >= 0 else -crop_x1
        paste_y = 0 if crop_y1 >= 0 else -crop_y1

        # Clamp crop region to source bounds
        crop_x1 = max(0, crop_x1)
        crop_y1 = max(0, crop_y1)
        crop_x2 = min(self.source.width, crop_x2)
        crop_y2 = min(self.source.height, crop_y2)

        # Crop from source
        cropped = self.source.crop((crop_x1, crop_y1, crop_x2, crop_y2))

        # Create canvas and paste cropped region at correct position
        canvas = Image.new('RGB', (canvas_width, canvas_height), (255, 255, 255))
        canvas.paste(cropped, (paste_x, paste_y))

        # Create mask for piece shape (always same size with consistent padding)
        mask = self._create_piece_mask(
            canvas_width, canvas_height,
            pad, pad, pad, pad,  # Consistent padding on all sides
            edges
        )

        # Apply mask (convert to RGBA)
        piece_img = canvas.convert('RGBA')
        piece_img.putalpha(mask)

        return PuzzlePiece(
            image=piece_img,
            row=row,
            col=col,
            edges=edges
        )

    def _create_piece_mask(
        self,
        width: int,
        height: int,
        pad_left: int,
        pad_top: int,
        pad_right: int,
        pad_bottom: int,
        edges: tuple[EdgeType, EdgeType, EdgeType, EdgeType]
    ) -> Image.Image:
        """Create alpha mask for piece shape with tabs/slots."""
        mask = Image.new('L', (width, height), 0)
        draw = ImageDraw.Draw(mask)

        top, right, bottom, left = edges

        # Base rectangle coordinates (the main piece area)
        base_left = pad_left
        base_top = pad_top
        base_right = width - pad_right
        base_bottom = height - pad_bottom

        # Build polygon points for the piece shape
        points = []

        # Start at top-left corner
        points.append((base_left, base_top))

        # Top edge
        if top == EdgeType.FLAT:
            points.append((base_right, base_top))
        else:
            # Add tab or slot on top edge
            mid_x = (base_left + base_right) // 2
            tab_w = self.tab_width // 2
            tab_h = self.tab_height

            points.append((mid_x - tab_w, base_top))
            if top == EdgeType.TAB:
                # Tab goes up (outside)
                points.extend(self._bezier_tab_points(
                    mid_x - tab_w, base_top,
                    mid_x + tab_w, base_top,
                    -tab_h, 'horizontal'
                ))
            else:
                # Slot goes down (inside)
                points.extend(self._bezier_tab_points(
                    mid_x - tab_w, base_top,
                    mid_x + tab_w, base_top,
                    tab_h, 'horizontal'
                ))
            points.append((mid_x + tab_w, base_top))
            points.append((base_right, base_top))

        # Right edge
        if right == EdgeType.FLAT:
            points.append((base_right, base_bottom))
        else:
            mid_y = (base_top + base_bottom) // 2
            tab_w = self.tab_width
            tab_h = self.tab_height // 2

            points.append((base_right, mid_y - tab_h))
            if right == EdgeType.TAB:
                points.extend(self._bezier_tab_points(
                    base_right, mid_y - tab_h,
                    base_right, mid_y + tab_h,
                    tab_w, 'vertical'
                ))
            else:
                points.extend(self._bezier_tab_points(
                    base_right, mid_y - tab_h,
                    base_right, mid_y + tab_h,
                    -tab_w, 'vertical'
                ))
            points.append((base_right, mid_y + tab_h))
            points.append((base_right, base_bottom))

        # Bottom edge (right to left)
        if bottom == EdgeType.FLAT:
            points.append((base_left, base_bottom))
        else:
            mid_x = (base_left + base_right) // 2
            tab_w = self.tab_width // 2
            tab_h = self.tab_height

            points.append((mid_x + tab_w, base_bottom))
            if bottom == EdgeType.TAB:
                points.extend(self._bezier_tab_points(
                    mid_x + tab_w, base_bottom,
                    mid_x - tab_w, base_bottom,
                    tab_h, 'horizontal'
                ))
            else:
                points.extend(self._bezier_tab_points(
                    mid_x + tab_w, base_bottom,
                    mid_x - tab_w, base_bottom,
                    -tab_h, 'horizontal'
                ))
            points.append((mid_x - tab_w, base_bottom))
            points.append((base_left, base_bottom))

        # Left edge (bottom to top)
        if left == EdgeType.FLAT:
            pass  # Will close back to start
        else:
            mid_y = (base_top + base_bottom) // 2
            tab_w = self.tab_width
            tab_h = self.tab_height // 2

            points.append((base_left, mid_y + tab_h))
            if left == EdgeType.TAB:
                points.extend(self._bezier_tab_points(
                    base_left, mid_y + tab_h,
                    base_left, mid_y - tab_h,
                    -tab_w, 'vertical'
                ))
            else:
                points.extend(self._bezier_tab_points(
                    base_left, mid_y + tab_h,
                    base_left, mid_y - tab_h,
                    tab_w, 'vertical'
                ))
            points.append((base_left, mid_y - tab_h))

        # Draw filled polygon
        draw.polygon(points, fill=255)

        return mask

    def _bezier_tab_points(
        self,
        x1: int, y1: int,
        x2: int, y2: int,
        offset: int,
        direction: str
    ) -> list[tuple[int, int]]:
        """
        Generate points for a curved tab/slot edge.
        Uses a simplified approach with control points.
        """
        points = []
        steps = 12  # Number of points for the curve

        if direction == 'horizontal':
            # Tab extends perpendicular to x-axis
            mid_x = (x1 + x2) // 2

            for i in range(1, steps):
                t = i / steps
                # Simple bulge curve
                x = x1 + (x2 - x1) * t
                # Sine-based bulge
                bulge = offset * (4 * t * (1 - t))  # Parabolic bulge
                y = y1 + bulge
                points.append((int(x), int(y)))
        else:
            # Tab extends perpendicular to y-axis
            mid_y = (y1 + y2) // 2

            for i in range(1, steps):
                t = i / steps
                y = y1 + (y2 - y1) * t
                bulge = offset * (4 * t * (1 - t))
                x = x1 + bulge
                points.append((int(x), int(y)))

        return points
