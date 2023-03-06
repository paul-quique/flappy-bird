package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed res/*
var fsys embed.FS

func init() {
	table = LoadPNG("res/table.png")
	ball = LoadPNG("res/ball.png")
	building = LoadPNG("res/building.png")
	howtoplay = LoadPNG("res/howtoplay.png")

	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	bigfont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    100,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatalln(err)
	}
	restart()

}

func restart() {
	ballX = 250
	ballY = 400
	score = 0
	lastX = 0
	shifting = 0
	buildings = []Building{
		{250, 150, 0},
		{150, 250, 350},
		{100, 300, 700},
	}
	_ = randomBuilding()
	_ = randomBuilding()
	_ = randomBuilding()
	rand.Seed(time.Now().Unix())
	distance = 250
	score = 0
}

type Building struct {
	Top, Bottom float64
	X           float64
}

type Game struct{}

var player *audio.Player
var bigfont font.Face
var howtoplay *ebiten.Image
var table *ebiten.Image
var ball *ebiten.Image
var building *ebiten.Image
var ballX, ballY float64
var playerA float64
var buildingSize float64
var shifting float64
var buildings []Building
var acceleration float64 = 0.5
var speed float64 = 0
var paused bool = true
var lastX float64
var score int
var distance float64
var needsRestart bool = true

func (g *Game) Update() error {
	if ballColidding() {
		paused = true
	}
	if !paused {
		if ballY < 0 || ballY >= 800 {
			return nil
		}
		speed += acceleration
		ballY += speed
		shifting -= 2
		distance += 2
		if shifting == -350 {
			shifting = 0
			score += 1
			buildings = buildings[1:]
			buildings = append(buildings, randomBuilding())
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(ebiten.TouchIDs()) != 0 {
			speed = -10
		}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(ebiten.TouchIDs()) != 0 {
		paused = false
	}
	if needsRestart {
		if ebiten.IsKeyPressed(ebiten.KeyR) || len(ebiten.TouchIDs()) == 2 {
			go restartAndWait()
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(table, op)
	for _, b := range buildings {
		//top
		op.GeoM.Reset()
		op.GeoM.Scale(1, b.Top)
		op.GeoM.Translate(b.X-distance+250, 0)
		screen.DrawImage(building, op)
		//bottom
		op.GeoM.Reset()
		op.GeoM.Scale(1, b.Bottom)
		op.GeoM.Translate(b.X-distance+250, 800-b.Bottom)
		screen.DrawImage(building, op)
	}
	op.GeoM.Reset()
	op.GeoM.Translate(-250, -250)
	op.GeoM.Scale(0.1, 0.1)
	op.GeoM.Translate(ballX, ballY)
	screen.DrawImage(ball, op)
	op.GeoM.Reset()
	//score
	r := text.BoundString(bigfont, fmt.Sprint(score))
	tx := (r.Max.X - r.Min.X) / 2
	text.Draw(screen, fmt.Sprint(score), bigfont, 250-tx-4, 154, color.NRGBA{200, 200, 200, 255})
	text.Draw(screen, fmt.Sprint(score), bigfont, 250-tx-8, 150, color.Black)
	if paused {
		op.GeoM.Translate(0, 350)
		screen.DrawImage(howtoplay, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 500, 800
}

func main() {
	ebiten.SetWindowSize(500, 800)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowTitle("Pong")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

func LoadPNG(filename string) *ebiten.Image {
	infile, err := fsys.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	src, _, err := image.Decode(infile)
	if err != nil {
		panic(err)
	}
	return ebiten.NewImageFromImage(src)
}

func randomBuilding() Building {
	topSize := rand.Intn(400)
	b := Building{
		Top:    float64(topSize),
		Bottom: float64(400 - topSize),
		X:      float64(lastX),
	}
	lastX += 350
	return b
}

func ballColidding() bool {
	for _, b := range buildings {
		if distance >= b.X && distance <= b.X+300 {
			if ballY <= b.Top {
				ballY = b.Top
				return true
			} else if ballY >= 800-b.Bottom {
				ballY = 800 - b.Bottom
				return true
			}
		}
	}
	return false
}

func restartAndWait() {
	needsRestart = false
	restart()
	time.Sleep(1 * time.Second)
	needsRestart = true
}
