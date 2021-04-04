# keptn-zendesk-service
Keptn Notification Service for Zendesk

> WORK IN PROGRESS. NOT READY FOR USE YET. CONTACT AUTHOR FOR MORE INFO.

## Create Secret
```
kubectl -n keptn create secret generic zendesk-details \
--from-literal="zendesk-base-url=***" \
--from-literal="zendesk-end-user-email=***" \
--from-literal="zendesk-api-token=***"
```

Expected output:
```
secret/zendesk-details created
```

## Deployment
Customise `deploy/service.yaml` then apply:
```
kubectl apply -f deploy/service.yaml
```
