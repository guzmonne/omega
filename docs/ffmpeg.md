# ffmpeg

## Build

```bash
rm -Rf /tmp/*.png | true && \
  rm -Rf /tmp/*.jpg | true && \
  make install && \
  omega chrome record -W 480 -H 320 -d 10000 -w 4
```

## Commands

Create an output video file with a valid aplpha channel that can be used
inside daVinci Resolve

```bash
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.png' \
  -codec prores_ks \
  -pix_fmt yuva444p10le \
  -alpha_bits 16 \
  -profile:v 4444 \
  -f mov /tmp/output.mov
```

A 1 second video weights 2.2MB aprox.

Same command as before but using `VP9` codec and a `webm` output file.

```bash
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.png' \
  -c:v vp9 \
  -pix_fmt yuva420p \
  /tmp/output.webm
```

Using two passes

```bash
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.png' \
  -c:v vp9 \
  -pass 1 \
  -pix_fmt yuva420p \
  -f webm NUL
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.png' \
  -c:v vp9 \
  -pass 2 \
  -pix_fmt yuva420p \
  /tmp/output.webm
```

Using ProRes4444 with Alpha Channel

```bash
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.png' \
  -c:v prores_ks \
  -pix_fmt yuva444p10le \
  /tmp/output.mov
```

Same command as before but using `VP9` codec and a `mp4` output file.

```bash
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.png' \
  -c:v vp9 \
  -pix_fmt yuva420p \
  -r 60 \
  /tmp/output.webm
```

PNG to MP4

```bash
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.png' \
  -c:v libx264 \
  -vf "fps=60,format=yuv420p" \
  /tmp/out.mp4
```

JPG to MP4

```bash
ffmpeg -y \
  -framerate 60 \
  -i '/tmp/%6d.jpg' \
  -c:v libx264 \
  -vf "fps=60,format=yuv420p" \
  /tmp/out.mp4
```