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

type CelestialBody struct {
	parent           *CelestialBody // Pointer to the parent entity, nil if no parent
	orbit            *Orbit  // Orbital parameters, nil if not in an orbit
	last_update_time int64   // Timestamp of the last position update
}

type Orbit struct {
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
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.orbit.apoapsis += 10
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.orbit.apoapsis -= 10
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.orbit.inclination += 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.orbit.inclination -= 1
	}
	return nil
}

// TrueAnomalyToPosition calculates the x, y position on the orbit using the true anomaly.
func TrueAnomalyToPosition(a, e, inclination, theta float64) (float64, float64) {
	// Calculate the radial distance for the current angle (true anomaly)
	r := a * (1 - e*e) / (1 + e*math.Cos(theta))

	// Position in the orbit (without inclination)
	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return x, y
}

// DrawOrbit draws the orbit as an ellipse around the center of the screen, with inclination.
func (g *Game) DrawOrbit(screen *ebiten.Image) {
	cx, cy := screenWidth/2, screenHeight/2 // Center of the screen

	// Semi-major axis
	a := (g.orbit.apoapsis + g.orbit.periapsis) / 2
	// Eccentricity is pre-set

	// Draw the orbit by sampling points along the true anomaly (theta)
	numSteps := 1000

	e := (a - g.orbit.periapsis) / a

	for i := 0; i < numSteps; i++ {
		t1 := float64(i) / float64(numSteps) * 2 * math.Pi
		t2 := float64(i+1) / float64(numSteps) * 2 * math.Pi

		// Get positions for the two points on the ellipse
		x1, y1 := TrueAnomalyToPosition(a, e, g.orbit.inclination, t1)
		x2, y2 := TrueAnomalyToPosition(a, e, g.orbit.inclination, t2)

		// Draw the orbit path
		vector.StrokeLine(screen, float32(float64(cx)+x1), float32(float64(cy)+y1), float32(float64(cx)+x2), float32(float64(cy)+y2), 1, color.White, false)
	}
}

// DrawObjectOnOrbit draws the current position of the object on the orbit.
func (g *Game) DrawObjectOnOrbit(screen *ebiten.Image) {
	cx, cy := screenWidth/2, screenHeight/2 // Center of the screen

	// Semi-major axis
	a := (g.orbit.apoapsis + g.orbit.periapsis) / 2

	// Calculate the current true anomaly based on position in the orbit
	theta := g.orbit.position_on_orbit * 2 * math.Pi
	e := (a - g.orbit.periapsis) / a

	r := a * (1 - e*e) / (1 + e*math.Cos(theta))
	// TODO:
	// move this into it's on method and call it at each game update
	// and use inherit mass / gravity
	v := math.Sqrt(5.9 * math.Pow(8, 2) / 60 * (2/r - 1/a))

	g.orbit.position_on_orbit += v / 100

	if g.orbit.position_on_orbit >= 1 {
		g.orbit.position_on_orbit = 0
	}

	x, y := TrueAnomalyToPosition(a, e, g.orbit.inclination, theta)

	// Draw the object as a small circle
	vector.DrawFilledCircle(screen, float32(float64(cx)+x), float32(float64(cy)+y), 5, color.RGBA{255, 0, 0, 255}, false)
}

func (g *Game) DrawFocalPoint(screen *ebiten.Image) {
	cx, cy := screenWidth/2, screenHeight/2 // Center of the screen

	vector.DrawFilledCircle(screen, float32(float64(cx)), float32(float64(cy)), 15, color.RGBA{0, 255, 0, 255}, false)
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.

	screen.Fill(color.Black)

	g.DrawOrbit(screen)

	g.DrawFocalPoint(screen)

	g.DrawObjectOnOrbit(screen)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (_, _ int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}
	game.orbit = Orbit{
		inclination:       0,    // Inclination (tilt of the orbit)
		apoapsis:          300,  // Apoapsis (furthest point)
		periapsis:         100,  // Periapsis (closest point)
		period:            1.0,  // Orbital period
		position_on_orbit: 0.25, // Starting position on orbit
	}

	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(renderWidth, renderHeight)
	ebiten.SetWindowTitle("gospace")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
