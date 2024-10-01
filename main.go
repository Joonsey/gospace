package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1600
	screenHeight = 900
	renderWidth  = 1080
	renderHeight = 720
)

type Vec2 struct {
	X int
	Y int
}

type Entity struct {
	entity_id        int     // Unique ID for the entity
	parent           *Entity // Pointer to the parent entity, nil if no parent
	orbit            *Orbit  // Orbital parameters, nil if not in an orbit
	last_update_time int64   // Timestamp of the last position update
}

type Orbit struct {
	semi_major_axis   float64 // Semi-major axis of the orbit
	eccentricity      float64 // Eccentricity of the orbit
	inclination       float64 // Inclination of the orbit
	apoapsis          float64 // Apoapsis
	periapsis         float64 // Periapsis
	period            float64 // Orbital period
	position_on_orbit float64 // Current position in the orbit as a fraction of the period
}

// Game implements ebiten.Game interface.
type Game struct {
	orbit Orbit
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	g.orbit.position_on_orbit += 0.001
	if g.orbit.position_on_orbit >= 1.0 {
		g.orbit.position_on_orbit = 0.0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.orbit.eccentricity += 0.01
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.orbit.eccentricity -= 0.01
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.orbit.inclination += 0.01
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.orbit.inclination -= 0.01
	}
	return nil
}

func (g *Game) DrawOrbit(screen *ebiten.Image) {
	cx, cy := renderWidth/2, renderHeight/2 // Center of the screen

	// Semi-minor axis derived from eccentricity: b = a * sqrt(1 - e^2)
	semi_minor_axis := g.orbit.semi_major_axis * math.Sqrt(1-g.orbit.eccentricity*g.orbit.eccentricity)

	// Draw the orbit (approximated as an ellipse)
	numSteps := 100 // Number of segments to draw the ellipse
	for i := 0; i < numSteps; i++ {
		t1 := float64(i) / float64(numSteps) * 2 * math.Pi
		t2 := float64(i+1) / float64(numSteps) * 2 * math.Pi

		x1 := float64(cx) + g.orbit.semi_major_axis*math.Cos(t1)
		y1 := float64(cy) + semi_minor_axis*math.Sin(t1)
		x2 := float64(cx) + g.orbit.semi_major_axis*math.Cos(t2)
		y2 := float64(cy) + semi_minor_axis*math.Sin(t2)

		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 1, color.White, false)
	}
}

func (g *Game) DrawObjectOnOrbit(screen *ebiten.Image) {
	cx, cy := renderWidth/2, renderHeight/2 // Center of the screen

	// Semi-minor axis derived from eccentricity: b = a * sqrt(1 - e^2)
	semi_minor_axis := g.orbit.semi_major_axis * math.Sqrt(1-g.orbit.eccentricity*g.orbit.eccentricity)

	// Current angle based on the position on the orbit
	angle := g.orbit.position_on_orbit * 2 * math.Pi

	// Calculate the current position of the object
	x := float64(cx) + g.orbit.semi_major_axis*math.Cos(angle)
	y := float64(cy) + semi_minor_axis*math.Sin(angle)

	// Draw the object as a small circle
	vector.DrawFilledCircle(screen, float32(x), float32(y), 5, color.RGBA{255, 0, 0, 255}, false)
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.

	screen.Fill(color.Black)

	g.DrawOrbit(screen)

	g.DrawObjectOnOrbit(screen)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return renderWidth, renderHeight
}

func main() {
	game := &Game{}
	game.orbit = Orbit{
		semi_major_axis:   200,  // Semi-major axis (size of orbit)
		eccentricity:      0.5,  // Orbit eccentricity (how elliptical the orbit is)
		inclination:       0,    // Inclination (tilt of the orbit)
		apoapsis:          300,  // Apoapsis (furthest point)
		periapsis:         100,  // Periapsis (closest point)
		period:            1.0,  // Orbital period
		position_on_orbit: 0.25, // Starting position on orbit
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("gospace")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
