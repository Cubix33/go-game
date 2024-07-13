package main

import (
	"log"
	"math/rand"
	"time"
	"os"
	"fmt"

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
	playerImagePath = "sprites/ship1.png"
	bulletImagePath = "sprites/bill1.png"
	enemyImagePath  = "sprites/zombii.png"
	backgroundImagePath = "sprites/bg.png"
	bulletSpeed     = 8.0
	bulletWidth     = 8
	bulletHeight    = 7
	enemySpeed      = 2.0
	enemyWidth      = 64
	enemyHeight     = 64
	maxEnemies      = 7
)

var (
	playerImage *ebiten.Image
	bulletImage *ebiten.Image
	enemyImage  *ebiten.Image
	backgroundImage *ebiten.Image
	playerX     = float64(screenWidth / 2)
	playerY     = float64(screenHeight - playerHeight - 20)
	bullets     []*bullet
	enemies     []*enemy
	gameOver    bool
	score       int 
)

type bullet struct {
	x, y  float64
	frame int
	alive bool
}

type enemy struct {
	x, y  float64
	alive bool
}

type game struct{}

func (g *game) Update() error {
	if gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			resetGame()
		} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			os.Exit(0)
		}
		return nil
	}
	handlePlayerMovement()
	handleShooting()
	updateBullets()
	updateEnemies()
	handleCollisions()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(backgroundImage, op)
	op.GeoM.Translate(playerX, playerY)
	screen.DrawImage(playerImage, op)

	drawBullets(screen)
	drawEnemies(screen)
	scoreText := fmt.Sprintf("Score: %d", score)
    ebitenutil.DebugPrintAt(screen, scoreText, 10, 10)

    if gameOver {
        gameOverText := fmt.Sprintf("Game Over! Total Score: %d\nPress 'R' to restart or 'Esc' to exit", score)
        ebitenutil.DebugPrintAt(screen, gameOverText, screenWidth/2-100, screenHeight/2)
    } else {
        ebitenutil.DebugPrint(screen, "Press arrow keys to move, space to shoot")
    }
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

	bulletImage, _, err = ebitenutil.NewImageFromFile(bulletImagePath)
	if err != nil {
		log.Fatal(err)
	}

	enemyImage, _, err = ebitenutil.NewImageFromFile(enemyImagePath)
	if err != nil {
		log.Fatal(err)
	}

	backgroundImage, _, err = ebitenutil.NewImageFromFile(backgroundImagePath)
	if err != nil {
		log.Fatal(err)
	}

	initializeEnemies()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Side-Scrolling Shooter Game")

	go spawnEnemies()

	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}

func initializeEnemies() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < maxEnemies; i++ {
		x := float64(rand.Intn(screenWidth - enemyWidth))
		y := float64(rand.Intn(screenHeight/2 - enemyHeight))
		enemies = append(enemies, &enemy{
			x:     x,
			y:     y,
			alive: true,
		})
	}
}

func spawnEnemies() {
	rand.Seed(time.Now().UnixNano())
	for {
		time.Sleep(time.Second)
		if countAliveEnemies() < maxEnemies { 
			x := float64(rand.Intn(screenWidth - enemyWidth))
			y := -float64(enemyHeight) 
			enemies = append(enemies, &enemy{
				x:     x,
				y:     y,
				alive: true,
			})
		}
	}
}

func countAliveEnemies() int { 
	count := 0
	for _, e := range enemies {
		if e.alive {
			count++
		}
	}
	return count
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
		bulletX := playerX + (playerWidth / 2) - (bulletWidth + 0.5 / 2)
		bulletY := playerY - bulletHeight
		
		bullets = append(bullets, &bullet{
			x:     bulletX,
			y:     bulletY,
			frame: 0,
			alive: true,
		})
	}
}

func updateBullets() {
	for i := len(bullets) - 1; i >= 0; i-- {
		b := bullets[i]
		if b.alive {
			b.y -= bulletSpeed
			if b.y < -bulletHeight {
				bullets = append(bullets[:i], bullets[i+1:]...)
			}
		}
	}
}

func updateEnemies() {
	for i := len(enemies) - 1; i >= 0; i-- {
		e := enemies[i]
		if e.alive {
			e.y += enemySpeed
			if e.y + enemyHeight >= screenHeight {
				gameOver = true
			}
		}
	}
}

func handleCollisions() {
	for i := len(bullets) - 1; i >= 0; i-- {
		b := bullets[i]
		if !b.alive {
			continue
		}
		for j := len(enemies) - 1; j >= 0; j-- {
			e := enemies[j]
			if !e.alive {
				continue
			}
			if collision(b.x, b.y, bulletWidth, bulletHeight, e.x, e.y, enemyWidth, enemyHeight) {
				bullets = append(bullets[:i], bullets[i+1:]...)
				enemies = append(enemies[:j], enemies[j+1:]...)
				score +=1
				break
			}
		}
	}
}

//func checkGameOver() {
//	for _, e := range enemies {
//		if e.alive && e.y > screenHeight {
//			gameOver = true
//			break
//		}
//	}
//}

func resetGame() {
	playerX = float64(screenWidth / 2)
	playerY = float64(screenHeight - playerHeight - 20)
	bullets = []*bullet{}
	enemies = []*enemy{}
	score   = 0
	gameOver = false
	initializeEnemies()
	go spawnEnemies()
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
			screen.DrawImage(enemyImage, op)
		}
	}
}

