package main

import (
	"log"
	"image/color"
	"time"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

)

const (
    screenWidth   = 800
    screenHeight  = 600
    playerSpeed   = 4.0
    playerWidth   = 64
    playerHeight  = 64
    playerImagePath = "sprites/ship1.png"
    bulletImagePath = "sprites/bill1.png"
    enemyImagePath = "sprites/zombii.png"
    backgroundImagePath = "sprites/bg.png"
    bulletSpeed   = 8.0
    bulletWidth   = 8
    bulletHeight  = 7
    enemySpeed   = 2.0
    enemyWidth   = 64
    enemyHeight   = 64
    maxEnemies   = 7
    bulletSoundPath = "sounds/bullet.wav"
    gameOverSoundPath = "sounds/game_over.wav"
    spaceshipSoundPath = "sounds/spaceship.wav"
    startButtonWidth   = 200
    startButtonHeight   = 50
    maxLives    = 3
    heartImagePath = "sprites/heart.png"
    explosionImagePath   = "sprites/explosion.png"
    damagedSpaceshipImage1 = "sprites/damaged.png"
    damagedSpaceshipImage2 = "sprites/damaged3.png"
    flameImagePath  = "sprites/enemy_damaged.png"
    flameDuration   = 10
)

var (
    playerImage *ebiten.Image
    bulletImage *ebiten.Image
    enemyImage *ebiten.Image
    flameImage *ebiten.Image
    flames []*flame
    backgroundImage *ebiten.Image
    playerX   = float64(screenWidth / 2)
    playerY   = float64(screenHeight - playerHeight - 20)
    bullets   []*bullet
    enemies   []*enemy
    gameOver  bool
    score    int
    lives    int
    heartImage *ebiten.Image
    gameOverSoundPlayed bool
    startButtonX  = float64((screenWidth - startButtonWidth) / 2)
    startButtonY  = float64((screenHeight - startButtonHeight) / 2)
    //buttonX      = float64((screenWidth - buttonWidth) / 2)
    //restartButtonY  = float64((screenHeight - buttonHeight) / 2)
    //exitButtonY    = restartButtonY + buttonHeight + 10

    audioContext *audio.Context
    bulletSound *audio.Player
    gameOverSound *audio.Player
    spaceshipSound *audio.Player
    spaceshipSoundPlaying bool
    gameStarted bool
    explosionImage     *ebiten.Image
    damagedSpaceshipImages []*ebiten.Image
    showExplosion     bool
    explosionX, explosionY float64
    explosionTimer     int
)

type bullet struct {
    x, y float64
    frame int
    alive bool
}

type enemy struct {
    x, y float64
    alive bool
    flame bool
    flameTimer int
}

type flame struct {
    x, y       float64
    timer      int
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
    updateEnemies()
    updateFlames()
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
    op.GeoM.Translate(playerX, playerY)
    screen.DrawImage(playerImage, op)

    drawBullets(screen)
    drawEnemies(screen)
    drawFlames(screen)
    ebitenutil.DebugPrint(screen, "Score: "+strconv.Itoa(score))
    for i := 0; i < lives; i++ {
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(10+(i*30)), 40)
    screen.DrawImage(heartImage, op)
  }
  if showExplosion {
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(explosionX, explosionY)
    screen.DrawImage(explosionImage, op)
    explosionTimer--
    if explosionTimer <= 0 {
      showExplosion = false
    }
  }
}



func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return screenWidth, screenHeight
}

//func main() {
//   c := make(chan struct{}, 0)
//   go runGame()
//   <-c
//}

