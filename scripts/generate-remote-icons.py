#!/usr/bin/env python3
"""Generate bright-yellow remote-mode icon variants from the existing green logo."""

import os
import sys
from PIL import Image

SRC_SIZES = [
    ("build/icons/256x256.png", "build/icons/remote/256x256.png"),
    ("build/icons/512x512.png", "build/icons/remote/512x512.png"),
    ("build/icons/128x128.png", "build/icons/remote/128x128.png"),
    ("build/icons/64x64.png",   "build/icons/remote/64x64.png"),
    ("build/icons/48x48.png",   "build/icons/remote/48x48.png"),
    ("build/icons/32x32.png",   "build/icons/remote/32x32.png"),
    ("build/icons/16x16.png",   "build/icons/remote/16x16.png"),
    ("build/icons/256x256.png", "public/logos/dora-logo-remote-256.png"),
    ("build/icons/512x512.png", "public/logos/dora-logo-remote-512.png"),
]

# Bright yellow gradient: lit highlight → shaded shadow
YELLOW_BRIGHT = (255, 218, 0)   # top of range
YELLOW_DARK   = (168, 122, 0)   # bottom of range (stays amber, not olive)
GAMMA         = 0.55            # < 1 boosts dark areas upward

def colorize_yellow(img: Image.Image) -> Image.Image:
    """Re-color the icon to bright yellow using gamma-corrected luminance as shading."""
    src = img.convert("RGBA")
    out = Image.new("RGBA", src.size, (0, 0, 0, 0))

    for y in range(src.height):
        for x in range(src.width):
            r, g, b, a = src.getpixel((x, y))
            if a == 0:
                out.putpixel((x, y), (0, 0, 0, 0))
                continue

            # Perceived luminance (0-1)
            lum = (0.299 * r + 0.587 * g + 0.114 * b) / 255.0

            # Near-white background → keep unchanged
            if lum > 0.92:
                out.putpixel((x, y), (r, g, b, a))
                continue

            # Gamma curve lifts dark pixels toward mid-range so the icon
            # stays visibly yellow even in shadow areas.
            t = lum ** GAMMA

            nr = int(YELLOW_DARK[0] + t * (YELLOW_BRIGHT[0] - YELLOW_DARK[0]))
            ng = int(YELLOW_DARK[1] + t * (YELLOW_BRIGHT[1] - YELLOW_DARK[1]))
            nb = int(YELLOW_DARK[2] + t * (YELLOW_BRIGHT[2] - YELLOW_DARK[2]))
            out.putpixel((x, y), (nr, ng, nb, a))

    return out

def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    root = os.path.dirname(script_dir)

    for src_rel, dst_rel in SRC_SIZES:
        src_path = os.path.join(root, src_rel)
        dst_path = os.path.join(root, dst_rel)
        os.makedirs(os.path.dirname(dst_path), exist_ok=True)

        if not os.path.exists(src_path):
            print(f"  skip (missing): {src_path}", file=sys.stderr)
            continue

        img = Image.open(src_path)
        yellow = colorize_yellow(img)
        yellow.save(dst_path, "PNG")
        print(f"  wrote {dst_path}")

if __name__ == "__main__":
    main()
