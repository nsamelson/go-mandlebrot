# go-mandlebrot

## To run app
- First go in `backends` folder and enter the command ```docker compose up```
- Then go in `loadbalancer` folder and enter the command ```./app --backends=http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034```
- Finally go into `frontend` folder and run the command ```air```
