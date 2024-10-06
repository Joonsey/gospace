package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1600
	screenHeight = 900
	renderWidth  = 1080
	renderHeight = 720
)

type Vec2 struct {
	X float64
	Y float64
}

type Entity struct {
	Id       uint
	position Vec2
}

type CelestialBodyDetails struct {
	entities map[uint]Entity
}

type CelestialBody struct {
	parent            *CelestialBody       // Pointer to the parent entity, nil if no parent
	orbit             *Orbit               // Orbital parameters, nil if not in an orbit
	last_update_time  int64                // Timestamp of the last position update
	position_on_orbit float64              // Current position in the orbit as a fraction of the period
	mass              float64              // Mass of the celestial body
	gravity           float64              // Gravitational pull force
	details           CelestialBodyDetails // The details of the celestial body
}

type Orbit struct {
	inclination float64 // Inclination of the orbit
	apoapsis    float64 // Apoapsis
	periapsis   float64 // Periapsis
	period      float64 // Orbital period
}

type Game struct {
	sun          CelestialBody
	earth        CelestialBody
	moon         CelestialBody
	focused_body *CelestialBody
}

func (g *Game) Update() error {
	g.sun.Update()
	g.earth.Update()
	g.moon.Update()

	o := g.earth.orbit
	a := (o.apoapsis + o.periapsis) / 2
	e := (a - o.periapsis) / a

	theta := g.earth.parent.position_on_orbit * 2 * math.Pi
	x, y := TrueAnomalyToPosition(a, e, o.inclination, theta)
	ax, ay := g.earth.GetPosition()
	cursor_x, cursor_y := ebiten.CursorPosition()
	dist_x, dist_y := ax+x-float64(cursor_x), ay+y-float64(cursor_y)

	log.Printf("%f %f\n", dist_x, dist_y)
	// 5 should be the the radius of the planet. or similar
	if math.Sqrt(dist_x*dist_x-dist_y*dist_y) < 5 {
		// we can do on-hover logic here
		// to draw cool shit when we hover the planet
		if ebiten.IsKeyPressed(ebiten.KeyS) {
			g.focused_body = &g.earth
		}
	}

	return nil
}

func TrueAnomalyToPosition(a, e, inclination, theta float64) (float64, float64) {
	// Calculate the radial distance for the current angle (true anomaly)
	r := a * (1 - e*e) / (1 + e*math.Cos(theta))

	// Position in the orbit (without inclination)
	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return x, y
}

func (o *Orbit) Draw(cx, cy float64, screen *ebiten.Image) {
	// Semi-major axis
	a := (o.apoapsis + o.periapsis) / 2
	// Eccentricity is pre-set

	// Draw the orbit by sampling points along the true anomaly (theta)
	numSteps := 1000

	e := (a - o.periapsis) / a

	for i := 0; i < numSteps; i++ {
		t1 := float64(i) / float64(numSteps) * 2 * math.Pi
		t2 := float64(i+1) / float64(numSteps) * 2 * math.Pi

		// Get positions for the two points on the ellipse
		x1, y1 := TrueAnomalyToPosition(a, e, o.inclination, t1)
		x2, y2 := TrueAnomalyToPosition(a, e, o.inclination, t2)

		// Draw the orbit path
		vector.StrokeLine(screen, float32(cx+x1), float32(cy+y1), float32(cx+x2), float32(cy+y2), 1, color.White, false)
	}
}

func (cb *CelestialBody) GetPosition() (x, y float64) {
	if cb.parent == nil {
		return screenWidth / 2, screenHeight / 2 // Center of the screen
	}

	theta := cb.parent.position_on_orbit * 2 * math.Pi
	a := (cb.parent.orbit.apoapsis + cb.parent.orbit.periapsis) / 2
	e := (a - cb.parent.orbit.periapsis) / a
	x, y = TrueAnomalyToPosition(a, e, cb.parent.orbit.inclination, theta)

	px, py := cb.parent.GetPosition()
	return x + px, y + py
}

func (cb *CelestialBody) Update() {
	// Semi-major axis
	a := (cb.orbit.apoapsis + cb.orbit.periapsis) / 2

	// Calculate the current true anomaly based on position in the orbit
	theta := cb.position_on_orbit * 2 * math.Pi
	e := (a - cb.orbit.periapsis) / a
	r := a * (1 - e*e) / (1 + e*math.Cos(theta))
	v := math.Sqrt(cb.mass * math.Pow(cb.gravity, 2) / 60 * (2/r - 1/a))

	// this is not real, magic trick to emulate the behaviour of kepler's second law
	cb.position_on_orbit += v * v / 100

	if cb.position_on_orbit >= 1 {
		cb.position_on_orbit = 0
	}

}

func (cb *CelestialBodyDetails) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "drawing details")
}

func (cb *CelestialBody) Draw(screen *ebiten.Image) {
	// Semi-major axis
	a := (cb.orbit.apoapsis + cb.orbit.periapsis) / 2

	// Calculate the current true anomaly based on position in the orbit
	theta := cb.position_on_orbit * 2 * math.Pi
	e := (a - cb.orbit.periapsis) / a

	x, y := TrueAnomalyToPosition(a, e, cb.orbit.inclination, theta)
	ax, ay := cb.GetPosition()

	cb.orbit.Draw(ax, ay, screen)

	vector.DrawFilledCircle(screen, float32(float64(ax)+x), float32(float64(ay)+y), 5, color.RGBA{255, 0, 0, 255}, false)
}

func (g *Game) DrawFocalPoint(screen *ebiten.Image) {
	cx, cy := screenWidth/2, screenHeight/2 // Center of the screen

	vector.DrawFilledCircle(screen, float32(float64(cx)), float32(float64(cy)), 15, color.RGBA{0, 255, 0, 255}, false)
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	if g.focused_body != nil {
		g.focused_body.details.Draw(screen)
	} else {
		g.sun.Draw(screen)
		g.earth.Draw(screen)
		g.moon.Draw(screen)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (_, _ int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}
	game.sun = CelestialBody{
		mass:    5.9,
		gravity: 8,
		orbit: &Orbit{
			inclination: 0,   // Inclination (tilt of the orbit)
			apoapsis:    300, // Apoapsis (furthest point)
			periapsis:   100, // Periapsis (closest point)
			period:      1.0, // Orbital period
		},
		position_on_orbit: 0.25, // Starting position on orbit
	}

	game.earth = CelestialBody{
		parent:  &game.sun,
		mass:    .9,
		gravity: 8,
		orbit: &Orbit{
			inclination: 0,   // Inclination (tilt of the orbit)
			apoapsis:    30,  // Apoapsis (furthest point)
			periapsis:   20,  // Periapsis (closest point)
			period:      1.0, // Orbital period
		},
		position_on_orbit: 0.25, // Starting position on orbit
	}

	// ignore that the moon shares orbit with he earth around the sun
	// also ignore that the sun rotates around a center point
	// this is not really very intuitive...
	game.moon = CelestialBody{
		parent:            &game.sun,
		mass:              .9,
		gravity:           3,
		orbit:             game.earth.orbit,
		position_on_orbit: .75,
	}

	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(renderWidth, renderHeight)
	ebiten.SetWindowTitle("gospace")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
