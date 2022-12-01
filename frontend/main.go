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
	nodes := 100

	for x := 0; x < width; x++ {
		if x%nodes == 0 {
			resp, err := http.Get("http://localhost:3030/mandel/?x_1=" + strconv.Itoa(x) + "&x_2=" + strconv.Itoa(x+nodes))
			if err != nil {
				log.Fatalln(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			var array [width][int(width / (rMax - rMin) * (iMax - iMin))]float64
			json.Unmarshal([]byte(body), &array)

			for x_1 := 0; x_1 < nodes; x_1++ {
				for y := 0; y < height; y++ {

					c := array[x_1][y]

					cr := uint8(float64(red) * c)
					cg := uint8(float64(green) * c)
					cb := uint8(float64(blue) * c)

					b.Set(x+x_1, y, color.NRGBA{R: cr, G: cg, B: cb, A: 255})

				}
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
