package main

// https://kasvith.me/posts/lets-create-a-simple-lb-go/

import (
	"context"
	"math"

	// "math"
	// "sort"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Attempts int = iota
	Retry
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

// Backend holds the data about a server
type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

// SetAlive for this backend
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

// IsAlive returns true when backend is alive
func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

// ServerPool holds information about reachable backends
type ServerPool struct {
	backends []*Backend
	current  uint64
}

// AddBackend to the server pool
func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

// NextIndex atomically increase the counter and return an index
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// MarkBackendStatus changes a status of a backend
func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

// GetNextPeer returns next active peer to take a connection
func (s *ServerPool) GetNextPeer() *Backend {
	// loop entire backends to find out an Alive backend
	next := s.NextIndex()
	l := len(s.backends) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.backends)     // take an index by modding
		if s.backends[idx].IsAlive() { // if we have an alive backend, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

// HealthCheck pings the backends and update the status
func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

// GetAttemptsFromContext returns the attempts for request
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

// GetAttemptsFromContext returns the attempts for request
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

type Pair struct {
	body  []byte
	order int
}

// Send Async request and push it into the channel
func getResponse(url string, ch chan<- Pair, order int) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	ch <- Pair{body, order}

}

// lb load balances the incoming request
func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	const width = 1000
	n_columns := 100

	// get vaules from url parameters

	values := r.URL.Query()
	for k, v := range values {
		fmt.Println(k, " => ", v)
	}
	x, _ := strconv.Atoi(values["x"][0])
	y, _ := strconv.Atoi(values["y"][0])
	z, _ := strconv.ParseFloat(values["z"][0], 32)
	// x := 400
	// y := 400

	// compute ratio
	ratio_x := x / width
	ratio_y := y / 800 // heigth

	// // compute actual plan range
	r_range := rMin - rMax
	i_range := iMin - iMax

	// actual center
	center_r := r_range / 2
	center_i := i_range / 2

	// new center
	new_center_r := center_r * float64(ratio_x)
	new_center_i := center_i * float64(ratio_y)

	// new plan coordinates
	new_rMin := new_center_r - math.Abs(r_range)/(2*z)
	new_rMax := new_center_r + math.Abs(r_range)/(2*z)
	new_iMin := new_center_i - math.Abs(i_range)/(2*z)
	new_iMax := new_center_i + math.Abs(i_range)/(2*z)

	// Parameters in a string
	str_new_rMin := fmt.Sprintf("%f", new_rMin)
	str_new_rMax := fmt.Sprintf("%f", new_rMax)
	str_new_iMin := fmt.Sprintf("%f", new_iMin)
	str_new_iMax := fmt.Sprintf("%f", new_iMax)
	// test := strconv.FormatFloat(-1.125, 'g', 5, 64)
	new_coords := "&rMin=" + str_new_rMin + "&rMax=" + str_new_rMax + "&iMin=" + str_new_iMin + "&iMax=" + str_new_iMax

	// Create channel
	ch := make(chan Pair)

	// divide the work by sending multiple columns for each node in async
	for x := 0; x < width/n_columns; x++ {
		fmt.Println("AAAAAAAAAAAAA")
		log.Printf("url")
		// fmt.Print(new_coords)
		peer := serverPool.GetNextPeer()
		go getResponse(peer.URL.String()+"/mandel/?x_1="+strconv.Itoa(x*n_columns)+"&x_2="+strconv.Itoa((x+1)*n_columns)+new_coords, ch, x)

	}

	// draw image
	scale := width / (rMax - rMin)
	height := int(scale * (iMax - iMin))
	bounds := image.Rect(0, 0, width, height)
	b := image.NewNRGBA(bounds)
	draw.Draw(b, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)

	// Read channel and set color for each pixel
	for i := 0; i < width/n_columns; i++ {

		// get first Pair in channel and get rid of it
		channel := <-ch
		x := channel.order * n_columns

		var array [width][int(width / (rMax - rMin) * (iMax - iMin))]float64
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

// isAlive checks whether a backend is Alive by establishing a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}

// healthCheck runs a routine for check status of the backends every 2 mins
func healthCheck() {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			serverPool.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

var serverPool ServerPool

func Serve() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	// parse servers
	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		serverUrl, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			// after 3 retries, mark this backend as down
			serverPool.MarkBackendStatus(serverUrl, false)

			// if the same request routing for few attempts with different backends, increase the count
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		serverPool.AddBackend(&Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured server: %s\n", serverUrl)
	}

	// create http server
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	// start health checking
	go healthCheck()

	log.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Printf("starting things \n")

	Serve()
}
