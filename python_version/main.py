#!/usr/bin/env python3
"""
CamShot - Photo to Jigsaw Puzzle Converter
A desktop application that converts photos into printable jigsaw puzzles.
"""

import tkinter as tk
from ui.app import CamShotApp


def main():
    root = tk.Tk()
    app = CamShotApp(root)
    root.mainloop()


if __name__ == "__main__":
    main()
