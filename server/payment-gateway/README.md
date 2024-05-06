# Payment Gateway

## Getting Started

### Pre requisite

- [Go v.1.22](https://go.dev/dl/) should be installed

### Running the server

- Navigate to payment gateway directory `cd payment-gateway`
- Copy sample.env to .env file `cp sample.env .env`
- Update the environment variables as per your requirements
- Run `go run main.go` inside payment gateway directory

### Functionalities

- Create Payment
- Fetch Payment
- Make Payment

## API Spec

Note {{baseUrl}} is configured via environment variables. By default it would be `http://localhost:8080`

### Create Payment

You can create a payment with the following request

URL - `POST {{baseUrl}}/payment`

Payload -

```json
{
  "accountNumber": "11200222",
  "ifscCode": "UBIT22222",
  "amount": 500,
  "redirectUrl": "http://localhost:3000"
}
```

Response -

```json
{
  "paymentLink": "http://localhost:8080/payment/pg/5235f102-3fc3-407e-8f8d-76f659d48325",
  "success": true
}
```

Note `redirectUrl` is the url to which you want to redirect after payment completion.
You can redirect the user to `paymentLink` for making the payment

### Fetch Payment

You can fetch the payment details using the following request

URL - `GET {{baseUrl}}/payment/{id}`

Response -

```json
{
  "id": "4209d078-2384-4652-984d-1106342b25a6",
  "accountNumber": "11200222",
  "ifscCode": "UBIT22222",
  "amount": 500,
  "redirectUrl": "http://localhost:3000",
  "status": "Created",
  "createdAt": "2024-04-05T01:20:48Z",
  "utr": null
}
```

Note `status` can be `Created`, `Success`, `Failed`.
Note `utr` will be set only after successful transaction.

### Make Payment

You can redirect your user to the url shared while creating the payment. In the screen you will get option to mark the transaction as successful or failed.
After clicking on the required button you will be redirected back to the configured redirect url (at the time of payment creation)

## Inconsistent server

We have configured servers in a way that by default nature you will receive Internal server error while making the requests. For payment callback after making the payment, you might receive internal server error, but the payment status will be properly updated at backend. You can control the behavior this error using `ERROR_RATE` environment variable

Postman collection is added for reference
