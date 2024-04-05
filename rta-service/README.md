# RTA Service

## Getting Started

### Pre requisite

- [Go v.1.22](https://go.dev/dl/) should be installed
- Payment Gateway (present in `../payment-gateway` diretory)

### Running the server

- Navigate to rta service directory `cd rta-service`
- Copy sample.env to .env file `cp sample.env .env`
- Update the environment variables as per your requirements
- Run `go run main.go` inside rta service directory

### Functionalities

- Create Order
- Fetch Order
- Get Market value

## API Spec

Note {{baseUrl}} is configured via environment variables. By default it would be `http://localhost:8081`

### Create Payment

You can create a order with the following request

URL - `POST {{baseUrl}}/order`

Payload -

```json
{
  "fund": "Arbitrage Fund 1",
  "amount": 500,
  "paymentID": "3cd4267a-75f6-40f7-86dc-5ec4802e7ca9"
}
```

Response -

```json
{
  "data": {
    "id": "54dca7e0-2316-4697-9640-83ac12a38328",
    "fund": "Arbitrage Fund 1",
    "amount": 500,
    "units": 0,
    "pricePerUnit": 0,
    "status": "Submitted",
    "paymentID": "3cd4267a-75f6-40f7-86dc-5ec4802e7ca9",
    "submittedAt": null,
    "succeededAt": null,
    "failedAt": null
  },
  "success": true
}
```

You have to make a successful payment to create an order. Payment Id should be passed along with the create request to create the order. If Payment is not successful, order will fail. If rta service not able to connect to payment gateway the order will again fail. When you submit the order, you will get the order details, rta serice will take some time to process the order. You can control that time via environment variable `PROCESS_ORDER_RATE` (value in seconds). At the time of processing order, based on the nav the units will be allotted.
You can keep calling fetch order to get the latest status of the order.

### Fetch Order

You can fetch the order details using the following request

URL - `GET {{baseUrl}}/order/{id}`

Response -

```json
{
  "id": "26bda99a-79d0-4563-9d8f-3b47da6554f3",
  "fund": "Arbitrage Fund 1",
  "amount": 500,
  "units": 25.993197491666407,
  "pricePerUnit": 19.235801988589643,
  "status": "Succeeded",
  "paymentID": "3cd4267a-75f6-40f7-86dc-5ec4802e7ca9",
  "submittedAt": "2024-04-05T02:39:28Z",
  "succeededAt": "2024-04-05T02:39:33Z",
  "failedAt": null
}
```

If the market value of the fund has changed when order is being processed, there will be a slight difference in the units allotted. If the units allotted is less than 1, then the order will be rejected.

### Fetch Market Value

You can fetch the order details using the following request

URL - `GET {{baseUrl}}/market-value/{Fund Name}`

Note, only the strategies provided in `strategies.json` at the root of the assignment is supported.
Market value get updated by a thread. You can control rate at which it shoudl change via environment variable `NAV_UPDATE_RATE` (value in seconds).

## Inconsistent server

By intentional you will receive Internal server error while making the requests. For payment callback after making the payment, you might receive internal server error, but the payment status will be properly updated at backend. You can control the behavior this error using `ERROR_RATE` environment variable

Postman collection is added for reference
