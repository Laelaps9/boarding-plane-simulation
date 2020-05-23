package main

import(
	"fmt"
	"math/rand"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type Passenger struct {
	PosX int
	PosY int
	Seat int
}

func generatePasses(size int) ([]Passenger) {
	var passes [144]int
	passengers := make([]Passenger, size)

	// Create boarding passes
	for i := range passes {
		passes[i] = i
	}
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(passes), func(i, j int) { passes[i], passes[j] = passes[j], passes[i]})

	// Assign boarding passes
	for i := range passengers {
		passengers[i].Seat = passes[i]
	}

	return passengers
}

func getSeatColumn(pass Passenger) (int){
	return pass.Seat / 6
}

func getSeatRow(pass Passenger) (int){
	return pass.Seat % 6
}

func createWindow() (*pixelgl.Window) {
	// Specify configuration window
	cfg := pixelgl.WindowConfig {
		Title: "Plane Boarding Simulator",
		Bounds: pixel.R(0, 0, 1430, 900),
		VSync: true,
	}
	// Create a new window
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	win.Clear(colornames.Black)

	return win
}

func drawPlane() (*imdraw.IMDraw) {
	plane := imdraw.New(nil)

	plane.Color = colornames.Lightgray
	plane.Push(pixel.V(0, 900))
	plane.Push(pixel.V(1480, 480))
	plane.Rectangle(0)

	plane.Color = colornames.Red

	// Front Entrance
	plane.Push(pixel.V(10, 482))
	plane.Push(pixel.V(50, 478))
	plane.Rectangle(0)

	// Back Entrance
	plane.Push(pixel.V(1370, 482))
	plane.Push(pixel.V(1410, 478))
	plane.Rectangle(0)

	// Draw seats
	plane.Color = colornames.Darkgray
	for i := 0; i < 3; i++ {
		for j := 1; j <= 24; j++ {
			plane.Push(pixel.V(float64(52 * j + 40), float64(860 - 50 * i)))
			plane.Push(pixel.V(float64(52 * j + 80), float64(820 - 50 * i)))
			plane.Rectangle(0)
		}
	}
	
	for i := 0; i < 3; i++ {
		for j := 1; j <= 24; j++ {
			plane.Push(pixel.V(float64(52 * j + 40), float64(660 - 50 * i)))
			plane.Push(pixel.V(float64(52 * j + 80), float64(620 - 50 * i)))
			plane.Rectangle(0)
		}
	}

	return plane
}

func drawLabels() (*text.Text, *text.Text, *text.Text) {
	// Prepare font to print text on screen
	txt := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// Other labels 
	others := text.New(pixel.V(55, 482), txt)
	others.Color = colornames.Red
	fmt.Fprintf(others, "Front")

	others.Dot = pixel.V(1035, 482)
	fmt.Fprintf(others, "Back")

	// Seats Columns
	seatNumsTop := text.New(pixel.V(104, 870), txt)
	seatNumsTop.Color = colornames.Black

	// Seat Rows
	seatRows := text.New(pixel.V(65, 827), txt)
	seatRows.Color = colornames.Darkblue

	labelString := ""
	charOrigin := pixel.V(105, 870)

	// Add labels for seat columns
	for i := 0; i < 24; i++ {
		if i <= 8 {
			charOrigin = pixel.V(float64(105 + i * 29), 870)
		} else {
			charOrigin = pixel.V(float64(100 + i * 29), 870)
		}
		
		seatNumsTop.Dot = charOrigin
		labelString = fmt.Sprintf("%d", i + 1)
		fmt.Fprintf(seatNumsTop, labelString)
	}

	// Add labels for seat rows
	for i := 0; i < 6; i++ {
		if i < 3 {
			charOrigin = pixel.V(65, float64(827 - i * 17))
		} else {
			charOrigin = pixel.V(65, float64(811 - i * 17))
		}

		seatRows.Dot = charOrigin
		labelString = fmt.Sprintf("%c", 70 - i)
		fmt.Fprintf(seatRows, labelString)
	}

	return seatNumsTop, seatRows, others
}

func run() {
	win := createWindow()
	plane := drawPlane()
	seatNumsTop, seatRows, others := drawLabels()

	passengers := generatePasses(5)
	fmt.Println(passengers)

	// Passengers
	y := 503.
	pass := imdraw.New(nil)
	pass.Color = colornames.Limegreen

	// Entrance 1 starting point (30, 503)

	// Seats x = 52 + 60 * seatNumber

	for !win.Closed() {	
		pass.Clear()
		win.Clear(colornames.Black)
		plane.Draw(win)

		others.Draw(win, pixel.IM.Scaled(others.Orig, 1.3))
		seatNumsTop.Draw(win, pixel.IM.Scaled(seatNumsTop.Orig, 1.8))
		seatRows.Draw(win, pixel.IM.Scaled(seatRows.Orig, 3))

		if y < 690 {
			y += 2
		}

		// Passengers
		pass.Push(pixel.V(30, y))
		pass.Circle(15, 0)
		pass.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}