func main(){
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

    flameImage, _, err = ebitenutil.NewImageFromFile(flameImagePath)
    if err != nil {
        log.Fatal(err)
    }

    backgroundImage, _, err = ebitenutil.NewImageFromFile(backgroundImagePath)
    if err != nil {
        log.Fatal(err)
    }

    heartImage, _, err = ebitenutil.NewImageFromFile(heartImagePath)
  if err != nil {
    log.Fatal(err)
  }

   explosionImage, _, err = ebitenutil.NewImageFromFile(explosionImagePath)
  if err != nil {
    log.Fatal(err)
  }

  damagedSpaceshipImages = make([]*ebiten.Image, 2)
  damagedSpaceshipImages[0], _, err = ebitenutil.NewImageFromFile(damagedSpaceshipImage1)
  if err != nil {
    log.Fatal(err)
  }
  damagedSpaceshipImages[1], _, err = ebitenutil.NewImageFromFile(damagedSpaceshipImage2)
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

     spaceshipSound, err = loadSound(audioContext, spaceshipSoundPath)
    if err != nil {
        log.Fatal(err)
    }

    initializeEnemies()
    // rand.Seed(time.Now().UnixNano())
    ebiten.SetWindowSize(screenWidth, screenHeight)
    ebiten.SetWindowTitle("Side-Scrolling Shooter Game")

    go spawnEnemies()

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

func initializeEnemies() {
    rand.Seed(time.Now().UnixNano())
    for i := 0; i < maxEnemies; i++ {
        x := float64(rand.Intn(screenWidth - enemyWidth))
        y := float64(rand.Intn(screenHeight/2 - enemyHeight))
        enemies = append(enemies, &enemy{
            x:   x,
            y:   y,
            alive: true,
	    flame : false,
	    flameTimer: 0,
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
                x:   x,
                y:   y,
                alive: true,
		flame : false,
		flameTimer: 0,
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
    moving := false
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

    if moving {
        if !spaceshipSoundPlaying {
            spaceshipSound.Rewind()
            spaceshipSound.Play()
            spaceshipSoundPlaying = true
        }
    } else {
        if spaceshipSoundPlaying {
            spaceshipSound.Pause()
            spaceshipSoundPlaying = false
        }
    }
}

func handleShooting() {
    if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        bulletSound.Rewind()
        bulletSound.Play()
        bulletX := playerX + (playerWidth / 2) - (bulletWidth + 0.5 / 2)
        bulletY := playerY - bulletHeight

        bullets = append(bullets, &bullet{
            x:   bulletX,
            y:   bulletY,
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
                enemies = append(enemies[:i], enemies[i+1:]...)
                lives--
                showExplosion = true
                explosionX = playerX + playerWidth/2 - float64(explosionImage.Bounds().Dx())/2
                explosionY = playerY - float64(explosionImage.Bounds().Dy())/2
                explosionTimer = 6
                if lives <= 0 {
                    gameOver = true
                }else{
                    playerImage = damagedSpaceshipImages[maxLives-lives-1]
                }

            }
        }
      if gameOver && !gameOverSoundPlayed {
        gameOverSound.Rewind()
        gameOverSound.Play()
        gameOverSoundPlayed = true
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
		e.alive = false
                score +=1
		flames = append(flames, &flame{
                    x: e.x,
                    y: e.y,
                    timer: flameDuration,
                })
                break
            }
        }
    }
}

func updateFlames() {
    for i := len(flames) - 1; i >= 0; i-- {
        f := flames[i]
        f.timer--
        if f.timer <= 0 {
            flames = append(flames[:i], flames[i+1:]...)
        }
    }
}

func drawFlames(screen *ebiten.Image) {
    for _, f := range flames {
        op := &ebiten.DrawImageOptions{}
        op.GeoM.Translate(f.x, f.y)
        screen.DrawImage(flameImage, op)
    }
}

//func checkGameOver() {
//   for _, e := range enemies {
//       if e.alive && e.y > screenHeight {
//           gameOver = true
//           break
//       }
//   }
//}

func resetGame() {
    playerX = float64(screenWidth / 2)
    playerY = float64(screenHeight - playerHeight - 20)
    bullets = []*bullet{}
    enemies = []*enemy{}
    score  = 0
    gameOver = false
    gameOverSoundPlayed = false
    lives  = maxLives
    //playerImage = "sprites/ship1.png"
    showExplosion = false
    explosionTimer = 0
     var err error
  playerImage, _, err = ebitenutil.NewImageFromFile(playerImagePath)
  if err != nil {
    log.Fatal(err)
  }
    initializeEnemies()
    go spawnEnemies()

     if spaceshipSoundPlaying {
            spaceshipSound.Pause()
            spaceshipSoundPlaying = false
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
            screen.DrawImage(enemyImage, op)
        }else if e.flame {
		op := &ebiten.DrawImageOptions{}
            op.GeoM.Translate(e.x, e.y)
            screen.DrawImage(flameImage, op)
        }
    }
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
