# go-mandlebrot

## Initilisation 
Run `docker-compose up --build` in the directory. 

## Structure
From the `docker-compose.yml` file, we can setup multiple backend workers. They will generate the image from specified parameters. 
A load balancer will also we spawned. He acts as a proxy and will communicate with the multiple backends to compose a response for the client. 
The `index.html` page will be our interface communicating with the load balancer. 

## Load balancing strategy 
