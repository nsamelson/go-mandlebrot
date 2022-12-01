package main

// https://firehydrant.com/blog/develop-a-go-app-with-docker-compose/
// https://blog.logrocket.com/creating-a-web-server-with-golang/
import (
	"encoding/json"
	"fmt"
	"strconv"

	// "image"
	// "image/color"
	// "image/draw"
	// "image/png"
	// "io/ioutil"
	// "os"
	"log"
	"math/cmplx"
	"net/http"
)

const (
	maxEsc = 100
	// rMin   = -2.
	// rMax   = .5
	// iMin   = -1.
	// iMax   = 1.

	// width = 1000
)

func mandelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	//TODO: get url parameters and copute mandel


	// link : http://localhost:3031/mandel/?x=21&y=22
	values := r.URL.Query()
	for k, v := range values {
		fmt.Println(k, " => ", v)
	}
	x_1, _ := strconv.Atoi(values["x_1"][0])
	x_2, _ := strconv.Atoi(values["x_2"][0])
	
	// insert mandelbrot
	var (
		x_range = x_2 - x_1
		width 	= 1000
		rMin   	= -2. 	
		rMax   	= .5 	
		iMin   	= -1. 	
		iMax   	= 1.	
	)
	
	scale 	:= float64(width) / (rMax - rMin)
	height 	:= int(scale * (iMax - iMin))
	// var mandelArray [width][int(width / (rMax - rMin) * (iMax - iMin))]float64

	var mandelArray = make([][]float64,width)
	for x := range mandelArray {
		mandelArray[x] = make([]float64,height)
	}

	for x := 0; x < x_range; x++ {
		for y := 0; y < height; y++ {
			c := mandelbrot(complex(
				float64(x_1)/scale+rMin,
				float64(y)/scale+iMin))
			mandelArray[x][y] = c
				
		}
		x_1 ++
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
