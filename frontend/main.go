package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	maxEsc = 100
	rMin   = -2.
	rMax   = .5
	iMin   = -1.
	iMax   = 1.
	width  = 1000
	red    = 800
	green  = 600
	blue   = 700
)

func mandelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	//TODO: get url parameters and copute mandel

	// link : http://localhost:3031/mandel/?x=21&y=22
	values := r.URL.Query()
	for k, v := range values {
		fmt.Println(k, " => ", v)
	}

	// send request to loadbalancer and wait for response
	// resp, err := http.Get("http://localhost:3030/mandel")

	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// draw image
	scale := width / (rMax - rMin)
	height := int(scale * (iMax - iMin))
	bounds := image.Rect(0, 0, width, height)
	b := image.NewNRGBA(bounds)
	draw.Draw(b, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)

	// create an array with the size of the image
	// var mandelArray [width][int(width / (rMax - rMin) * (iMax - iMin))]float64
	// json.Unmarshal([]byte(body), &mandelArray)

	//TODO: send requests for each line of pixels : https://vanhunteradams.com/DE1/Mandelbrot/Mandelbrot_Implementation.html
	for x := 0; x < width; x++ {
		if x/10 == 0 {
			resp, err := http.Get("http://localhost:3030/mandel/?x=" + strconv.Itoa(x))
			if err != nil {
				log.Fatalln(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			var array [10][int(width / (rMax - rMin) * (iMax - iMin))]float64
			json.Unmarshal([]byte(body), &array)

			for y := 0; y < height; y++ {

				c := array[x][y]

				cr := uint8(float64(red) * c)
				cg := uint8(float64(green) * c)
				cb := uint8(float64(blue) * c)

				b.Set(x, y, color.NRGBA{R: cr, G: cg, B: cb, A: 255})

			}
		}

	}

	// create image
	f, err := os.Create("mandelbrot.png")
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = png.Encode(f, b); err != nil {
		fmt.Println(err)
	}
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}

	// render image
	buf, _ := ioutil.ReadFile("mandelbrot.png")
	w.Header().Set("Content-Type", "mandelbrot.png")
	w.Write(buf)
}

func main() {
	http.HandleFunc("/mandel/", mandelHandler)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
