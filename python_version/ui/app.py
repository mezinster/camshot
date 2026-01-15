"""
Main application window with tabbed interface.
"""

import tkinter as tk
from tkinter import ttk
from PIL import Image

from ui.image_tab import ImageTab
from ui.puzzle_tab import PuzzleTab
from ui.export_tab import ExportTab


class CamShotApp:
    """Main application class managing the UI and state."""

    def __init__(self, root: tk.Tk):
        self.root = root
        self.root.title("CamShot - Photo to Puzzle")
        self.root.geometry("900x700")
        self.root.minsize(800, 600)

        # Application state
        self.current_image: Image.Image | None = None
        self.puzzle_pieces: list = []
        self.grid_size = (4, 3)  # columns x rows

        self._setup_styles()
        self._create_widgets()

    def _setup_styles(self):
        """Configure ttk styles for a cleaner look."""
        style = ttk.Style()
        style.theme_use('clam')  # Modern-ish theme available on all platforms

        # Custom button style
        style.configure('Action.TButton', padding=10, font=('Helvetica', 10))

    def _create_widgets(self):
        """Create the main UI components."""
        # Main container
        self.main_frame = ttk.Frame(self.root, padding="10")
        self.main_frame.pack(fill=tk.BOTH, expand=True)

        # Header
        header = ttk.Label(
            self.main_frame,
            text="CamShot",
            font=('Helvetica', 24, 'bold')
        )
        header.pack(pady=(0, 10))

        # Notebook (tabbed interface)
        self.notebook = ttk.Notebook(self.main_frame)
        self.notebook.pack(fill=tk.BOTH, expand=True)

        # Create tabs
        self.image_tab = ImageTab(self.notebook, self)
        self.puzzle_tab = PuzzleTab(self.notebook, self)
        self.export_tab = ExportTab(self.notebook, self)

        self.notebook.add(self.image_tab.frame, text="1. Load Image")
        self.notebook.add(self.puzzle_tab.frame, text="2. Create Puzzle")
        self.notebook.add(self.export_tab.frame, text="3. Export")

        # Status bar
        self.status_var = tk.StringVar(value="Ready - Load an image to begin")
        status_bar = ttk.Label(
            self.main_frame,
            textvariable=self.status_var,
            relief=tk.SUNKEN,
            anchor=tk.W,
            padding=(5, 2)
        )
        status_bar.pack(fill=tk.X, pady=(10, 0))

    def set_status(self, message: str):
        """Update the status bar message."""
        self.status_var.set(message)

    def set_image(self, image: Image.Image):
        """Set the current working image."""
        self.current_image = image
        self.puzzle_pieces = []  # Clear old pieces
        self.set_status(f"Image loaded: {image.width}x{image.height}")

        # Update puzzle tab preview
        self.puzzle_tab.update_preview()

        # Switch to puzzle tab
        self.notebook.select(1)

    def set_puzzle_pieces(self, pieces: list):
        """Store generated puzzle pieces."""
        self.puzzle_pieces = pieces
        self.set_status(f"Generated {len(pieces)} puzzle pieces")

        # Update export tab
        self.export_tab.update_preview()

        # Switch to export tab
        self.notebook.select(2)
