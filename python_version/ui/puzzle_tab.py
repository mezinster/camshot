"""
Puzzle creation tab - configure grid size and generate pieces.
"""

import tkinter as tk
from tkinter import ttk
from PIL import Image, ImageTk
from typing import TYPE_CHECKING

from puzzle.generator import PuzzleGenerator

if TYPE_CHECKING:
    from ui.app import CamShotApp


class PuzzleTab:
    """Tab for configuring and generating puzzle pieces."""

    MAX_PREVIEW_SIZE = (500, 400)

    def __init__(self, parent: ttk.Notebook, app: 'CamShotApp'):
        self.app = app
        self.frame = ttk.Frame(parent, padding="20")
        self.preview_photo = None

        self._create_widgets()

    def _create_widgets(self):
        """Create the tab's UI components."""
        # Top section: Grid configuration
        config_frame = ttk.LabelFrame(self.frame, text="Puzzle Configuration", padding="15")
        config_frame.pack(fill=tk.X, pady=(0, 20))

        # Grid size controls
        grid_frame = ttk.Frame(config_frame)
        grid_frame.pack(fill=tk.X)

        # Columns
        ttk.Label(grid_frame, text="Columns:").pack(side=tk.LEFT, padx=(0, 5))
        self.cols_var = tk.IntVar(value=4)
        cols_spin = ttk.Spinbox(
            grid_frame,
            from_=2, to=10,
            textvariable=self.cols_var,
            width=5,
            command=self._on_grid_change
        )
        cols_spin.pack(side=tk.LEFT, padx=(0, 20))

        # Rows
        ttk.Label(grid_frame, text="Rows:").pack(side=tk.LEFT, padx=(0, 5))
        self.rows_var = tk.IntVar(value=3)
        rows_spin = ttk.Spinbox(
            grid_frame,
            from_=2, to=10,
            textvariable=self.rows_var,
            width=5,
            command=self._on_grid_change
        )
        rows_spin.pack(side=tk.LEFT, padx=(0, 20))

        # Piece count info
        self.piece_count_var = tk.StringVar(value="12 pieces")
        ttk.Label(
            grid_frame,
            textvariable=self.piece_count_var,
            font=('Helvetica', 10, 'italic')
        ).pack(side=tk.LEFT, padx=20)

        # Generate button
        self.generate_btn = ttk.Button(
            config_frame,
            text="Generate Puzzle",
            style='Action.TButton',
            command=self._generate_puzzle
        )
        self.generate_btn.pack(pady=(15, 0))

        # Preview area
        preview_frame = ttk.LabelFrame(self.frame, text="Preview (Grid Overlay)", padding="10")
        preview_frame.pack(fill=tk.BOTH, expand=True)

        self.preview_canvas = tk.Canvas(
            preview_frame,
            bg='#f0f0f0',
            highlightthickness=0
        )
        self.preview_canvas.pack(fill=tk.BOTH, expand=True)

    def _on_grid_change(self):
        """Handle grid size change."""
        cols = self.cols_var.get()
        rows = self.rows_var.get()
        total = cols * rows
        self.piece_count_var.set(f"{total} pieces")
        self.app.grid_size = (cols, rows)
        self.update_preview()

    def update_preview(self):
        """Update the preview with grid overlay."""
        if not self.app.current_image:
            return

        image = self.app.current_image
        cols = self.cols_var.get()
        rows = self.rows_var.get()

        # Calculate scaling
        canvas_w = self.preview_canvas.winfo_width() or 500
        canvas_h = self.preview_canvas.winfo_height() or 400

        img_ratio = image.width / image.height
        canvas_ratio = canvas_w / canvas_h

        if img_ratio > canvas_ratio:
            new_w = min(image.width, canvas_w - 20)
            new_h = int(new_w / img_ratio)
        else:
            new_h = min(image.height, canvas_h - 20)
            new_w = int(new_h * img_ratio)

        # Create preview with grid
        preview = image.copy()
        preview.thumbnail((new_w, new_h), Image.Resampling.LANCZOS)

        self.preview_photo = ImageTk.PhotoImage(preview)

        # Clear canvas and draw image
        self.preview_canvas.delete("all")

        # Center the image
        x_offset = (canvas_w - new_w) // 2
        y_offset = (canvas_h - new_h) // 2

        self.preview_canvas.create_image(
            x_offset, y_offset,
            anchor=tk.NW,
            image=self.preview_photo
        )

        # Draw grid lines
        piece_w = new_w / cols
        piece_h = new_h / rows

        # Vertical lines
        for i in range(1, cols):
            x = x_offset + int(i * piece_w)
            self.preview_canvas.create_line(
                x, y_offset, x, y_offset + new_h,
                fill='#ff6b6b', width=2, dash=(4, 4)
            )

        # Horizontal lines
        for i in range(1, rows):
            y = y_offset + int(i * piece_h)
            self.preview_canvas.create_line(
                x_offset, y, x_offset + new_w, y,
                fill='#ff6b6b', width=2, dash=(4, 4)
            )

        # Border
        self.preview_canvas.create_rectangle(
            x_offset, y_offset,
            x_offset + new_w, y_offset + new_h,
            outline='#ff6b6b', width=2
        )

    def _generate_puzzle(self):
        """Generate puzzle pieces from current image."""
        if not self.app.current_image:
            self.app.set_status("No image loaded!")
            return

        self.app.set_status("Generating puzzle pieces...")
        self.generate_btn.configure(state='disabled')
        self.frame.update()

        try:
            generator = PuzzleGenerator(
                self.app.current_image,
                cols=self.cols_var.get(),
                rows=self.rows_var.get()
            )
            pieces = generator.generate()
            self.app.set_puzzle_pieces(pieces)

        except Exception as e:
            self.app.set_status(f"Error: {e}")

        finally:
            self.generate_btn.configure(state='normal')
