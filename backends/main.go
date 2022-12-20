package main

// https://firehydrant.com/blog/develop-a-go-app-with-docker-compose/
// https://blog.logrocket.com/creating-a-web-server-with-golang/
import (
	"encoding/json"
	"fmt"
	"strconv"
	"log"
	"math/cmplx"
	"net/http"
)

const (
	maxEsc = 100
	rMin   = -2.
	rMax   = .5
	iMin   = -1.
	iMax   = 1.

	// width = 1000
)

func mandelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// get the parameters from the url : http://localhost:3031/mandel/?x=21&y=22
	values := r.URL.Query()

	// range of columns of pixels to compute
	x_1, _ := strconv.Atoi(values["x_1"][0])
	x_2, _ := strconv.Atoi(values["x_2"][0])

	// new plan coordinates
	rMin, _ := strconv.ParseFloat(values["rMin"][0], 32)
	rMax, _ := strconv.ParseFloat(values["rMax"][0], 32)
	iMin, _ := strconv.ParseFloat(values["iMin"][0], 32)
	iMax, _ := strconv.ParseFloat(values["iMax"][0], 32)

	// insert mandelbrot
	var (
		x_range = x_2 - x_1
		width   = 1000
		// rMin    = -2. / 1
		// rMax    = 0.5 / 1
		// iMin    = -1. /1
		// iMax    = 1.
	)

	scale := float64(width) / (rMax - rMin)
	height := int(scale * (iMax - iMin))
	// var mandelArray [width][int(width / (rMax - rMin) * (iMax - iMin))]float64

	var mandelArray = make([][]float64, width)
	for x := range mandelArray {
		mandelArray[x] = make([]float64, height)
	}

	// compute mandelbrot
	for x := 0; x < x_range; x++ {
		for y := 0; y < height; y++ {
			c := mandelbrot(complex(
				float64(x_1)/scale+rMin,
				float64(y)/scale+iMin))
			mandelArray[x][y] = c

		}
		x_1++
	}
	// send json response
	json.NewEncoder(w).Encode(mandelArray)

}

func main() {
	http.HandleFunc("/mandel/", mandelHandler)

	fmt.Printf("Starting server at port 3031\n")
	if err := http.ListenAndServe(":3031", nil); err != nil {
		log.Fatal(err)
	}
}

func mandelbrot(a complex128) float64 {
	i := 0
	for z := a; cmplx.Abs(z) < 2 && i < maxEsc; i++ {
		z = z*z + a
	}
	return float64(maxEsc-i) / maxEsc
}
