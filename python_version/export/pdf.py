"""
PDF export module - arranges puzzle pieces on A4 pages.
"""

from reportlab.lib.pagesizes import A4
from reportlab.pdfgen import canvas
from reportlab.lib.utils import ImageReader
from PIL import Image
import io


# A4 dimensions in points (72 points per inch)
A4_WIDTH, A4_HEIGHT = A4
MARGIN = 36  # 0.5 inch margin


def export_to_pdf(pieces: list, filepath: str, pieces_per_row: int = 4):
    """
    Export puzzle pieces to a multi-page PDF.

    Args:
        pieces: List of PuzzlePiece objects
        filepath: Output PDF path
        pieces_per_row: Number of pieces per row on each page
    """
    c = canvas.Canvas(filepath, pagesize=A4)

    # Calculate piece size to fit page
    usable_width = A4_WIDTH - (2 * MARGIN)
    usable_height = A4_HEIGHT - (2 * MARGIN)

    # Find max piece dimensions
    max_piece_w = max(p.image.width for p in pieces)
    max_piece_h = max(p.image.height for p in pieces)
    piece_aspect = max_piece_w / max_piece_h

    # Calculate cell size
    cell_width = usable_width / pieces_per_row
    cell_height = cell_width / piece_aspect

    # How many rows fit per page?
    rows_per_page = int(usable_height / cell_height)
    if rows_per_page < 1:
        rows_per_page = 1
        cell_height = usable_height
        cell_width = cell_height * piece_aspect

    pieces_per_page = pieces_per_row * rows_per_page

    # Process pieces
    for page_idx in range((len(pieces) + pieces_per_page - 1) // pieces_per_page):
        if page_idx > 0:
            c.showPage()

        # Add page header
        c.setFont("Helvetica-Bold", 14)
        c.drawString(MARGIN, A4_HEIGHT - 30, f"CamShot Puzzle - Page {page_idx + 1}")

        # Draw pieces on this page
        start_idx = page_idx * pieces_per_page
        end_idx = min(start_idx + pieces_per_page, len(pieces))

        for idx in range(start_idx, end_idx):
            piece = pieces[idx]
            local_idx = idx - start_idx

            row = local_idx // pieces_per_row
            col = local_idx % pieces_per_row

            # Calculate position (reportlab uses bottom-left origin)
            x = MARGIN + col * cell_width
            y = A4_HEIGHT - MARGIN - 40 - (row + 1) * cell_height  # 40 for header

            # Scale piece to fit cell (with some padding)
            padding = 5
            target_w = cell_width - padding * 2
            target_h = cell_height - padding * 2

            # Convert RGBA to RGB with white background for PDF
            piece_img = piece.image
            if piece_img.mode == 'RGBA':
                background = Image.new('RGB', piece_img.size, (255, 255, 255))
                background.paste(piece_img, mask=piece_img.split()[3])
                piece_img = background

            # Save to bytes buffer
            img_buffer = io.BytesIO()
            piece_img.save(img_buffer, format='PNG')
            img_buffer.seek(0)

            # Draw image
            c.drawImage(
                ImageReader(img_buffer),
                x + padding,
                y + padding,
                width=target_w,
                height=target_h,
                preserveAspectRatio=True,
                anchor='c'
            )

            # Add piece label
            c.setFont("Helvetica", 8)
            c.drawString(
                x + cell_width / 2 - 15,
                y + 2,
                f"({piece.row}, {piece.col})"
            )

    c.save()
