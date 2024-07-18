package main

import (
	"image/color"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth         = 800
	screenHeight        = 600
	playerSpeed         = 4.0
	playerWidth         = 64
	playerHeight        = 64
	bulletSpeed         = 8.0
	bulletWidth         = 8
	bulletHeight        = 8
	startButtonWidth    = 200
	startButtonHeight   = 50
	playerImagePath     = "sprites/ship1.png"
	bulletImagePath     = "sprites/bill1.png"
	obstacleImagePath   = "sprites/obstacle.png"
	obstacleWidth       = 64
	obstacleHeight      = 64
	enemyImagePath      = "sprites/zombii.png"
	backgroundImagePath = "sprites/bg.png"
	bulletSoundPath     = "sounds/bullet.wav"
	gameOverSoundPath   = "sounds/game_over.wav"
)

var (
	playerImage     *ebiten.Image
	bulletImage     *ebiten.Image
	enemyImage      *ebiten.Image
	backgroundImage *ebiten.Image
	obstacleImage   *ebiten.Image
	playerX         = float64(screenWidth / 2)
	playerY         = float64(screenHeight - playerHeight - 20)
	bullets         []*bullet
	obstacles       []*obstacle
	enemies         []*enemy
	score           int
	gameOver        bool
	gameStarted     bool
	startButtonX    = float64((screenWidth - startButtonWidth) / 2)
	startButtonY    = float64((screenHeight - startButtonHeight) / 2)
	audioContext    *audio.Context
	bulletSound     *audio.Player
	gameOverSound   *audio.Player
)

type bullet struct {
	x, y  float64
	alive bool
}

type obstacle struct {
	x, y   float64
	deadly bool
	rotation float64
}

type enemy struct {
    x, y    float64
    alive   bool
    deadly  bool
}



type game struct{}

func (g *game) Update() error {
	if !gameStarted {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mouseX, mouseY := ebiten.CursorPosition()
			if float64(mouseX) >= startButtonX && float64(mouseX) <= startButtonX+startButtonWidth &&
				float64(mouseY) >= startButtonY && float64(mouseY) <= startButtonY+startButtonHeight {
				gameStarted = true
				resetGame()
			}
		}
		return nil
	}

	if gameOver {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mouseX, mouseY := ebiten.CursorPosition()
			if float64(mouseX) >= startButtonX && float64(mouseX) <= startButtonX+startButtonWidth &&
				float64(mouseY) >= startButtonY && float64(mouseY) <= startButtonY+startButtonHeight {
				resetGame()
			}
		}
		return nil
	}

	handlePlayerMovement()
	handleShooting()
	updateBullets()
	updateObstacles()
	updateEnemies()
	handleCollisions()

	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	if !gameStarted {
		drawStartButton(screen)
		return
	}

	if gameOver {
		drawGameOverScreen(screen)
		return
	}

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(backgroundImage, op)

	op.GeoM.Reset()
	op.GeoM.Translate(playerX, playerY)
	screen.DrawImage(playerImage, op)

	drawBullets(screen)
	drawObstacles(screen)
	drawEnemies(screen)

	ebitenutil.DebugPrint(screen, "Score: "+strconv.Itoa(score))
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

	obstacleImage, _, err = ebitenutil.NewImageFromFile(obstacleImagePath)
        if err != nil {
                log.Fatal(err)
        }

	
	audioContext = audio.NewContext(44100)

	bulletSound, err = loadSound(audioContext, bulletSoundPath)
	if err != nil {
		log.Fatal(err)
	}

	gameOverSound, err = loadSound(audioContext, gameOverSoundPath)
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Side-Scrolling Shooter Game")


	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}

func loadSound(context *audio.Context, path string) (*audio.Player, error) {
	f, err := ebitenutil.OpenFile(path)
	if err != nil {
		return nil, err
	}
	d, err := wav.Decode(context, f)
	if err != nil {
		return nil, err
	}

	p, err := context.NewPlayer(d)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func handlePlayerMovement() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		playerX -= playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		playerX += playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		playerY -= playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		playerY += playerSpeed
	}
	if playerX < 0 {
		playerX = 0
	}
	if playerX > screenWidth-playerWidth {
		playerX = screenWidth - playerWidth
	}
	if playerY < 0 {
		playerY = 0
	}
	if playerY > screenHeight-playerHeight {
		playerY = screenHeight - playerHeight
	}
}

func handleShooting() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		log.Println("Shooting bullet")
		bulletSound.Rewind()
		bulletSound.Play()

		bullets = append(bullets, &bullet{
			x:     playerX + (playerWidth)/2 - (bulletWidth+ 5)/2,
			y:     playerY - bulletHeight,
			alive: true,
		})
		log.Println("Bullet added at position:", playerX+playerWidth/2-bulletWidth/2, playerY)
	}
}

func updateBullets() {
	for _, b := range bullets {
		if b.alive {
			b.y -= bulletSpeed
			if b.y < 0 {
				b.alive = false
			}
		}
	}

	newBullets := []*bullet{}
	for _, b := range bullets {
		if b.alive {
			newBullets = append(newBullets, b)
		}
	}
	bullets = newBullets
}

