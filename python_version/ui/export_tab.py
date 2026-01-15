"""
Export tab - save puzzle as PNG files or PDF.
"""

import tkinter as tk
from tkinter import ttk, filedialog, messagebox
from PIL import Image, ImageTk
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from ui.app import CamShotApp


class ExportTab:
    """Tab for exporting puzzle pieces."""

    PREVIEW_PIECE_SIZE = 80

    def __init__(self, parent: ttk.Notebook, app: 'CamShotApp'):
        self.app = app
        self.frame = ttk.Frame(parent, padding="20")
        self.piece_photos = []  # Keep references

        self._create_widgets()

    def _create_widgets(self):
        """Create the tab's UI components."""
        # Export options
        options_frame = ttk.LabelFrame(self.frame, text="Export Options", padding="15")
        options_frame.pack(fill=tk.X, pady=(0, 20))

        # Export buttons
        btn_frame = ttk.Frame(options_frame)
        btn_frame.pack(fill=tk.X)

        self.export_png_btn = ttk.Button(
            btn_frame,
            text="Export as PNG Files",
            style='Action.TButton',
            command=self._export_png
        )
        self.export_png_btn.pack(side=tk.LEFT, padx=10)

        self.export_pdf_btn = ttk.Button(
            btn_frame,
            text="Export as PDF (A4)",
            style='Action.TButton',
            command=self._export_pdf
        )
        self.export_pdf_btn.pack(side=tk.LEFT, padx=10)

        # Preview section
        preview_frame = ttk.LabelFrame(self.frame, text="Generated Pieces Preview", padding="10")
        preview_frame.pack(fill=tk.BOTH, expand=True)

        # Canvas with scrollbar for pieces
        canvas_frame = ttk.Frame(preview_frame)
        canvas_frame.pack(fill=tk.BOTH, expand=True)

        self.preview_canvas = tk.Canvas(
            canvas_frame,
            bg='#e0e0e0',
            highlightthickness=0
        )

        scrollbar_y = ttk.Scrollbar(canvas_frame, orient=tk.VERTICAL, command=self.preview_canvas.yview)
        scrollbar_x = ttk.Scrollbar(preview_frame, orient=tk.HORIZONTAL, command=self.preview_canvas.xview)

        self.preview_canvas.configure(
            yscrollcommand=scrollbar_y.set,
            xscrollcommand=scrollbar_x.set
        )

        scrollbar_y.pack(side=tk.RIGHT, fill=tk.Y)
        self.preview_canvas.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        scrollbar_x.pack(fill=tk.X)

        # Info label
        self.info_var = tk.StringVar(value="No pieces generated yet")
        ttk.Label(
            self.frame,
            textvariable=self.info_var,
            font=('Helvetica', 10)
        ).pack(pady=(10, 0))

    def update_preview(self):
        """Update preview with generated pieces."""
        self.preview_canvas.delete("all")
        self.piece_photos.clear()

        if not self.app.puzzle_pieces:
            self.info_var.set("No pieces generated yet")
            return

        pieces = self.app.puzzle_pieces
        self.info_var.set(f"{len(pieces)} pieces ready for export")

        # Calculate layout
        padding = 10
        piece_size = self.PREVIEW_PIECE_SIZE
        cols = 6  # Pieces per row in preview

        for idx, piece in enumerate(pieces):
            row = idx // cols
            col = idx % cols

            x = padding + col * (piece_size + padding)
            y = padding + row * (piece_size + padding)

            # Scale piece for preview
            thumb = piece.image.copy()
            thumb.thumbnail((piece_size, piece_size), Image.Resampling.LANCZOS)

            # Create checkered background for transparency
            bg = self._create_checkered_bg(thumb.width, thumb.height)
            bg.paste(thumb, (0, 0), thumb)

            photo = ImageTk.PhotoImage(bg)
            self.piece_photos.append(photo)

            self.preview_canvas.create_image(x, y, anchor=tk.NW, image=photo)

            # Label with piece position
            self.preview_canvas.create_text(
                x + piece_size // 2,
                y + piece_size + 5,
                text=f"({piece.row},{piece.col})",
                font=('Helvetica', 8),
                fill='#666'
            )

        # Update scroll region
        rows = (len(pieces) + cols - 1) // cols
        total_height = rows * (piece_size + padding + 20) + padding
        total_width = cols * (piece_size + padding) + padding
        self.preview_canvas.configure(scrollregion=(0, 0, total_width, total_height))

    def _create_checkered_bg(self, width: int, height: int, cell_size: int = 10) -> Image.Image:
        """Create a checkered background for transparency preview."""
        bg = Image.new('RGB', (width, height))
        colors = [(200, 200, 200), (255, 255, 255)]

        for y in range(0, height, cell_size):
            for x in range(0, width, cell_size):
                color_idx = ((x // cell_size) + (y // cell_size)) % 2
                for dy in range(min(cell_size, height - y)):
                    for dx in range(min(cell_size, width - x)):
                        bg.putpixel((x + dx, y + dy), colors[color_idx])

        return bg

    def _export_png(self):
        """Export pieces as individual PNG files."""
        if not self.app.puzzle_pieces:
            messagebox.showwarning("Export", "No puzzle pieces to export!")
            return

        # Ask for directory
        directory = filedialog.askdirectory(title="Select Export Directory")
        if not directory:
            return

        try:
            export_path = Path(directory)

            for piece in self.app.puzzle_pieces:
                filename = f"piece_{piece.row}_{piece.col}.png"
                piece.image.save(export_path / filename)

            self.app.set_status(f"Exported {len(self.app.puzzle_pieces)} pieces to {directory}")
            messagebox.showinfo(
                "Export Complete",
                f"Successfully exported {len(self.app.puzzle_pieces)} pieces!"
            )

        except Exception as e:
            messagebox.showerror("Export Error", f"Failed to export:\n{e}")

    def _export_pdf(self):
        """Export pieces arranged on A4 PDF."""
        if not self.app.puzzle_pieces:
            messagebox.showwarning("Export", "No puzzle pieces to export!")
            return

        # Ask for file path
        filepath = filedialog.asksaveasfilename(
            title="Save PDF",
            defaultextension=".pdf",
            filetypes=[("PDF files", "*.pdf"), ("All files", "*.*")]
        )

        if not filepath:
            return

        try:
            from export.pdf import export_to_pdf
            export_to_pdf(self.app.puzzle_pieces, filepath)

            self.app.set_status(f"Exported PDF to {filepath}")
            messagebox.showinfo("Export Complete", f"Successfully exported PDF!")

        except ImportError:
            messagebox.showerror(
                "Export Error",
                "reportlab is required for PDF export.\n"
                "Install with: pip install reportlab"
            )
        except Exception as e:
            messagebox.showerror("Export Error", f"Failed to export PDF:\n{e}")
