// genicon makes the 256x256 app icon PNG; go-winres turns it into the
// multi-size .ico embedded in the exe. It also writes a small-size .ico next
// to it (same name, .ico) for the tray icon. Given the academy logo it cuts out the
// round medallion (transparent outside the circle); without one it draws a
// gold "S" on a navy disc.
//
//	go run ./tools/genicon build/winres/icon.png [build/winres/logo.png]
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const size = 256

var (
	navy = color.RGBA{0x14, 0x23, 0x4b, 0xff}
	gold = color.RGBA{0xf2, 0xb7, 0x05, 0xff}
)

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		log.Fatal("usage: genicon out.png [logo.png]")
	}
	var img *image.RGBA
	if len(os.Args) == 3 {
		img = medallion(os.Args[2])
	} else {
		img = drawnS()
	}
	out, err := os.Create(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		log.Fatal(err)
	}
	ico := strings.TrimSuffix(os.Args[1], ".png") + ".ico"
	if err := os.WriteFile(ico, smallICO(img, 16, 20, 24, 32, 40, 48, 64), 0o644); err != nil {
		log.Fatal(err)
	}
}

// smallICO scales img to each size and packs the PNGs into one .ico file
// (PNG entries are supported since Windows Vista).
func smallICO(img image.Image, sizes ...int) []byte {
	var pngs [][]byte
	for _, n := range sizes {
		dst := image.NewRGBA(image.Rect(0, 0, n, n))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
		var b bytes.Buffer
		if err := png.Encode(&b, dst); err != nil {
			log.Fatal(err)
		}
		pngs = append(pngs, b.Bytes())
	}
	var out bytes.Buffer
	le := func(v any) { binary.Write(&out, binary.LittleEndian, v) }
	le([3]uint16{0, 1, uint16(len(sizes))}) // ICONDIR: reserved, type icon, count
	off := 6 + 16*len(sizes)
	for i, n := range sizes {
		le([4]uint8{uint8(n), uint8(n), 0, 0}) // width, height, colours, reserved
		le([2]uint16{1, 32})                   // planes, bits per pixel
		le([2]uint32{uint32(len(pngs[i])), uint32(off)})
		off += len(pngs[i])
	}
	for _, b := range pngs {
		out.Write(b)
	}
	return out.Bytes()
}

// Where the medallion sits in logo.png (1254x1254): centre and radius,
// including its gold rim.
const (
	medalX, medalY, medalR = 627, 562, 368
)

// medallion crops the circular medallion out of the logo, scales it to the
// icon size and makes everything outside the circle transparent.
func medallion(path string) *image.RGBA {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	box := image.Rect(medalX-medalR, medalY-medalR, medalX+medalR, medalY+medalR)
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(img, img.Bounds(), src, box, draw.Src, nil)
	const r = size/2 - 1
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			in := 0 // 4x4 supersampled coverage of the circle
			for sy := 0; sy < 4; sy++ {
				for sx := 0; sx < 4; sx++ {
					dx := float64(x) + (float64(sx)+0.5)/4 - size/2
					dy := float64(y) + (float64(sy)+0.5)/4 - size/2
					if dx*dx+dy*dy <= r*r {
						in++
					}
				}
			}
			c := img.RGBAAt(x, y)
			a := uint8(255 * in / 16)
			img.SetRGBA(x, y, color.RGBA{premul(c.R, a), premul(c.G, a), premul(c.B, a), a})
		}
	}
	return img
}

// drawnS is the fallback icon: a gold "S" on a navy disc with a gold ring.
func drawnS() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Anti-aliased disc with a gold ring: 4x4 supersampling per pixel.
	const r, ring = size/2 - 2, 12.0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var in, edge int
			for sy := 0; sy < 4; sy++ {
				for sx := 0; sx < 4; sx++ {
					dx := float64(x) + (float64(sx)+0.5)/4 - size/2
					dy := float64(y) + (float64(sy)+0.5)/4 - size/2
					d2 := dx*dx + dy*dy
					if d2 <= r*r {
						in++
						if d2 > (r-ring)*(r-ring) {
							edge++
						}
					}
				}
			}
			if in == 0 {
				continue
			}
			c := mix(navy, gold, float64(edge)/float64(in))
			c.A = uint8(255 * in / 16)
			c.R, c.G, c.B = premul(c.R, c.A), premul(c.G, c.A), premul(c.B, c.A)
			img.SetRGBA(x, y, c)
		}
	}

	f, err := opentype.Parse(gobold.TTF)
	if err != nil {
		log.Fatal(err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 170, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		log.Fatal(err)
	}
	d := &font.Drawer{Dst: img, Src: image.NewUniform(gold), Face: face}
	b, _ := d.BoundString("S")
	w, h := b.Max.X-b.Min.X, b.Max.Y-b.Min.Y
	d.Dot = fixed.Point26_6{
		X: fixed.I(size/2) - w/2 - b.Min.X,
		Y: fixed.I(size/2) - h/2 - b.Min.Y,
	}
	d.DrawString("S")
	return img
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	l := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t) }
	return color.RGBA{l(a.R, b.R), l(a.G, b.G), l(a.B, b.B), 0xff}
}

func premul(v, a uint8) uint8 { return uint8(uint32(v) * uint32(a) / 255) }
