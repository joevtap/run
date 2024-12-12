package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	GUARD = iota
	SCOUT
)

type Vector struct {
	X, Y float64
}

func (v Vector) Add(v2 Vector) Vector {
	return Vector{v.X + v2.X, v.Y + v2.Y}
}

func (v Vector) Sub(v2 Vector) Vector {
	return Vector{v.X - v2.X, v.Y - v2.Y}
}

func (v Vector) Mul(s float64) Vector {
	return Vector{v.X * s, v.Y * s}
}

func (v Vector) Div(s float64) Vector {
	return Vector{v.X / s, v.Y / s}
}

func (v Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector) Normalize() Vector {
	length := v.Length()
	if length == 0 {
		return Vector{0, 0}
	}
	return v.Div(length)
}

type Entity struct {
	Position Vector
	Velocity Vector
	Speed    float64
}

func (e *Entity) Update() {
	if e.Position.X-10 < 0 {
		e.Position.X = 10
		e.Velocity.X = 0
	}
	if e.Position.X+10 > 5000 {
		e.Position.X = 4990
		e.Velocity.X = 0
	}
	if e.Position.Y-10 < 0 {
		e.Position.Y = 10
		e.Velocity.Y = 0
	}
	if e.Position.Y+10 > 5000 {
		e.Position.Y = 4990
		e.Velocity.Y = 0
	}
}

type Player struct {
	Entity
	Life int
}

func (p *Player) Update(g *Game) {
	p.Entity.Update()

	p.Velocity = Vector{0, 0}

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		p.Velocity.Y--
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		p.Velocity.Y++
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		p.Velocity.X--
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		p.Velocity.X++
	}

	p.Position = p.Position.Add(p.Velocity.Normalize().Mul(p.Speed))

}

func (p Player) Draw(screen *ebiten.Image, g *Game) {
	if g.Debug {
		for i := range g.Enemies {
			direction := Vector{
				X: g.Enemies[i].Position.X - p.Position.X,
				Y: g.Enemies[i].Position.Y - p.Position.Y,
			}

			endPosition := Vector{
				X: p.Position.X + direction.X,
				Y: p.Position.Y + direction.Y,
			}

			clr := color.RGBA{100, 100, 100, 255}

			if g.Enemies[i].State == "CHASE" {
				clr = color.RGBA{255, 0, 0, 255}
			}

			vector.StrokeLine(
				screen,
				float32(p.Position.X+g.Camera.Position.X),
				float32(p.Position.Y+g.Camera.Position.Y),
				float32(endPosition.X+g.Camera.Position.X),
				float32(endPosition.Y+g.Camera.Position.Y),
				1,
				clr,
				true,
			)
		}
	}

	clr := uint8(255 * (float64(p.Life) / 100))

	vector.DrawFilledCircle(screen, float32(p.Position.X+g.Camera.Position.X), float32(p.Position.Y+g.Camera.Position.Y), 10, color.RGBA{0, clr, clr, clr}, true)
}

type Enemy struct {
	Entity
	Origin       Vector
	Type         int
	State        string
	Bullets      int
	LastShotTime time.Time
}

func (e *Enemy) Update(p *Player, g *Game) {
	e.Entity.Update()
	e.Velocity = Vector{0, 0}

	if e.CollidesWith(*p) {
		e.Velocity = Vector{0, 0}

		if time.Since(e.LastShotTime) > 500*time.Millisecond {
			p.Life -= 50
			e.LastShotTime = time.Now()
		}
		return
	}

	distance := p.Position.Sub(e.Position).Length()

	if distance < 800 {
		e.State = "CHASE"

		if distance < 400 {
			e.Velocity = p.Position.Sub(e.Position).Normalize().Mul(e.Speed)

			e.Position = e.Position.Add(e.Velocity)
		} else {
			if e.Bullets > 0 {
				if time.Since(e.LastShotTime) > g.ShootingCoolDown {
					e.Shoot(*p, g)
					e.Bullets--
					e.LastShotTime = time.Now()
				}
			}
		}

	} else {
		if e.Type == GUARD {
			e.State = "PATROL"

			if e.Position.Sub(e.Origin).Length() > 5 {
				e.Velocity = e.Origin.Sub(e.Position).Normalize().Mul(e.Speed / 2)

				e.Position = e.Position.Add(e.Velocity)
			} else {
				e.Position = e.Origin

				e.Bullets = 100

				e.State = "IDLE"
			}
		}
	}
}

func (e Enemy) Draw(screen *ebiten.Image, g *Game) {
	switch e.State {
	case "CHASE":
		vector.DrawFilledCircle(screen, float32(e.Position.X+g.Camera.Position.X), float32(e.Position.Y+g.Camera.Position.Y), 10, color.RGBA{250, 0, 0, 255}, true)
	case "PATROL":
		vector.DrawFilledCircle(screen, float32(e.Position.X+g.Camera.Position.X), float32(e.Position.Y+g.Camera.Position.Y), 10, color.RGBA{0, 250, 0, 255}, true)
	case "IDLE":
		vector.DrawFilledCircle(screen, float32(e.Position.X+g.Camera.Position.X), float32(e.Position.Y+g.Camera.Position.Y), 10, color.RGBA{100, 100, 100, 255}, true)
	}

	if g.Debug {
		vector.StrokeCircle(screen, float32(e.Position.X+g.Camera.Position.X), float32(e.Position.Y+g.Camera.Position.Y), 400, 1, color.RGBA{50, 50, 50, 255}, true)

		vector.StrokeCircle(screen, float32(e.Position.X+g.Camera.Position.X), float32(e.Position.Y+g.Camera.Position.Y), 800, 1, color.RGBA{100, 50, 50, 255}, true)
	}
}

