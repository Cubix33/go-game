package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	screenWidth  = 800
	screenHeight = 600
	playerSpeed  = 300.0
)

var (
	playerPos   = pixel.V(screenWidth/2, screenHeight/2)
	playerSize  = pixel.V(50, 50)
	playerColor = colornames.Red

	bullets []*bullet
	enemies []*enemy

	batch       *pixel.Batch
	enemySprite *pixel.Sprite

	gameStarted bool
	gamePaused  bool

	score        int
	hurdleCount  int
	maxHurdles   = 5
	gameOver     bool
)

type bullet struct {
	pos pixel.Vec
	dir pixel.Vec
}

type enemy struct {
	pos  pixel.Vec
	vel  pixel.Vec
	size float64
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "3D Shooter Game",
		Bounds: pixel.R(0, 0, screenWidth, screenHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatal(err)
	}

	loadAssets()

	last := time.Now()

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		win.Clear(colornames.Black)

		handleInput(win, dt)
		update(dt)
		render(win)

		win.Update()
	}
}

func loadAssets() {
	for i := 0; i < 5; i++ {
		enemy := &enemy{
			pos:  pixel.V(rand.Float64()*float64(screenWidth), rand.Float64()*float64(screenHeight)),
			vel:  pixel.V(rand.Float64()*100-50, rand.Float64()*100-50),
			size: 20,
		}
		enemies = append(enemies, enemy)
	}

	enemyPic := pixel.MakePictureData(pixel.R(-1, -1, 1, 1))
	imd := imdraw.New(nil)
	imd.Color = colornames.White
	imd.Push(pixel.V(-0.5, -0.5))
	imd.Push(pixel.V(0.5, 0.5))
	imd.Rectangle(0)
	enemySprite = pixel.NewSprite(enemyPic, enemyPic.Bounds())
}

func handleInput(win *pixelgl.Window, dt float64) {
	if win.JustPressed(pixelgl.KeySpace) {
		gameStarted = true
	}
	if win.JustPressed(pixelgl.KeyP) {
		gamePaused = !gamePaused
	}
	if win.JustPressed(pixelgl.KeyEscape) {
		win.SetClosed(true)
	}

	if gameStarted && !gamePaused && !gameOver {
		if win.Pressed(pixelgl.KeyLeft) {
			playerPos.X -= playerSpeed * dt
		}
		if win.Pressed(pixelgl.KeyRight) {
			playerPos.X += playerSpeed * dt
		}
		if win.Pressed(pixelgl.KeyUp) {
			playerPos.Y += playerSpeed * dt
		}
		if win.Pressed(pixelgl.KeyDown) {
			playerPos.Y -= playerSpeed * dt
		}

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			dir := win.MousePosition().Sub(playerPos).Unit()
			bullets = append(bullets, &bullet{pos: playerPos, dir: dir})
		}
	}
}

func update(dt float64) {
	if gameStarted && !gamePaused && !gameOver {
		for i := len(bullets) - 1; i >= 0; i-- {
			b := bullets[i]
			b.pos = b.pos.Add(b.dir.Scaled(playerSpeed * dt * 2))
			if !pixel.R(0, 0, screenWidth, screenHeight).Contains(b.pos) {
				bullets = append(bullets[:i], bullets[i+1:]...)
			}
		}

		for _, e := range enemies {
			e.pos = e.pos.Add(e.vel.Scaled(dt))
			if e.pos.X < 0 || e.pos.X > screenWidth {
				e.vel.X = -e.vel.X
			}
			if e.pos.Y < 0 || e.pos.Y > screenHeight {
				e.vel.Y = -e.vel.Y
			}
		}

		for _, b := range bullets {
			for j, e := range enemies {
				if b.pos.Sub(e.pos).Len() < 10 {
					bullets = append(bullets[:j], bullets[j+1:]...)
					enemies = append(enemies[:j], enemies[j+1:]...)
					score += 10
				}
			}
		}

		if hurdleCount >= maxHurdles {
			gameOver = true
		}
	}
}

func render(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = playerColor
	imd.Push(playerPos.Sub(playerSize.Scaled(0.5)), playerPos.Add(playerSize.Scaled(0.5)))
	imd.Rectangle(0)
	imd.Draw(win)

	for _, b := range bullets {
		imd.Color = colornames.Yellow
		imd.Push(b.pos)
		imd.Circle(3, 0)
		imd.Draw(win)
	}

	for _, e := range enemies {
		enemySprite.Draw(win, pixel.IM.Moved(e.pos).Scaled(e.pos, 0.1))
	}

	if !gameStarted {
		drawStartUI(win)
	}
	if gamePaused {
		drawPauseUI(win)
	}
	if gameOver {
		drawGameOverUI(win)
	}

	drawScore(win)
}

func drawStartUI(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = colornames.White
	imd.Push(pixel.V(screenWidth/2-100, screenHeight/2-25))
	imd.Push(pixel.V(screenWidth/2+100, screenHeight/2+25))
	imd.Rectangle(0)
	imd.Draw(win)

	txt := pixelgl.NewText(pixelgl.TextAtlas{})
	txt.Color = colornames.Black
	txt.WriteString("Start Game")
	txt.Draw(win, pixel.IM.Scaled(txt.Bounds().Center(), 2).Moved(pixel.V(screenWidth/2-txt.Bounds().W()/2, screenHeight/2-txt.Bounds().H()/2)))
}

func drawPauseUI(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = colornames.White
	imd.Push(pixel.V(screenWidth/2-100, screenHeight/2-25))
	imd.Push(pixel.V(screenWidth/2+100, screenHeight/2+25))
	imd.Rectangle(0)
	imd.Draw(win)

	txt := pixelgl.NewText(pixelgl.TextAtlas{})
	txt.Color = colornames.Black
	txt.WriteString("Paused - Press P to Resume")
	txt.Draw(win, pixel.IM.Scaled(txt.Bounds().Center(), 1.5).Moved(pixel.V(screenWidth/2-txt.Bounds().W()/2, screenHeight/2-txt.Bounds().H()/2)))
}

func drawGameOverUI(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = colornames.White
	imd.Push(pixel.V(screenWidth/2-100, screenHeight/2-25))
	imd.Push(pixel.V(screenWidth/2+100, screenHeight/2+25))
	imd.Rectangle(0)
	imd.Draw(win)

	txt := pixelgl.NewText(pixelgl.TextAtlas{})
	txt.Color = colornames.Black
	txt.WriteString("Game Over - Press Esc to Exit")
	txt.Draw(win, pixel.IM.Scaled(txt.Bounds().Center(), 1.5).Moved(pixel.V(screenWidth/2-txt.Bounds().W()/2, screenHeight/2-txt.Bounds().H()/2)))
}

func drawScore(win *pixelgl.Window) {
	txt := pixelgl.NewText(pixelgl.TextAtlas{})
	txt.Color = colornames.Black
	txt.WriteString(fmt.Sprintf("Score: %d", score))
	txt.Draw(win, pixel.IM.Moved(pixel.V(10, screenHeight-20)))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	pixelgl.Run(run)
}
