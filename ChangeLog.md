
# 0.7.5
- Converted code to entriely go (with a tiny bit of c). THis simplifies the build process and reduces dependencies. This lets us easily port to new platforms like arm64.

# 0.5.4
- Added environment variables to pass through chafa options.
- Added environment variable `TERM_EVERYTHING_PIXEL_TYPE` to workaround certain cases where there is a bgr/rgb swap.
- Added `--max-frame-rate` option to ...set the maximum frame rate.
- Fixed a build bug with node_canvas's included version of libstdc++.so
# 0.5.3
fixed right mouse click and sub menus
Set XDG_SESSION_TYPE to wayland so app that have
a wayland support will use it.
# 0.5.2
fixed scrolling inversion https://github.com/mmulet/term.everything/issues/4
# 0.5.0
First Beta release
