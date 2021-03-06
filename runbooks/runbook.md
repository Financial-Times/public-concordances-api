<!--
    Written in the format prescribed by https://github.com/Financial-Times/runbook.md.
    Any future edits should abide by this format.
-->
# UPP - Public Concordances API

Public Concordances API provides a public API for retrieving concordances stored in the Neo4j graph database.

## Code

public-concordances-api

## Primary URL

https://api.ft.com/concordances

## Service Tier

Platinum

## Lifecycle Stage

Production

## Host Platform

AWS

## Architecture

The service provides the following endpoints for the retrieval of concordance data:

*   GET `/concordances?conceptId={thingUri}` - returns a list of all identifiers for a given concept.
*   GET `/concordances?conceptId={thingUri}&conceptId={thingUri}...` - returns a list of all identifiers for each provided concept.
*   GET `/concordances?authority={identifierUri}&identifierValue{identifierValue}` - returns the apiURL that matches the corresponding identifier.
*   GET `/concordances?authority={identifierUri}&identifierValue={identifierValue}&identifierValue={identifierValue}` - returns a list of all apiURLs for the corresponding identifiers.

## Contains Personal Data

No

## Contains Sensitive Data

No

<!-- Placeholder - remove HTML comment markers to activate
## Can Download Personal Data
Choose Yes or No

...or delete this placeholder if not applicable to this system
-->

<!-- Placeholder - remove HTML comment markers to activate
## Can Contact Individuals
Choose Yes or No

...or delete this placeholder if not applicable to this system
-->

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

<!-- Placeholder - remove HTML comment markers to activate
## Heroku Pipeline Name
Enter descriptive text satisfying the following:
This is the name of the Heroku pipeline for this system. If you don't have a pipeline, this is the name of the app in Heroku. A pipeline is a group of Heroku apps that share the same codebase where each app in a pipeline represents the different stages in a continuous delivery workflow, i.e. staging, production.

...or delete this placeholder if not applicable to this system
-->

## Key Management Process Type

NotApplicable

## Key Management Details

There is no key rotation procedure for this system.

## Monitoring

Service in UPP K8S delivery clusters:

*   Delivery-Prod-EU health: <https://upp-prod-delivery-eu.upp.ft.com/__health/__pods-health?service-name=public-concordances-api>
*   Delivery-Prod-US health: <https://upp-prod-delivery-us.upp.ft.com/__health/__pods-health?service-name=public-concordances-api>

## First Line Troubleshooting

[First Line Troubleshooting guide](https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting)

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.

Additional information can be found in the [Google Sites panic guide](https://sites.google.com/a/ft.com/universal-publishing/ops-guides/panic-guides/concordances-read).