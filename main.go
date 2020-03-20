package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/gonutz/prototype/draw"
)

func main() {
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

	if f, err := os.Open("level"); err == nil {
		defer f.Close()
		tiles = nil
		enc := binary.LittleEndian
		for {
			var x, y uint16
			var color uint8
			if binary.Read(f, enc, &x) != nil {
				break
			}
			binary.Read(f, enc, &y)
			binary.Read(f, enc, &color)
			tiles = append(tiles, tile{
				x:     int(x),
				y:     int(y),
				color: int(color),
			})
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
			c := circle{x: newX, y: newY, r: ballRadius}
			hitAny := false
			var deflectX, deflectY float64
			for i, t := range tiles {
				r := rect{
					x: float64(t.x),
					y: float64(t.y),
					w: tileW,
					h: tileH,
				}
				hit, dx, dy := collideCircleWithRect(c, r)
				if hit {
					hitAny = true
					deflectX += dx
					deflectY += dy

					tiles[i].hit++
					if tiles[i].hit >= 3 {
						tiles[i].x = -999
					}
				}
			}
			if hitAny {
				normalize(&deflectX, &deflectY)
				ball.vx, ball.vy = bounceDir(ball.vx, ball.vy, deflectX, deflectY)
				ball.vx, ball.vy = makeNonHorizontal(ball.vx, ball.vy)
			}

			var horHit, verHit bool
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
			window.DrawImageFile(allTiles[tile.color], tile.x, tile.y)
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

type rect struct {
	x, y, w, h float64
}

type circle struct {
	x, y, r float64
}

func collideCircleWithRect(c circle, r rect) (hit bool, deflectX, deflectY float64) {
	if r.y <= c.y && c.y <= r.y+r.h {
		hit = r.x-c.r <= c.x && c.x <= r.x+r.w+c.r
		if hit {
			if c.x < r.x+r.w/2 {
				deflectX = -1
			} else {
				deflectX = 1
			}
		}
	} else if r.x <= c.x && c.x <= r.x+r.w {
		hit = r.y-c.r <= c.y && c.y <= r.y+r.h+c.r
		if hit {
			if c.y < r.y+r.h/2 {
				deflectY = -1
			} else {
				deflectY = 1
			}
		}
	} else {
		for _, corner := range [][2]float64{
			{r.x, r.y},
			{r.x + r.w, r.y},
			{r.x + r.w, r.y + r.h},
			{r.x, r.y + r.h},
		} {
			dx := corner[0] - c.x
			dy := corner[1] - c.y
			if dx*dx+dy*dy < c.r*c.r {
				hit = true
				deflectX = c.x - corner[0]
				deflectY = c.y - corner[1]
				normalize(&deflectX, &deflectY)
			}
		}
	}
	return
}

func bounceDir(dirX, dirY, surfaceNormalX, surfaceNormalY float64) (bx, by float64) {
	normalize(&surfaceNormalX, &surfaceNormalY)
	f := 2.0 * (dirX*surfaceNormalX + dirY*surfaceNormalY)
	bx = dirX - f*surfaceNormalX
	by = dirY - f*surfaceNormalY
	normalize(&bx, &by)
	return
}

func normalize(x, y *float64) {
	if *x != 0 || *y != 0 {
		f := 1.0 / math.Hypot(*x, *y)
		*x *= f
		*y *= f
	}
}

func makeNonHorizontal(dx, dy float64) (float64, float64) {
	const min = math.Pi / 10
	angle := math.Atan2(dy, dx)
	for angle > 2*math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < 0 {
		angle += 2 * math.Pi
	}
	if angle < min {
		angle = min
	}
	if math.Pi-min < angle && angle < math.Pi+min {
		if angle < math.Pi {
			angle = math.Pi - min
		} else {
			angle = math.Pi + min
		}
	}
	if angle > 2*math.Pi-min {
		angle = 2*math.Pi - min
	}
	dy, dx = math.Sincos(angle)
	return dx, dy
}
