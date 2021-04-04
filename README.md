# keptn-zendesk-service
Keptn Notification Service for Zendesk

![image](https://user-images.githubusercontent.com/13639658/113499414-33e1dc00-9559-11eb-8973-c73d8ffd8390.png)

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
