# Public API for Concordances (public-concordances-api)

Provides a public API for Concordances stored in the Neo4j graph database.

This Go app provides an endpoint for the retrieval of concordance data given one or more parameters. This app
directly depends upon a connection to a Neo4j database.

## Build

```shell
go get github.com/Financial-Times/public-concordances-api
cd $GOPATH/src/github.com/Financial-Times/public-concordances-api
go build -mod=readonly .
```

## Running the tests

Run unit tests only:

```shell script
go test -v -race ./...
```

To run the integration tests, you must provide `GITHUB_USERNAME` and
`GITHUB_TOKEN` because the service depends on internal repositories:

```shell
GITHUB_USERNAME="<user-name>" GITHUB_TOKEN="<personal-access-token>" \
docker-compose -f docker-compose-tests.yml up -d --build && \
docker logs -f public-concordances-api_test-runner_1 && \
docker-compose -f docker-compose-tests.yml down
```

## API Endpoints

Based on the following [google doc](https://docs.google.com/a/ft.com/document/d/1onyyb-XoByB00RQNZvjNoL_IsO_eHKe-vOpUuAVHyJE)

- GET `/concordances?conceptId={thingUri}` - Returns a list of all identifiers for given concept
- GET `/concordances?conceptId={thingUri}&conceptId={thingUri}...` - Returns a list of all identifiers for each concept provided
- GET `/concordances?authority={identifierUri}&identifierValue{identifierValue}` - Returns the apiUrl that matches the corresponding identifier
- GET `/concordances?authority={identifierUri}&identifierValue={identifierValue}&identifierValue={identifierValue}` - Returns a list of all apiUrls for the corresponding identifiers

## Admin endpoints

- GET `/__health`
- GET `/__build-info`
- GET `/__gtg`

## Error handling

[Runbook](https://runbooks.ftops.tech/public-concordances-api) - [Panic guide](https://sites.google.com/a/ft.com/universal-publishing/ops-guides/panic-guides/concordances-read)

- The service expects at least 1 conceptId or (authority + identifierValue pair) parameter and will respond with an Error HTTP status code if these are not provided.
- The service will respond with Error HTTP codes if both a conceptId is presented with an authority parameter or if an identifierValue is presented without the authority parameter.
- The service will never respond with Error HTTP status codes if none of the conceptId's or identifierValues are present in concordance,
instead it will return an empty array of Concepts or Identifiers.
