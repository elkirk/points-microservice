This repo contains the files necessary to run the microservice that solves the Fetch Backend Engineering exercise.

### Running the app
There are a couple of ways to run and interact with this app. These instructions assume a basic understanding of github and Go specifically.

#### Option 1: Run locally
Clone the repo to your local machine. The app can be run with the command:

```console
go run main.go
```

The app can also be run with the executable:

```console
./points
```

You should see something similar to the below, which indicates the app is running locally:
```console
❯ go run main.go
Starting server on :3000
```

### Interacting with the app
There are 4 endpoints that accept requests:
- `add-transaction` accepts POST requests representing points transactions that have the format described below and created a database entry for each:
    ```
    { "payer": "DANNON", "points": 1000, "timestamp": "2020-11-02T14:00:00Z" }
    { "payer": "UNILEVER", "points": 200, "timestamp": "2020-10-31T11:00:00Z" }
    { "payer": "DANNON", "points": -200, "timestamp": "2020-10-31T15:00:00Z" }
    { "payer": "MILLER COORS", "points": 10000, "timestamp": "2020-11-01T14:00:00Z" }
    { "payer": "DANNON", "points": 300, "timestamp": "2020-10-31T10:00:00Z" }
    ```
    - You can use Postman to make requests, or, from the command line (make sure the app is running first):
    ```console
    curl -d "{ "payer": "DANNON", "points": 1000, "timestamp": "2020-11-02T14:00:00Z" }" -X POST localhost:3000/add-transaction
    ```
- `balance` accepts GET requests and returns total point balances for each payer with at least one transaction in the database. Use Postman, or the following command:
    ```console
    curl localhost:3000/balance
    ```
- `balance/{payer}` accepts GET requests and returns the total point balance for **only** the payer passed as a URL parameter.
    ```console
    curl localhost:3000/balance/dannon
    ```
- `spend` accepts POST requests representing a points spend request that have the format described below. If there are enough points in the database to cover the spend, this endpoint returns JSON describing the payers and their contributions to the spend request. Points are spent oldest-to-newest. If there are not enough points to cover the request, a response will be returned with status code 422 indicating there are insufficient points.

Requests have the format:

    ```
    { "points": 5000 }
    ```
Since points are spent oldest-to-newest, the spend request described above will return the following:

    ```
    [
    {
        "payer": "DANNON",
        "points": -100
    },
    {
        "payer": "UNILEVER",
        "points": -200
    },
    {
        "payer": "MILLER COORS",
        "points": -4700
    }
    ]
    ```
    - New records will be added to the database reflecting these negative transactions.
    - Make a request with Postman or the following command:
    ```console
    curl -d { "points": 5000 } -X PUT localhost:3000/spend
    ```
- `check` accepts GET requests and returns a list of JSON entries representing every transaction currently stored in the database. Hitting this endpoint after the requests described above will return:
    ```
    [
    {
        "timestamp": "2020-11-02T14:00:00",
        "_id": 1,
        "payer": "DANNON",
        "points": 1000
    },
    {
        "timestamp": "2020-10-31T11:00:00",
        "_id": 2,
        "payer": "UNILEVER",
        "points": 200
    },
    {
        "timestamp": "2020-10-31T15:00:00",
        "_id": 3,
        "payer": "DANNON",
        "points": -200
    },
    {
        "timestamp": "2020-11-01T14:00:00",
        "_id": 4,
        "payer": "MILLER COORS",
        "points": 10000
    },
    {
        "timestamp": "2020-10-31T10:00:00",
        "_id": 5,
        "payer": "DANNON",
        "points": 300
    },
    {
        "timestamp": "2022-08-19T16:18:01.859377",
        "_id": 6,
        "payer": "DANNON",
        "points": -100
    },
    {
        "timestamp": "2022-08-19T16:18:01.878487",
        "_id": 7,
        "payer": "UNILEVER",
        "points": -200
    },
    {
        "timestamp": "2022-08-19T16:18:01.885005",
        "_id": 8,
        "payer": "MILLER COORS",
        "points": -4700
    }
    ]
    ```
    - Make a request with Postman or the following command:
    ```console
    curl localhost:3000/check
    ```

