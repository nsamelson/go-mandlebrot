# go-mandlebrot
![image](https://user-images.githubusercontent.com/35641452/209961306-8b353fef-a0bf-446e-900d-a7040b9e73e1.png)

## Initilisation 
Run `docker-compose up --build` in the directory. 

## Structure
From the `docker-compose.yml` file, we can setup multiple backend workers.  
They will generate the image from specified parameters.   
A load balancer will also we spawned. It acts as a proxy and will communicate with the multiple backends to compose a response for the client.  
The `index.html` page will be our interface communicating with the load balancer. 

## Load balancing strategy 

1. Our loadbalancer first receives coordinates of the center of the image requested (x,y) and a zoom factor (z).   

2. The loadbalancer uses these 3 coordinates to calculate the minimum and maximum values for the real and imaginary parts of the plane, which are represented by `rMin`, `rMax`, `iMin`, and `iMax`, respectively.

3. The loadbalancer is configured to divide the image into columns, with each column being a specific width. For example, if the image is 1000 pixels wide and the loadbalancer is configured to divide it into 5 columns, each column will be 200 pixels wide. 

4. The loadbalancer then computes the parameters he needs for each backends and sends the appropriate request. The backends are selected in a Round Robin fashion ; it is a scheduling algorithm that consists in assigning tasks of equal computing time to backends in a fixed order.   

5. When the backend receives the request, it computes the mandlebrot set for his set range of pixels and returns the result to the loadbalancer. 
6. Finally, the loadbalancer combines the results, generates the full image and sends it back to the client. 

## API
### Backend 

By default, 3 backends are spawned on the following addresses : 
- http://app1:3031;
- http://app2:3031;
- http://app3:3031.
 

The parameters to be sent in a `GET` request are : 
- x_1 : The image left x coordinate in pixels
- x_2 : The image right x coordinate in pixels

- rMin : Minimum real part of the complex number.
- rMax : Maximum real part of the complex number.
- iMin : Minimum imaginary part of the complex number.
- iMax : Maximum imaginary part of the complex number.
- maxEsc : Maximum number of iterations

The backend will then compute the mandlebrot set between the range defined by x_1 and x_2 and return an image of that set.  

### Loadbalancer

Our loadbalancer runs at http://localhost:3030. 

To generate an image, there are three parameters to be sent : 
- x : x coordinate of the center of the requested image
- y : y coordinate of the center of the requested image
- z : zoom

The fractal is projected in a space of 1000 x 800. 

To begin with a centered image, you can call http://localhost:3030?x=500&y=400&z=1



## Used libraries
The golang libraries used here are standard : 
- math: Standard math library in Go. 
- math/cmplx : Working with complex numbers. 

- encoding/json: Encoding and decoding data in JavaScript Object Notation (JSON) format.

- flag: Parsing command-line flags.

- fmt: Formatting and printing strings. 

- image: Working with images. 

- image/color: Working with colors in images.

- image/draw: Drawing geometric shapes, text, and other graphical elements onto images.

- image/png: Reading and writing PNG (Portable Network Graphics) image files.

- io/ioutil: Reading and writing data to files and other I/O sources.

- math/rand: Generating random numbers.

- net/http: Standard HTTP library.
- os: Standard OS library. It provides functions for interacting with the file system, environment variables, and other system-level resources.
- strconv: Int/string conversion. 
- time: Standard time library. It provides functions for representing and manipulating times, as well as for formatting and parsing time values as strings.

