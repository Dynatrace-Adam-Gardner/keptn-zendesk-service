# Keptn Zendesk Service
## ⚠️This repository is now deprecated, will receive no further updates and users should replace this service with [these instructions](https://artifacthub.io/packages/keptn/keptn-integrations/zendesk-integration) or try the [Keptn Lifecycle Toolkit](https://lifecycle.keptn.sh)
### Any questions should be directed to [keptn slack](https://slack.keptn.sh/)

![image](https://user-images.githubusercontent.com/13639658/113554176-3e28d680-963c-11eb-8851-a49aeb66aa7a.png)

This service provides an integration between Keptn and Zendesk. Quality Evaluations and Remediation Actions will automatically create fully tagged tickets in Zendesk.

By default, this service [listens](https://github.com/Dynatrace-Adam-Gardner/keptn-zendesk-service/blob/3bf81ca9ca0f7376ef8beab32edfa48ff1c8ea85/deploy/service.yaml#L140) for the following events:
- `sh.keptn.event.evaluation.finished`
- `sh.keptn.event.remediation.finished`

## Gather Required Details

- Your zendesk base URL is `https://***.zendesk.com` WITHOUT trailing slash
- You need to generate an API token. Do so at `https://***.zendesk.com/agent/admin/api/settings`
- Your end user email is the user who will create tickets (The customer, NOT the agent). Add a new user here: `https://***.zendesk.com/agent/users/new`

![image](https://user-images.githubusercontent.com/13639658/113497995-4ace0180-954c-11eb-9cbd-70984a2f34e5.png)


## Create Secret
```
kubectl -n keptn create secret generic zendesk-details \
--from-literal="zendesk-base-url=***" \
--from-literal="zendesk-end-user-email=***" \
--from-literal="zendesk-api-token=***" \
--from-literal="zendesk-create-ticket-for-problems=true" \
--from-literal="zendesk-create-ticket-for-evaluations=true"
```

Expected output:
```
secret/zendesk-details created
```

## Deployment
Images will be tagged corresponding to their supported Keptn version. Modify the image tag to match the version of Keptn you're running. For example, tag `0.8.0` is designed to work with Keptn `0.8.0`.

Customise `deploy/service.yaml` then apply:
```
kubectl apply -f deploy/service.yaml
```
## Debugging
Get Pod:

```
kubectl get pods -n keptn -l app=zendesk-service
```

Logs can be shown using `kubectl logs` and specifying the `zendesk-service` container:

```
kubectl logs -n keptn -l app=zendesk-service -c zendesk-service
```

## Uninstall

```
kubectl delete -f deploy/service.yaml
kubectl delete secret -n keptn zendesk-service
```

## Compatability Matrix

| Keptn Version                                                      | Zendesk Service Tag                                                            |
|--------------------------------------------------------------------|--------------------------------------------------------------------------------|
|    [0.8.2](https://github.com/keptn/keptn/releases/tag/0.8.2)      |  [0.8.2](https://hub.docker.com/r/adamgardnerdt/keptn-zendesk-service/tags)    |
|    [0.8.1](https://github.com/keptn/keptn/releases/tag/0.8.1)      |  [0.8.1](https://hub.docker.com/r/adamgardnerdt/keptn-zendesk-service/tags)    |
|    [0.8.0](https://github.com/keptn/keptn/releases/tag/0.8.0)      |  [0.8.0](https://hub.docker.com/r/adamgardnerdt/keptn-zendesk-service/tags)    |


