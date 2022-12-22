package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"time"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	maxEsc = 100
	width  = 1000
	height = 800
	red    = 800
	green  = 600
	blue   = 700

	rMin   = -2.
	rMax   = .5
	iMin   = -1.
	iMax   = 1.
)

var (
	healthyBackends = []string{}

)
type Pair struct {
	body  []byte
	order int
}

// Send Async request and push it into the channel
func getResponse(url string, ch chan<- Pair, order int) {

	// send http request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	// get body content
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// add response to channel
	ch <- Pair{body, order}

}

// HealthCheck pings the backends and update the status
func healthCheck(backends []string) {
	

	// Run health check the checkBackends function to run every 3 minutes.
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {

		go checkBackends(backends)

	}
	
}

func checkBackends(backends []string){
	fmt.Println("Running health check")

	var _healthyBackends []string
	// Loop through the backend addresses and make an HTTP request to each one.
	for _, backend := range backends {
		resp, err := http.Get(backend)
		if err != nil {
			// If there is an error making the request, print an error message.
			fmt.Println("Error checking backend %s: %v\n", backend, err)
			continue
		}
		defer resp.Body.Close()

		fmt.Println("Backend %s: is alive", backend)
		// If the request is successful, add the backend to the list of healthy backends.
		_healthyBackends = append(_healthyBackends, backend)
	}

	healthyBackends = _healthyBackends
}


func loadBalancer(backends []string) http.HandlerFunc {
	// Use a map to track the number of requests sent to each backend
	backendRequestCount := make(map[string]int)

	return func(w http.ResponseWriter, r *http.Request) {

		// ALGO
		const width = 1000
		n_columns := 200

		// setup default parameters
		x_px := 500.
		y_px := 400.
		z_px := 1.

		// Check if the URL has a non-empty raw query string
		if r.URL.RawQuery != "" {
			fmt.Println("URL has query parameters :")

			// get vaules from url parameters
			values := r.URL.Query()

			for k, v := range values {
				fmt.Println(k, " => ", v)
			}

			if xStr := values.Get("x"); xStr != "" {
				x_px, _ = strconv.ParseFloat(xStr, 32)
			}
			if yStr := values.Get("y"); yStr != "" {
				y_px, _ = strconv.ParseFloat(yStr, 32)
			}
			if zStr := values.Get("z"); zStr != "" {
				z_px, _ = strconv.ParseFloat(zStr, 32)

				// avoid division by 0
				if z_px == 0 {
					z_px =1
				}
			}
		}

		// Transform x,y position into imaginary plane coordinates
		new_rMin := rMin + (x_px * z_px * (rMax - rMin) / (width * z_px)) - ((rMax - rMin) / (2 * z_px))
		new_rMax := rMin + (x_px * z_px * (rMax - rMin) / (width * z_px)) + ((rMax - rMin) / (2 * z_px))
		new_iMin := iMin + (y_px * z_px * (iMax - iMin) / (800 * z_px)) - ((iMax - iMin) / (2 * z_px))
		new_iMax := iMin + (y_px * z_px * (iMax - iMin) / (800 * z_px)) + ((iMax - iMin) / (2 * z_px))

		// Parameters in a string
		str_new_rMin := fmt.Sprintf("%f", new_rMin)
		str_new_rMax := fmt.Sprintf("%f", new_rMax)
		str_new_iMin := fmt.Sprintf("%f", new_iMin)
		str_new_iMax := fmt.Sprintf("%f", new_iMax)
		new_coords := "&rMin=" + str_new_rMin + "&rMax=" + str_new_rMax + "&iMin=" + str_new_iMin + "&iMax=" + str_new_iMax

		// Create channel
		ch := make(chan Pair)


		// divide the work by sending multiple columns for each node in async
		for x := 0; x < width/n_columns; x++ {

			backend := backends[rand.Intn(len(backends))]

			// Increment the request count for the selected backend
			backendRequestCount[backend]++

			// Set the url with the host corresponding to the backend
			new_url := backend + "/mandel/?" + "x_1=" + strconv.Itoa(x*n_columns) + "&x_2=" + strconv.Itoa((x+1)*n_columns) + new_coords

			// fmt.Println(new_url)
			// peer := serverPool.GetNextPeer()
			go getResponse(new_url, ch, x)

		}


		// draw image
		bounds := image.Rect(0, 0, width, height)
		b := image.NewNRGBA(bounds)
		draw.Draw(b, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)

		// Read channel and set color for each pixel
		for i := 0; i < width/n_columns; i++ {

			// get first Pair in channel and get rid of it
			channel := <-ch
			x := channel.order * n_columns

			var array [width][height]float64
			json.Unmarshal(channel.body, &array)

			for x_1 := 0; x_1 < n_columns; x_1++ {
				for y := 0; y < height; y++ {

					c := array[x_1][y]

					cr := uint8(float64(red) * c)
					cg := uint8(float64(green) * c)
					cb := uint8(float64(blue) * c)

					b.Set(x+x_1, y, color.NRGBA{R: cr, G: cg, B: cb, A: 255})

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
}


func main() {
	// Set up a slice of backend servers to load balance between
	var backendAddresses string
	var port int
	flag.StringVar(&backendAddresses, "backends", "", "comma-separated list of backend addresses")
	flag.IntVar(&port, "port", 3030, "Port to serve")

	// Parse the command-line flags
	flag.Parse()
	

	if len(backendAddresses) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	// Split the backend addresses on commas
	backends := strings.Split(backendAddresses, ",")
	healthyBackends = backends

	// run healthceck
	go healthCheck(backends)


	// Create a request multiplexer and register the load balancer handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", loadBalancer(backends))

	// Start the server
	fmt.Println("Load balancer listening on :3030")
	http.ListenAndServe(":3030", mux)

}