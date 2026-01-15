"""
Image loading tab - handles file upload and camera capture.
"""

import tkinter as tk
from tkinter import ttk, filedialog, messagebox
from PIL import Image, ImageTk
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from ui.app import CamShotApp


class ImageTab:
    """Tab for loading images from file or camera."""

    MAX_PREVIEW_SIZE = (500, 400)

    def __init__(self, parent: ttk.Notebook, app: 'CamShotApp'):
        self.app = app
        self.frame = ttk.Frame(parent, padding="20")
        self.preview_photo = None  # Keep reference to prevent garbage collection

        self._create_widgets()

    def _create_widgets(self):
        """Create the tab's UI components."""
        # Instructions
        instructions = ttk.Label(
            self.frame,
            text="Load an image to convert into a jigsaw puzzle",
            font=('Helvetica', 12)
        )
        instructions.pack(pady=(0, 20))

        # Button frame
        btn_frame = ttk.Frame(self.frame)
        btn_frame.pack(pady=10)

        # Load from file button
        self.load_btn = ttk.Button(
            btn_frame,
            text="Load from File",
            style='Action.TButton',
            command=self._load_from_file
        )
        self.load_btn.pack(side=tk.LEFT, padx=10)

        # Camera capture button
        self.camera_btn = ttk.Button(
            btn_frame,
            text="Capture from Camera",
            style='Action.TButton',
            command=self._capture_from_camera
        )
        self.camera_btn.pack(side=tk.LEFT, padx=10)

        # Preview frame
        preview_frame = ttk.LabelFrame(self.frame, text="Preview", padding="10")
        preview_frame.pack(fill=tk.BOTH, expand=True, pady=20)

        # Preview label (will hold the image)
        self.preview_label = ttk.Label(
            preview_frame,
            text="No image loaded",
            anchor=tk.CENTER
        )
        self.preview_label.pack(fill=tk.BOTH, expand=True)

        # Image info
        self.info_var = tk.StringVar(value="")
        self.info_label = ttk.Label(
            self.frame,
            textvariable=self.info_var,
            font=('Helvetica', 10)
        )
        self.info_label.pack()

    def _load_from_file(self):
        """Open file dialog and load selected image."""
        filetypes = [
            ("Image files", "*.png *.jpg *.jpeg *.gif *.bmp *.webp"),
            ("PNG files", "*.png"),
            ("JPEG files", "*.jpg *.jpeg"),
            ("All files", "*.*")
        ]

        filepath = filedialog.askopenfilename(
            title="Select an image",
            filetypes=filetypes
        )

        if filepath:
            try:
                image = Image.open(filepath)
                # Convert to RGB if necessary (handles RGBA, palette, etc.)
                if image.mode != 'RGB':
                    image = image.convert('RGB')

                self._display_preview(image)
                self.app.set_image(image)

            except Exception as e:
                messagebox.showerror("Error", f"Failed to load image:\n{e}")

    def _capture_from_camera(self):
        """Capture image from webcam using OpenCV."""
        try:
            from camera.capture import capture_image
            image = capture_image()

            if image:
                self._display_preview(image)
                self.app.set_image(image)
            else:
                messagebox.showwarning("Camera", "Failed to capture image from camera")

        except ImportError as e:
            messagebox.showerror(
                "Camera Error",
                "OpenCV is required for camera capture.\n"
                "Install with: pip install opencv-python"
            )
        except Exception as e:
            messagebox.showerror("Camera Error", f"Camera capture failed:\n{e}")

    def _display_preview(self, image: Image.Image):
        """Display image preview, scaled to fit."""
        # Calculate scaling to fit preview area
        img_ratio = image.width / image.height
        max_w, max_h = self.MAX_PREVIEW_SIZE

        if img_ratio > max_w / max_h:
            # Width is limiting factor
            new_width = min(image.width, max_w)
            new_height = int(new_width / img_ratio)
        else:
            # Height is limiting factor
            new_height = min(image.height, max_h)
            new_width = int(new_height * img_ratio)

        # Create preview image
        preview = image.copy()
        preview.thumbnail((new_width, new_height), Image.Resampling.LANCZOS)

        # Convert to PhotoImage (Tkinter-compatible)
        self.preview_photo = ImageTk.PhotoImage(preview)

        # Update label
        self.preview_label.configure(image=self.preview_photo, text="")

        # Update info
        self.info_var.set(f"Size: {image.width} x {image.height} pixels")
