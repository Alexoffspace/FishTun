//go:build ignore

// Command genico builds icon.ico from icon.png.
//
// It is a one-off maintenance helper: run `go run genico.go` after changing
// icon.png, then commit the regenerated icon.ico.
//
// The output follows the ICO container format: entries of 16/32/48 px are
// stored as uncompressed 32-bit DIBs (maximum compatibility, required by small
// tray sizes on older Windows shells); the 256 px entry is stored as PNG, which
// every Vista-or-newer loader understands.
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

var sizes = []int{16, 32, 48, 256}

func main() {
	src, err := os.Open("icon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	img, err := png.Decode(src)
	if err != nil {
		log.Fatal(err)
	}

	type entry struct {
		size      int
		payload   []byte
		pngFormat bool
	}

	encoded := make([]entry, 0, len(sizes))
	total := 0
	for _, size := range sizes {
		payload, pngFormat := encodeEntry(resizeRGBA(img, size), size)
		encoded = append(encoded, entry{size: size, payload: payload, pngFormat: pngFormat})
		total += len(payload)
	}

	offset := 6 + 16*len(encoded)
	dir := new(bytes.Buffer)
	body := new(bytes.Buffer)
	for _, e := range encoded {
		writeDirEntry(dir, e.size, uint32(len(e.payload)), e.pngFormat, uint32(offset))
		body.Write(e.payload)
		offset += len(e.payload)
	}

	out := new(bytes.Buffer)
	binary.Write(out, binary.LittleEndian, uint16(0))          // reserved
	binary.Write(out, binary.LittleEndian, uint16(1))          // type: icon
	binary.Write(out, binary.LittleEndian, uint16(len(sizes))) // image count
	out.Write(dir.Bytes())
	out.Write(body.Bytes())

	if err := os.WriteFile("icon.ico", out.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote icon.ico (%d bytes, %d entries)", out.Len(), len(sizes))

	// Keep the Wails Windows resource icon in sync so the exe, window and tray
	// all show the same artwork.
	if err := os.WriteFile(filepath.Join("build", "windows", "icon.ico"), out.Bytes(), 0o644); err != nil {
		log.Fatalf("syncing build/windows/icon.ico: %v", err)
	}
}

// writeDirEntry encodes one ICONDIRENTRY.
func writeDirEntry(buf *bytes.Buffer, size int, payloadLen uint32, pngFormat bool, offset uint32) {
	dim := byte(size)
	if size >= 256 {
		dim = 0
	}
	buf.WriteByte(dim) // width
	buf.WriteByte(dim) // height
	buf.WriteByte(0)   // palette colours
	buf.WriteByte(0)   // reserved
	if pngFormat {
		binary.Write(buf, binary.LittleEndian, uint16(1))
	} else {
		binary.Write(buf, binary.LittleEndian, uint16(0))
	}
	binary.Write(buf, binary.LittleEndian, uint16(32)) // bits per pixel
	binary.Write(buf, binary.LittleEndian, payloadLen)
	binary.Write(buf, binary.LittleEndian, offset)
}

// encodeEntry renders img at size and returns its ICO payload plus whether it is
// PNG-encoded (true) or DIB-encoded (false).
func encodeEntry(img *image.RGBA, size int) ([]byte, bool) {
	if size >= 256 {
		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			log.Fatal(err)
		}
		return buf.Bytes(), true
	}
	return dibPayload(img, size), false
}

// dibPayload encodes an uncompressed bottom-up 32-bit BGRA DIB with an AND mask.
func dibPayload(img *image.RGBA, size int) []byte {
	buf := new(bytes.Buffer)

	header := make([]byte, 40)
	binary.LittleEndian.PutUint32(header[0:], 40)             // biSize
	binary.LittleEndian.PutUint32(header[4:], uint32(size))   // biWidth
	binary.LittleEndian.PutUint32(header[8:], uint32(size*2)) // biHeight: XOR + AND
	binary.LittleEndian.PutUint16(header[12:], 1)             // biPlanes
	binary.LittleEndian.PutUint16(header[14:], 32)            // biBitCount
	buf.Write(header)

	// XOR: pixels bottom-up in BGRA order.
	for y := size - 1; y >= 0; y-- {
		for x := 0; x < size; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			buf.Write([]byte{byte(b >> 8), byte(g >> 8), byte(r >> 8), byte(a >> 8)})
		}
	}

	// AND mask: fully transparent pixels are 1, rows padded to 4 bytes.
	stride := (size + 31) / 32 * 4
	mask := make([]byte, stride*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if _, _, _, a := img.At(x, size-1-y).RGBA(); a == 0 {
				mask[y*stride+x/8] |= 0x80 >> uint(x%8)
			}
		}
	}
	buf.Write(mask)

	return buf.Bytes()
}

// resizeRGBA scales src to size×size with nearest-neighbour sampling, which is
// adequate for a flat, high-contrast application logo.
func resizeRGBA(src image.Image, size int) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			sx := b.Min.X + x*b.Dx()/size
			sy := b.Min.Y + y*b.Dy()/size
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
