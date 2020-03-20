package main

import (
	"fmt"
	"github.com/gonutz/prototype/draw"
	"math"
)

func main() {
	const (
		windowW, windowH                 = 720, 600
		backgroundTileW, backgroundTileH = 64, 48
		score                            = 1300
		ballRadius                       = 11
		tileW, tileH                     = 64, 32
		panelW, panelH                   = 144, 24
		panelY                           = windowH - panelH
		panelSpeed                       = 8
		ballSpeed                        = 8
	)
	lives := 5
	var ball ball
	ball.x, ball.y = 230, 412
	ball.vx, ball.vy = math.Cos(0.5), -math.Sin(0.5)
	panelX := 0

	var tiles []tile
	for x := 0; x < 9; x++ {
		for y := 0; y < 6; y++ {
			tiles = append(tiles, tile{x: 72 + x*tileW, y: 74 + y*tileH, color: y})
		}
	}

	lastMouseX := -1
	err := draw.RunWindow("Space Blocks", windowW, windowH, func(window draw.Window) {
		// Handle input.
		if window.WasKeyPressed(draw.KeyEscape) {
			window.Close()
		}
		var panelDx int
		if window.IsKeyDown(draw.KeyLeft) {
			panelDx--
		}
		if window.IsKeyDown(draw.KeyRight) {
			panelDx++
		}
		panelX += panelSpeed * panelDx

		mx, _ := window.MousePosition()
		if mx != lastMouseX {
			panelX = mx
			lastMouseX = mx
		}

		// Update world.
		for step := 0; step < ballSpeed; step++ {
			newX := ball.x + ball.vx
			newY := ball.y + ball.vy
			var horHit, verHit bool
			// Let ball hit tiles.
			for i, t := range tiles {
				var h, v bool
				if horIntersect(newX, newY, ballRadius, float64(t.x), float64(t.x+tileW), float64(t.y)) ||
					horIntersect(newX, newY, ballRadius, float64(t.x), float64(t.x+tileW), float64(t.y+tileH)) {
					h = true
				}
				if verIntersect(newX, newY, ballRadius, float64(t.x), float64(t.y), float64(t.y+tileH)) ||
					verIntersect(newX, newY, ballRadius, float64(t.x+tileW), float64(t.y), float64(t.y+tileH)) {
					v = true
				}
				if h || v {
					tiles[i].hit++
					if tiles[i].hit >= 3 {
						tiles[i].x = -999
					}
				}
				horHit = horHit || h
				verHit = verHit || v
			}
			// Let ball hit walls.
			if horIntersect(newX, newY, ballRadius, 0, windowW, 0) {
				horHit = true
			}
			if verIntersect(newX, newY, ballRadius, 0, 0, windowH) ||
				verIntersect(newX, newY, ballRadius, windowW, 0, windowH) {
				verHit = true
			}
			// Let ball hit panel.
			if horIntersect(
				newX, newY, ballRadius,
				float64(panelX-panelW/2), float64(panelX+panelW/2), panelY) {
				// Intercept other collisions, the position where the ball hit
				// the panel determines the new direction of the ball.
				x := (newX - float64(panelX)) / (0.5 * panelW) // [-1..1]
				if x < -1 {
					x = -1
				}
				if x > 1 {
					x = 1
				}
				ball.vy, ball.vx = math.Sincos(3*math.Pi/2 + x*math.Pi/3)
				horHit = false
				verHit = false
			}
			// Turn ball around if something was hit.
			if horHit {
				ball.vy *= -1
			}
			if verHit {
				ball.vx *= -1
			}
			ball.x += ball.vx
			ball.y += ball.vy
		}

		if ball.y > windowH {
			lives--
			ball.x = float64(panelX)
			ball.y = panelY - ballRadius - 0.1
			ball.vx, ball.vy = 0, 0
		}

		for _, c := range window.Clicks() {
			if c.Button == draw.LeftButton {
				if ball.vx == 0 && ball.vy == 0 {
					ball.vy = -1
				}
			}
		}

		// Draw background.
		for x := 0; x < windowW; x += backgroundTileW {
			for y := 0; y < windowH; y += backgroundTileH {
				window.DrawImageFile("background.png", x, y)
			}
		}

		// Draw tiles.
		for _, tile := range tiles {
			file := []string{
				"red", "purple", "blue", "green", "yellow", "gray",
			}[tile.color]
			window.DrawImageFile(file+".png", tile.x, tile.y)
			if tile.hit == 1 {
				window.DrawImageFile("damage1.png", tile.x, tile.y)
			}
			if tile.hit == 2 {
				window.DrawImageFile("damage2.png", tile.x, tile.y)
			}
		}

		// Draw panel and ball.
		window.DrawImageFile("panel.png", panelX-panelW/2, panelY)
		window.DrawImageFile("ball.png", round(ball.x)-ballRadius, round(ball.y)-ballRadius)

		// Draw score and lives texts.
		const textScale = 2
		scoreText := fmt.Sprintf("SCORE: %d", score)
		window.DrawScaledText(scoreText, 10, 10, textScale, draw.White)
		livesText := fmt.Sprintf("LIVES: %d", lives)
		w, _ := window.GetScaledTextSize(livesText, textScale)
		window.DrawScaledText(livesText, windowW-10-w, 10, textScale, draw.White)
	})
	if err != nil {
		panic(err)
	}
}

type ball struct {
	x, y   float64
	vx, vy float64
}

type tile struct {
	x, y  int
	hit   int
	color int
}

func round(x float64) int {
	if x < 0 {
		return int(x - 0.5)
	}
	return int(x + 0.5)
}

// x,y,r describe a circle (center and radius), lineX1,lineX2,lineY descirbe a
// horizontal line segment.
func horIntersect(x, y, r, lineX1, lineX2, lineY float64) bool {
	if abs(y-lineY) <= r {
		if lineX2 < lineX1 {
			lineX1, lineX2 = lineX2, lineX1
		}
		if lineX1 <= x && x <= lineX2 {
			return true
		}
		if square(x-lineX1)+square(y-lineY) <= r*r {
			return true
		}
		if square(x-lineX2)+square(y-lineY) <= r*r {
			return true
		}
	}
	return false
}

// x,y,r describe a circle (center and radius), lineX,lineY1,lineY2 descirbe a
// vertical line segment.
func verIntersect(x, y, r, lineX, lineY1, lineY2 float64) bool {
	if abs(x-lineX) <= r {
		if lineY2 < lineY1 {
			lineY1, lineY2 = lineY2, lineY1
		}
		if lineY1 <= y && y <= lineY2 {
			return true
		}
		if square(y-lineY1)+square(x-lineX) <= r*r {
			return true
		}
		if square(y-lineY2)+square(x-lineX) <= r*r {
			return true
		}
	}
	return false
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func square(x float64) float64 {
	return x * x
}
