# keptn-zendesk-service
Keptn Notification Service for Zendesk

> WORK IN PROGRESS. NOT READY FOR USE YET. CONTACT AUTHOR FOR MORE INFO.

## Gather Required Details

- Your zendesk base URL is `https://***.zendesk.com` WITHOUT trailing slash
- You need to generate an API token. Do so at `https://***.zendesk.com/agent/admin/api/settings`
- Your end user email is the user who will create tickets. Obviously this user must exist first

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
Customise `deploy/service.yaml` then apply:
```
kubectl apply -f deploy/service.yaml
```
