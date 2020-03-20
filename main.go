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
		ballRadius                       = 10
		tileW, tileH                     = 64, 32
		panelW, panelH                   = 100, 24
		panelY                           = windowH - panelH
		panelSpeed                       = 8
		ballSpeed                        = 8
	)
	const (
		stateIdle = iota
		statePlaying
	)
	lives := 5
	var ball ball
	ball.x, ball.y = 230, 412
	ball.vx, ball.vy = math.Cos(0.5), -math.Sin(0.5)
	panelX := 0
	state := stateIdle

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
		ballSpeed := ballSpeed
		if dead {
			ball.x = float64(panelX)

			wasLeftClicked := false
			for _, c := range window.Clicks() {
				if c.Button == draw.LeftButton {
					wasLeftClicked = true
				}
			}

			if wasLeftClicked {
				if ball.vx == 0 && ball.vy == 0 {
					ball.vy = -1
				}
				dead = false
			} else {
				ballSpeed = 0
			}
		}

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

		// Draw background.
		for x := 0; x < windowW; x += backgroundTileW {
			for y := 0; y < windowH; y += backgroundTileH {
				window.DrawImageFile("background.png", x, y)
			}
		}

		// Draw tiles.
		for _, tile := range tiles {
			c := tileColors[tile.color]
			window.FillRect(tile.x, tile.y, tileW, tileH, c[fillColor])
			window.FillRect(tile.x, tile.y+2, tileW, 6, c[topColor])
			window.FillRect(tile.x+2, tile.y, 6, tileH, c[sideColor])
			window.FillRect(tile.x+tileW-8, tile.y, 6, tileH, c[sideColor])
			window.FillRect(tile.x, tile.y+tileH-8, tileW, 6, c[bottomColor])
			window.DrawRect(tile.x, tile.y, tileW, tileH, c[borderColor])
			window.DrawRect(tile.x+1, tile.y+1, tileW-2, tileH-2, c[borderColor])
			if tile.hit == 1 {
				window.DrawImageFile("damage1.png", tile.x, tile.y)
			}
			if tile.hit == 2 {
				window.DrawImageFile("damage2.png", tile.x, tile.y)
			}
		}

		// Draw mouse panel.
		panelLeft := panelX - panelW/2
		window.FillEllipse(
			panelLeft-panelH/2-4,
			panelY,
			panelH,
			panelH,
			draw.Red,
		)
		window.FillEllipse(
			panelLeft-panelH/2+panelW+4,
			panelY,
			panelH,
			panelH,
			draw.Red,
		)
		window.FillRect(
			panelLeft,
			panelY,
			panelW,
			panelH/2,
			draw.RGB(0.85, 0.85, 0.85),
		)
		window.FillRect(
			panelLeft,
			panelY+panelH/2,
			panelW,
			panelH/2,
			draw.RGB(0.8, 0.8, 0.8),
		)

		// Draw ball.
		window.FillEllipse(
			round(ball.x)-ballRadius,
			round(ball.y)-ballRadius,
			2*ballRadius,
			2*ballRadius,
			draw.White,
		)
		window.DrawEllipse(
			round(ball.x)-ballRadius,
			round(ball.y)-ballRadius,
			2*ballRadius,
			2*ballRadius,
			draw.LightGray,
		)

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

const (
	fillColor = iota
	topColor
	bottomColor
	sideColor
	borderColor
)

var tileColors = [][5]draw.Color{
	[5]draw.Color{
		draw.RGB(242/255.0, 55/255.0, 55/255.0),
		draw.RGB(244/255.0, 113/255.0, 113/255.0),
		draw.RGB(190/255.0, 14/255.0, 14/255.0),
		draw.RGB(222/255.0, 16/255.0, 16/255.0),
		draw.RGB(142/255.0, 11/255.0, 11/255.0),
	},
	[5]draw.Color{
		draw.RGB(131/255.0, 89/255.0, 149/255.0),
		draw.RGB(155/255.0, 117/255.0, 172/255.0),
		draw.RGB(98/255.0, 67/255.0, 112/255.0),
		draw.RGB(121/255.0, 83/255.0, 138/255.0),
		draw.RGB(88/255.0, 61/255.0, 101/255.0),
	},
	[5]draw.Color{
		draw.RGB(74/255.0, 191/255.0, 240/255.0),
		draw.RGB(122/255.0, 207/255.0, 243/255.0),
		draw.RGB(20/255.0, 165/255.0, 226/255.0),
		draw.RGB(52/255.0, 182/255.0, 237/255.0),
		draw.RGB(16/255.0, 125/255.0, 171/255.0),
	},
	[5]draw.Color{
		draw.RGB(159/255.0, 206/255.0, 49/255.0),
		draw.RGB(179/255.0, 216/255.0, 90/255.0),
		draw.RGB(133/255.0, 172/255.0, 40/255.0),
		draw.RGB(149/255.0, 192/255.0, 46/255.0),
		draw.RGB(106/255.0, 137/255.0, 33/255.0),
	},
	[5]draw.Color{
		draw.RGB(255/255.0, 204/255.0, 0/255.0),
		draw.RGB(254/255.0, 219/255.0, 78/255.0),
		draw.RGB(202/255.0, 162/255.0, 2/255.0),
		draw.RGB(227/255.0, 182/255.0, 2/255.0),
		draw.RGB(167/255.0, 133/255.0, 3/255.0),
	},
	[5]draw.Color{
		draw.RGB(204/255.0, 204/255.0, 204/255.0),
		draw.RGB(221/255.0, 221/255.0, 221/255.0),
		draw.RGB(166/255.0, 166/255.0, 166/255.0),
		draw.RGB(187/255.0, 187/255.0, 187/255.0),
		draw.RGB(119/255.0, 119/255.0, 119/255.0),
	},
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
