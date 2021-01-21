# UPP - Public Concordances API

Public Concordances API provides a public API for retrieving concordances stored in the Neo4j graph database.

## Code

public-concordances-api

## Primary URL

<https://api.ft.com/concordances>

## Service Tier

Platinum

## Lifecycle Stage

Production

## Delivered By

content

## Supported By

content

## Known About By

- dimitar.terziev
- elitsa.pavlova
- kalin.arsov
- ivan.nikolov
- miroslav.gatsanoga
- marina.chompalova

## Host Platform

AWS

## Architecture

The service provides the following endpoints for the retrieval of concordance data:

- GET `/concordances?conceptId={thingUri}` - returns a list of all identifiers for a given concept.
- GET `/concordances?conceptId={thingUri}&conceptId={thingUri}...` - returns a list of all identifiers for each provided concept.
- GET `/concordances?authority={identifierUri}&identifierValue{identifierValue}` - returns the apiURL that matches the corresponding identifier.
- GET `/concordances?authority={identifierUri}&identifierValue={identifierValue}&identifierValue={identifierValue}` - returns a list of all apiURLs for the corresponding identifiers.

## Contains Personal Data

No

## Contains Sensitive Data

No

## Dependencies

- upp-neo4j-cluster

## Failover Architecture Type

ActiveActive

## Failover Process Type

FullyAutomated

## Failback Process Type

FullyAutomated

## Failover Details

The service is deployed in both Delivery clusters. The failover guide for the cluster is located here:
<https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/delivery-cluster>

## Data Recovery Process Type

NotApplicable

## Data Recovery Details

The service does not store data, so it does not require any data recovery steps.

## Release Process Type

PartiallyAutomated

## Rollback Process Type

Manual

## Release Details

The release is triggered by making a Github release which is then picked up by a Jenkins multibranch pipeline. The Jenkins pipeline should be manually started in order for it to deploy the helm package to the Kubernetes clusters.

## Key Management Process Type

NotApplicable

## Key Management Details

There is no key rotation procedure for this system.

## Monitoring

Service in UPP K8S delivery clusters:

- Delivery-Prod-EU health: <https://upp-prod-delivery-eu.upp.ft.com/__health/__pods-health?service-name=public-concordances-api>
- Delivery-Prod-US health: <https://upp-prod-delivery-us.upp.ft.com/__health/__pods-health?service-name=public-concordances-api>

## First Line Troubleshooting

[First Line Troubleshooting guide](https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting)

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.

Additional information can be found in the [Google Sites panic guide](https://sites.google.com/a/ft.com/universal-publishing/ops-guides/panic-guides/concordances-read).
