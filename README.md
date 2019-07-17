[![Go Report Card](https://goreportcard.com/badge/github.com/steotia/go-analytics-crypto-api)](https://goreportcard.com/report/github.com/steotia/go-analytics-crypto-api) [![Maintainability](https://api.codeclimate.com/v1/badges/a81bbfae1dcea67e5074/maintainability)](https://codeclimate.com/github/steotia/go-analytics-crypto-api/maintainability)

# go-analytics-crypto-api

This repo, sets up an API, which returns data about Volume, High and Low growth rates of cryptocurrency pairs in a 5 minute interval. The API hits a mongo backend, where the time series data is persisted.

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

### Parameter Validation
The API does basic Parameter validation

#### Sample output
`curl 'http://localhost:12345/export/analytics?from=2019-07-14T18:00:00&to=2019-07-14T19:00:00&format=json'`

## TODO
- Make the period size configurable. Presently it is a fixed 5 minute interval.
- Fetch only the pairs configured. Presently all pairs are fetched. However, the change should be straightforward, as the documents can be filter-able via the pair name.
- DB Indexing
- Refactoring to interface the persistence layer where any backend can be plugged in (not just mongo)
- General refactoring, for e.x. the Handler code is pretty long and can be further refactores, better interfacing in the code
- Less hardcoding, there are some strings still lying around in the code as hardcoded strings
- More test coverage! and documentation of exported functions, etc

