# go-mandlebrot

## Initilisation 
Run `docker-compose up --build` in the directory. 

## Structure
From the `docker-compose.yml` file, we can setup multiple backend workers. They will generate the image from specified parameters. 
A load balancer will also we spawned. He acts as a proxy and will communicate with the multiple backends to compose a response for the client. 
The `index.html` page will be our interface communicating with the load balancer. 

## API
Our loadbalancer runs at http://localhost:3030. 
To generate an image, there are three parameters to be sent : 
- x : x coordinate
- y : y coordinate
- z : zoom

To begin with a centered image, you can call http://localhost:3030?x=500&y=400&z=1


## Load balancing strategy 


## Used libraries
