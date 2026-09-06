#!/usr/bin/env python3
"""Minimal ink/brass mark for Electron — no mascot."""

from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

OUT = Path(__file__).resolve().parents[1] / "apps" / "desktop" / "icon.png"

SIZE = 1024
img = Image.new("RGB", (SIZE, SIZE), "#0E1114")
draw = ImageDraw.Draw(img)
margin = 96
draw.rounded_rectangle(
    [margin, margin, SIZE - margin, SIZE - margin],
    radius=96,
    outline="#C4A574",
    width=18,
)
try:
    font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 520)
except OSError:
    font = ImageFont.load_default()
text = "C"
bbox = draw.textbbox((0, 0), text, font=font)
tw, th = bbox[2] - bbox[0], bbox[3] - bbox[1]
draw.text(((SIZE - tw) / 2 - bbox[0], (SIZE - th) / 2 - bbox[1] - 20), text, font=font, fill="#C4A574")
img.save(OUT, "PNG")
print(f"wrote {OUT}")
