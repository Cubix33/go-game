package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth     = 800
	screenHeight    = 600
	playerSpeed     = 4.0
	playerWidth     = 64
	playerHeight    = 64
	playerImagePath = "sprites/ship.png"
	bulletSpeed     = 8.0
	bulletWidth     = 8
	bulletHeight    = 8
	enemySpeed      = 2.0
)

var (
	playerImage *ebiten.Image
	bulletImage *ebiten.Image
	playerX     = float64(screenWidth / 2)
	playerY     = float64(screenHeight - playerHeight - 20)
	bullets     []*bullet
	enemies     []*enemy
)

type bullet struct {
	x, y  float64
	alive bool
}

type enemy struct {
	x, y  float64
	alive bool
}
type game struct{}

func (g *game) Update() error {

	handlePlayerMovement()

	handleShooting()

	updateBullets()

	updateEnemies()

	handleCollisions()

	return nil
}

func (g *game) Draw(screen *ebiten.Image) {

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(playerX, playerY)
	screen.DrawImage(playerImage, op)

	drawBullets(screen)

	drawEnemies(screen)

	ebitenutil.DebugPrint(screen, "Press arrow keys to move, space to shoot")
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	var err error
	playerImage, _, err = ebitenutil.NewImageFromFile(playerImagePath)
	if err != nil {
		log.Fatal(err)
	}

	bulletImage = ebiten.NewImage(bulletWidth, bulletHeight)
	// ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	bulletImage.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Side-Scrolling Shooter Game")

	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}

func handlePlayerMovement() {

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		playerX -= playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		playerX += playerSpeed
	}

	if playerX < 0 {
		playerX = 0
	}
	if playerX > screenWidth-playerWidth {
		playerX = screenWidth - playerWidth
	}
}

func handleShooting() {

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		bullets = append(bullets, &bullet{
			x:     playerX + playerWidth/2 - bulletWidth/2,
			y:     playerY,
			alive: true,
		})
	}
}

func updateBullets() {

	for _, b := range bullets {
		if b.alive {
			b.y -= bulletSpeed
		}
	}
}

func updateEnemies() {

	for _, e := range enemies {
		if e.alive {
			e.y += enemySpeed
		}
	}
}

func handleCollisions() {

	for _, b := range bullets {
		if !b.alive {
			continue
		}
		for _, e := range enemies {
			if !e.alive {
				continue
			}
			if collision(b.x, b.y, bulletWidth, bulletHeight, e.x, e.y, playerWidth, playerHeight) {
				b.alive = false
				e.alive = false

			}
		}
	}
}

func collision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

func drawBullets(screen *ebiten.Image) {

	for _, b := range bullets {
		if b.alive {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(b.x, b.y)
			screen.DrawImage(bulletImage, op)
		}
	}
}

func drawEnemies(screen *ebiten.Image) {

	for _, e := range enemies {
		if e.alive {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(e.x, e.y)
			screen.DrawImage(playerImage, op)
		}
	}
}
