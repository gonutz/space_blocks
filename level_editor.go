//+build ignore

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gonutz/prototype/draw"
)

func main() {
	var tiles []tile
	defer func() {
		enc := binary.LittleEndian
		var buf bytes.Buffer
		for _, t := range tiles {
			binary.Write(&buf, enc, uint16(t.x))
			binary.Write(&buf, enc, uint16(t.y))
			binary.Write(&buf, enc, uint8(t.color))
		}
		ioutil.WriteFile("level", buf.Bytes(), 0666)
	}()

	addTile := func(t tile) {
		for _, t2 := range tiles {
			if overlap(t, t2) {
				return
			}
		}
		tiles = append(tiles, t)
	}

	alignX := 16
	alignY := 16
	settingsPath := filepath.Join(os.Getenv("APPDATA"), "space_level_editor.set")
	defer func() {
		ioutil.WriteFile(settingsPath, []byte{byte(alignX), byte(alignY)}, 0666)
	}()
	settings, err := ioutil.ReadFile(settingsPath)
	if err == nil && len(settings) == 2 {
		alignX = int(settings[0])
		alignY = int(settings[1])
	}

	currentTile := 0
	draw.RunWindow("Space Blocks - Level Editor", windowW, windowH, func(window draw.Window) {
		if window.WasKeyPressed(draw.KeyEscape) {
			window.Close()
		}

		for i := range allTiles {
			if window.WasKeyPressed(draw.Key1 + draw.Key(i)) {
				currentTile = i
			}
		}

		if window.WasKeyPressed(draw.KeyZ) && len(tiles) > 0 {
			tiles = tiles[:len(tiles)-1]
		}

		if window.WasKeyPressed(draw.KeyLeft) && alignX > 1 {
			alignX--
		}
		if window.WasKeyPressed(draw.KeyRight) && alignX < 255 {
			alignX++
		}
		if window.WasKeyPressed(draw.KeyDown) && alignY > 1 {
			alignY--
		}
		if window.WasKeyPressed(draw.KeyUp) && alignY < 255 {
			alignY++
		}

		mx, my := window.MousePosition()
		mx += tileW / 2
		my += tileH / 2
		tileX, tileY := mx-tileW/2, my-tileH/2
		tileX = (tileX / alignX) * alignX
		tileY = (tileY / alignY) * alignY

		wasLeftClicked := func() bool {
			for _, c := range window.Clicks() {
				if c.Button == draw.LeftButton {
					return true
				}
			}
			return false
		}

		if wasLeftClicked() || window.IsMouseDown(draw.LeftButton) {
			addTile(tile{
				x:     tileX,
				y:     tileY,
				color: currentTile,
			})
		}

		for _, t := range tiles {
			window.DrawImageFile(allTiles[t.color], t.x, t.y)
		}
		window.DrawImageFile(allTiles[currentTile], tileX, tileY)

		text := fmt.Sprint("Align Y: ", alignY, " pixels (UP/DOWN)")
		window.DrawText(text, 0, windowH-15, draw.White)
		text = fmt.Sprint("Align X: ", alignX, " pixels (LEFT/RIGHT)")
		window.DrawText(text, 0, windowH-30, draw.White)
	})
}

type tile struct {
	x, y, color int
}

func overlap(a, b tile) bool {
	if a.x+tileW <= b.x {
		return false
	}
	if b.x+tileW <= a.x {
		return false
	}
	if a.y+tileH <= b.y {
		return false
	}
	if b.y+tileH <= a.y {
		return false
	}
	return true
}
