package main

import(
	"fmt"
	"image"
	"os"
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

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return pixel.PictureDataFromImage(img), nil
}

func run() {
	y := 300.
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

	plane := imdraw.New(nil)
	plane.Color = colornames.Lightgray
	plane.Push(pixel.V(0, 900))
	plane.Push(pixel.V(1480, 480))
	plane.Rectangle(0)

	// Prepare font to print text on screen
	txt := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// Other labels 
	others := text.New(pixel.V(55, 482), txt)
	others.Color = colornames.Red
	fmt.Fprintf(others, "Front")

	others.Dot = pixel.V(1035, 482)
	fmt.Fprintf(others, "Back")

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

	imd := imdraw.New(nil)

	for !win.Closed() {	
		imd.Clear()
		//win.Clear(colornames.Black)
		plane.Draw(win)

		others.Draw(win, pixel.IM.Scaled(others.Orig, 1.3))
		seatNumsTop.Draw(win, pixel.IM.Scaled(seatNumsTop.Orig, 1.8))
		seatRows.Draw(win, pixel.IM.Scaled(seatRows.Orig, 3))

		// Passengers
		imd.Color = colornames.Limegreen
		imd.Push(pixel.V(163, y))
		imd.Circle(10, 0)
		imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
	a := make([]int, 50)
	for i := range a {
		a[i] = i
	}
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i]})
	fmt.Println(a[:5])
}