func updateEnemies() {
    // Spawn enemies at random positions at the top
    if rand.Intn(120) == 0 {
	    numEnemies := 5 // Number of enemies to spawn at once
	    for i := 0; i < numEnemies; i++ {
		    enemyX := float64(rand.Intn(screenWidth - obstacleWidth))
		    enemyY := 0.0
		    spawnAllowed := true
            for _, e := range enemies {
                if e.alive && e.x == enemyX {
                    spawnAllowed = false
                    break
                }
            }
	         if spawnAllowed {
                enemies = append(enemies, &enemy{
                    x:      enemyX,
                    y:      enemyY,
                    alive:  true,
                    deadly: true,
                })
            }
        }
    }





    for _, e := range enemies {
        if e.alive {
            e.y += 2.0 // Adjust the enemy speed as needed
            if e.y > screenHeight {
                e.alive = false
            }
        }
    }

    // Remove dead enemies
    newEnemies := []*enemy{}
    for _, e := range enemies {
        if e.alive {
            newEnemies = append(newEnemies, e)
        }
    }
    enemies = newEnemies
}

func updateObstacles() {
    if rand.Intn(120) == 0 {
        isDeadly := rand.Intn(2) == 0
        obstacleY := float64(rand.Intn(screenHeight - obstacleHeight))
        for obstacleY > playerY-obstacleHeight && obstacleY < playerY+playerHeight {
            obstacleY = float64(rand.Intn(screenHeight - obstacleHeight))
        }
        obstacles = append(obstacles, &obstacle{
            x:        float64(screenWidth),
            y:        obstacleY,
            deadly:   isDeadly,
            rotation: 0.0,
        })
    }

    for _, o := range obstacles {
        o.x -= 2.0
        if !o.deadly {
            o.rotation += 0.05 // Adjust rotation speed as needed
            if o.rotation >= 2*3.14159 {
                o.rotation = 0
            }
        }
        if o.x < -obstacleWidth {
            o.x = screenWidth
            obstacleY := float64(rand.Intn(screenHeight - obstacleHeight))
            for obstacleY > playerY-obstacleHeight && obstacleY < playerY+playerHeight {
                obstacleY = float64(rand.Intn(screenHeight - obstacleHeight))
            }
            o.y = obstacleY
            o.deadly = rand.Intn(2) == 0
            o.rotation = 0.0
        }
    }
}

  
    
  

func handleCollisions() {
	for _, b := range bullets {
		if !b.alive {
			continue
		}
		for _, o := range obstacles {
			if collision(b.x, b.y, bulletWidth, bulletHeight, o.x, o.y, obstacleWidth, obstacleHeight) {
				b.alive = false
				if o.deadly {
					o.deadly = false
					score++
				}
			}
		}

		for _, e := range enemies {
            if collision(b.x, b.y, bulletWidth, bulletHeight, e.x, e.y, obstacleWidth, obstacleHeight) {
                b.alive = false
                e.alive = false
                score++
            }
        }
}

	for _, o := range obstacles {
		if collision(playerX, playerY, playerWidth, playerHeight, o.x, o.y, obstacleWidth, obstacleHeight) {
			if o.deadly {
				gameOverSound.Rewind()
				gameOverSound.Play()
				gameOver = true
			} else {
				playerY = o.y - playerHeight
			}
		}
	}

	for _, e := range enemies {
        if collision(playerX, playerY, playerWidth, playerHeight, e.x, e.y, obstacleWidth, obstacleHeight) {
            if e.deadly {
                gameOverSound.Rewind()
                gameOverSound.Play()
                gameOver = true
            }
        }
    }
}

func collision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

func drawStartButton(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, startButtonX, startButtonY, startButtonWidth, startButtonHeight, color.White)
	ebitenutil.DebugPrintAt(screen, "START GAME", int(startButtonX)+10, int(startButtonY)+10)
}

func drawGameOverScreen(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, startButtonX, startButtonY, startButtonWidth, startButtonHeight, color.RGBA{255, 0, 0, 255})
	ebitenutil.DebugPrintAt(screen, "GAME OVER", int(startButtonX)+10, int(startButtonY)+10)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+strconv.Itoa(score), int(startButtonX)+10, int(startButtonY)+30)
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

func drawObstacles(screen *ebiten.Image) {
	for _, o := range obstacles {
		op := &ebiten.DrawImageOptions{}
		if o.deadly {
                        op.GeoM.Translate(o.x, o.y)
                        screen.DrawImage(enemyImage, op)
                } else {
                        op.GeoM.Translate(-float64(obstacleWidth)/2, -float64(obstacleHeight)/2)
                        op.GeoM.Rotate(o.rotation)
                        op.GeoM.Translate(o.x+float64(obstacleWidth)/2, o.y+float64(obstacleHeight)/2)
                        screen.DrawImage(obstacleImage, op)
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
                

func resetGame() {
	playerX = float64(screenWidth / 2)
	playerY = float64(screenHeight - playerHeight - 20)
	bullets = []*bullet{}
	obstacles = []*obstacle{}
	enemies   = []*enemy{}
	score = 0
	gameOver = false
}

