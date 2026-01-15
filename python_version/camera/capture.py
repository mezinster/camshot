"""
Camera capture module using OpenCV.
Cross-platform webcam support.
"""

from PIL import Image
import numpy as np

# OpenCV import with graceful fallback
try:
    import cv2
    OPENCV_AVAILABLE = True
except ImportError:
    OPENCV_AVAILABLE = False


def capture_image(camera_index: int = 0) -> Image.Image | None:
    """
    Capture a single frame from the webcam.

    Args:
        camera_index: Camera device index (0 for default camera)

    Returns:
        PIL Image or None if capture failed
    """
    if not OPENCV_AVAILABLE:
        raise ImportError("OpenCV is not installed. Run: pip install opencv-python")

    cap = cv2.VideoCapture(camera_index)

    if not cap.isOpened():
        return None

    try:
        # Allow camera to warm up
        for _ in range(5):
            cap.read()

        # Capture frame
        ret, frame = cap.read()

        if not ret or frame is None:
            return None

        # Convert BGR (OpenCV) to RGB (PIL)
        frame_rgb = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)

        # Convert to PIL Image
        image = Image.fromarray(frame_rgb)

        return image

    finally:
        cap.release()


def list_cameras(max_cameras: int = 5) -> list[int]:
    """
    List available camera indices.

    Returns:
        List of working camera indices
    """
    if not OPENCV_AVAILABLE:
        return []

    available = []

    for i in range(max_cameras):
        cap = cv2.VideoCapture(i)
        if cap.isOpened():
            available.append(i)
            cap.release()

    return available


class CameraPreview:
    """
    Live camera preview for Tkinter integration.

    Usage:
        preview = CameraPreview()
        preview.start()
        # In your update loop:
        frame = preview.get_frame()
        # When done:
        preview.stop()
    """

    def __init__(self, camera_index: int = 0):
        self.camera_index = camera_index
        self.cap = None
        self.running = False

    def start(self) -> bool:
        """Start camera capture."""
        if not OPENCV_AVAILABLE:
            return False

        self.cap = cv2.VideoCapture(self.camera_index)
        self.running = self.cap.isOpened()
        return self.running

    def stop(self):
        """Stop camera capture."""
        self.running = False
        if self.cap:
            self.cap.release()
            self.cap = None

    def get_frame(self) -> Image.Image | None:
        """Get current camera frame as PIL Image."""
        if not self.running or not self.cap:
            return None

        ret, frame = self.cap.read()

        if not ret or frame is None:
            return None

        frame_rgb = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
        return Image.fromarray(frame_rgb)
