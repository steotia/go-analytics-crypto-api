[![Go Report Card](https://goreportcard.com/badge/github.com/steotia/go-analytics-crypto-api)](https://goreportcard.com/report/github.com/steotia/go-analytics-crypto-api) [![Maintainability](https://api.codeclimate.com/v1/badges/a81bbfae1dcea67e5074/maintainability)](https://codeclimate.com/github/steotia/go-analytics-crypto-api/maintainability)

# go-analytics-crypto-api

This repo, sets up an API, which returns data about Volume, High and Low growth rates of cryptocurrency pairs in a 5 minute interval.

## Installation
1. `docker build -t go-analytics-crypto-api -f Dockerfile .` => this creates local Docker image
2. Follow installation instructions in https://github.com/steotia/go-analytics-crypto-scraper to setup the underlying MongoDB 
and the system which populates the DB with cryptocurrency exchange data. The instructions mentioned there also sets up the
API container via docker compose

## API details
The API returns `json` data for 5 min periods, between the `from` and `to` parameters. If the data is not present in the DB,
it returns blanks, else it returns back the calculated growth percentages for Volume, High and Low values of the
crypto currency pairs.

### Parameters
`from` is the start time, `to` is the end time of interest, `format` is the return format. `from` cannot be later than `to`.

#### Sample output
`curl 'http://localhost:12345/export/analytics?from=2019-07-14T18:00:00&to=2019-07-14T19:00:00&format=json'`
