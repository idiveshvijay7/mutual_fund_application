```
# Mutual Fund Application

This is a full-stack application that includes a React-based frontend and two Golang-based backend services: a Payment Gateway and an RTA (Registrar and Transfer Agent) service.

## Prerequisites

- Node.js v18.13.0
- npm 9.2.0
- Go 1.22.2 (linux/amd64)

## Getting Started

### Client (Frontend):

1. Navigate to the client directory: `cd /client/mutual-fund-app`
2. Install dependencies: `npm install`
3. Start the client application: `npm start`
   
   The frontend will be available at [http://localhost:3000](http://localhost:3000)

### Server (Backend):

#### Payment Gateway:

1. Navigate to the payment gateway directory: `cd /server/payment-gateway`
2. Run the payment gateway server: `go run main.go`
   
   The payment gateway service will be available at [http://localhost:8080](http://localhost:8080)

#### RTA Service:

1. Navigate to the RTA service directory: `cd /server/rta-service`
2. Run the RTA service: `go run main.go`
   
   The RTA service will be available at [http://localhost:8081](http://localhost:8081)

Now you have the complete application running with the following components:

- Frontend: Running on [http://localhost:3000](http://localhost:3000)
- Payment Gateway: Running on [http://localhost:8080](http://localhost:8080)
- RTA Service: Running on [http://localhost:8081](http://localhost:8081)

You can now test the application by interacting with the frontend, which will communicate with the backend services.

## Versions

- Node.js: v18.13.0
- npm: 9.2.0
- Go: 1.22.2 (linux/amd64)

## Project Structure

The project is organized as follows:

- `/client`
  - `/mutual-fund-app`
    - `src/`
    - `package.json`
    - `...`
- `/server`
  - `/payment-gateway`
    - `main.go`
  - `/rta-service`
    - `main.go`

The client-side code is located in the `/client/mutual-fund-app` directory, and the server-side code is divided into two directories: `/server/payment-gateway` and `/server/rta-service`.

## Running Tests

You can run tests over the entire system to ensure its functionality.

## Conclusion

This Mutual Fund application is a full-stack project that includes a React-based frontend and two Golang-based backend services. The application is structured with a clear separation of concerns, making it easier to maintain and scale.
```