func (e Enemy) CollidesWith(p Player) bool {
	return e.Position.Sub(p.Position).Length() < 20
}

func (e Enemy) Shoot(p Player, g *Game) {
	if e.Bullets == 0 {
		return
	}

	direction := p.Position.Sub(e.Position).Normalize()

	projectile := Projectile{
		Position: e.Position,
		Velocity: direction.Mul(20),
	}

	g.Projectiles = append(g.Projectiles, projectile)
}

type Camera struct {
	Position Vector
}

func (c *Camera) FollowTarget(target Entity, screenW, screenH int) {
	factor := 0.05

	targetPosition := Vector{
		X: -target.Position.X + float64(screenW)/2,
		Y: -target.Position.Y + float64(screenH)/2,
	}

	c.Position = c.Position.Add(targetPosition.Sub(c.Position).Mul(factor))
}

type Projectile struct {
	Position Vector
	Velocity Vector
}

func (p *Projectile) Update(g *Game) {
	p.Position = p.Position.Add(p.Velocity)
}

func (p Projectile) Draw(screen *ebiten.Image, g *Game) {
	vector.DrawFilledCircle(screen, float32(p.Position.X+g.Camera.Position.X), float32(p.Position.Y+g.Camera.Position.Y), 2, color.RGBA{255, 255, 255, 255}, true)
}

func (p *Projectile) ShouldBeRemoved() bool {
	return p.Position.X < 0 || p.Position.X > 5000 || p.Position.Y < 0 || p.Position.Y > 5000
}

func (p Projectile) CollidesWith(e Entity) bool {
	return p.Position.Sub(e.Position).Length() < 10
}

type Game struct {
	Player           Player
	Enemies          []Enemy
	ShootingCoolDown time.Duration
	Projectiles      []Projectile
	Camera           Camera

	LastSpawnTime time.Time
	SpawnRate     time.Duration
	Debug         bool
	Key1Pressed   bool
}

func (g *Game) Update() error {
	g.Player.Update(g)

	if ebiten.IsKeyPressed(ebiten.KeyControlLeft) {
		if ebiten.IsKeyPressed(ebiten.Key1) && !g.Key1Pressed {
			g.Debug = !g.Debug
			g.Key1Pressed = true
		}
	} else {
		g.Key1Pressed = false
	}

	if g.Player.Life <= 0 {
		g.Enemies = []Enemy{}
		g.Player.Life = 100
		g.Projectiles = []Projectile{}
		g.LastSpawnTime = time.Now().Add(-999 * time.Hour)
		g.Debug = false
		g.Init()
	}

	for i := range g.Enemies {
		g.Enemies[i].Update(&g.Player, g)
	}

	g.Camera.FollowTarget(g.Player.Entity, 1920, 1080)

	for i := len(g.Projectiles) - 1; i >= 0; i-- {
		g.Projectiles[i].Update(g)
		if g.Projectiles[i].CollidesWith(g.Player.Entity) {

			g.Player.Life -= 10

		}

		if g.Projectiles[i].ShouldBeRemoved() {
			g.Projectiles = append(g.Projectiles[:i], g.Projectiles[i+1:]...)
		}

	}

	if time.Since(g.LastSpawnTime) > 10*time.Second {
		for i := 0; i < 10; i++ {
			randomX := rand.Intn(5000)
			randomY := rand.Intn(5000)

			position := Vector{float64(randomX), float64(randomY)}

			g.Enemies = append(g.Enemies, Enemy{
				Entity: Entity{
					Position: position,
					Velocity: Vector{0, 0},
					Speed:    6,
				},
				Origin:  position,
				Type:    GUARD,
				State:   "IDLE",
				Bullets: 100,
			})
		}

		g.LastSpawnTime = time.Now()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i := range g.Enemies {
		g.Enemies[i].Draw(screen, g)
	}

	g.Player.Draw(screen, g)

	for i := range g.Projectiles {
		g.Projectiles[i].Draw(screen, g)
	}

	vector.StrokeRect(
		screen,
		float32(0+g.Camera.Position.X),
		float32(0+g.Camera.Position.Y),
		float32(5000),
		float32(5000),
		1,
		color.RGBA{255, 255, 255, 255},
		true,
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1920, 1080
}

func (g *Game) Init() {
	randomX := rand.Intn(5000)
	randomY := rand.Intn(5000)

	g.Debug = false

	g.Player = Player{
		Entity: Entity{
			Position: Vector{float64(randomX), float64(randomY)},
			Velocity: Vector{0, 0},
			Speed:    8,
		},
		Life: 100,
	}

	g.Camera = Camera{
		Position: Vector{0, 0},
	}

	g.ShootingCoolDown = 200 * time.Millisecond

}

func main() {
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Run")

	game := &Game{}

	game.Init()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
