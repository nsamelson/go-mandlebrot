# go-mandlebrot
![image](https://user-images.githubusercontent.com/35641452/209961306-8b353fef-a0bf-446e-900d-a7040b9e73e1.png)

## Initilisation 
Run `docker-compose up --build` in the directory. 

## Structure
From the `docker-compose.yml` file, we can setup multiple backend workers. They will generate the image from specified parameters. 
A load balancer will also we spawned. He acts as a proxy and will communicate with the multiple backends to compose a response for the client. 
The `index.html` page will be our interface communicating with the load balancer. 

## Load balancing strategy 

Our loadbalancer receives coordinates of the center of the image requested (x,y) and a zoom factor. 
The loadbalancer will compute the boundaries of the image and the divide it in columns. There will be as much columns ad there are backends :3 backends means that the image will be split in 3 ranges (For 1000 pixels : 0-333, ...). The loadbalancer then computes the parameters he needs to send to the backends and sends the appropriate request. The backends are selected in a Round Robin fashion ; the one that received a request the longest time ago will be the first one selected. 
When the backend receives the request, it computes the mandlebrot set for his range and returns the result to the loadbalncer. 
The loadbalancer combines the results, generates the full image and sends it back to the client. 

## API
### Backend 

By default, 3 backends are spawned on the ports 3031, 3032 and 3033. 

The parameters to be sent in a get request are : 
- x_1 : The image left x coordinate in pixels
- x_2 : The image right x coordinate in pixels

- rMin : Minimum real part of the complex number.
- rMax : Maximum real part of the complex number.
- iMin : Minimum imaginary part of the complex number.
- iMax : Maximum imaginary part of the complex number.
- maxEsc : Maximum number of iterations

The backend will the compute the mandlebrot set between the range defined by x_1 and x_2 and return an image of that set.  

### Loadbalancer

Our loadbalancer runs at http://localhost:3030. 

To generate an image, there are three parameters to be sent : 
- x : x coordinate of the center of the requested image
- y : y coordinate of the center of the requested image
- z : zoom

The fractal is projected in a space of 1000 x 800. 

To begin with a centered image, you can call http://localhost:3030?x=500&y=400&z=1



## Used libraries
