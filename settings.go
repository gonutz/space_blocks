package main

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

var allTiles = []string{
	"red.png",
	"purple.png",
	"blue.png",
	"green.png",
	"yellow.png",
	"gray.png",
}
