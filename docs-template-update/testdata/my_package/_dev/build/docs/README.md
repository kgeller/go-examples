<!-- This template can be used as a starting point for writing documentation for your new integration. For each section, fill in the details
described in the comments.

Find more detailed documentation guidelines in https://www.elastic.co/docs/extend/integrations/documentation-guidelines
-->

<!-- Do not remove header -->
{{ `{{header}}` }}
# 1Password Events Reporting Integration for Elastic

## Overview

The 1Password Events Reporting integration for Elastic enables collection of 1Password account activity and facilitates security monitoring and analysis.  This integration allows you to send your account activity to your security information and event management (SIEM) system using the 1Password Events API. Get reports about 1Password activity, such as sign-in attempts and item usage, while you manage all your company’s applications and services from a central location.

With 1Password Events Reporting and Elastic SIEM, you can:

-	Control your 1Password data retention
-	Build custom graphs and dashboards
-	Set up custom alerts that trigger specific actions
-	Cross-reference 1Password events with the data from other services

### Compatibility

This integration is compatible with 1Password Business.

### How it works

This integration uses the 1Password Events API to retrieve event data.

## What data does this integration collect?

The 1Password Events Reporting integration collects the following types of events:
* Sign-in attempts
* Item usages
* Audit events

### Supported use cases

This integration allows security teams to monitor and analyze 1Password activity for potential security threats.  It enables organizations to centralize their security monitoring and gain insights into user behavior.

## What do I need to use this integration?

Elastic Agent must be installed. For more details, check the Elastic Agent [installation instructions](docs-content://reference/fleet/install-elastic-agents.md). You can install only one Elastic Agent per host.

Elastic Agent is required to stream data from the syslog or log file receiver and ship the data to Elastic, where the events will then be processed via the integration's ingest pipelines.

You must be an owner or administrator of a 1Password Business account. You also need to configure the Events Reporting integration in 1Password.  [Learn how to set up the Elastic Events Reporting integration](https://support.1password.com/events-reporting).


## How do I deploy this integration?

### Onboard / configure

<!-- List the steps that will need to be followed in order to completely set up a working integration.
For integrations that support multiple input types, be sure to add steps for all inputs.
-->

### Validation

<!-- How can the user test whether the integration is working? Including example commands or test files if applicable -->

## Troubleshooting

For help with Elastic ingest tools, check [Common problems](https://www.elastic.co/docs/troubleshoot/ingest/fleet/common-problems).


## Scaling

For more information on architectures that can be used for scaling this integration, check the [Ingest Architectures](https://www.elastic.co/docs/manage-data/ingest/ingest-reference-architectures) documentation.

## Reference

### ECS field Reference

{{ `{{fields "item_usages"}}` }}

### Sample Event

{{ `{{event "item_usages"}}` }}

### Inputs used

<!-- List inputs used in this integration, and link to the documentation -->
These inputs can be used with this integration:
* ...

### API usage

This integration utilizes the 1Password Events API.

