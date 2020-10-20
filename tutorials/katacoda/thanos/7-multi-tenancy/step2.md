## Reuse More!

What if we can have one set of Queries instead of separate set for each tenant? 

Let's stop fruit and veggies queriers and run single one spanning all the tenant's Prometheus data:

```
docker stop querier-fruit && docker stop querier-veggie
```{{execute}}

```
docker run -d --net=host --rm \
    --name querier-multi \
    quay.io/thanos/thanos:v0.16.0-rc.1 \
    query \
    --http-address 0.0.0.0:29090 \
    --grpc-address 0.0.0.0:29190 \
    --query.replica-label replica \
    --store 127.0.0.1:19190 \
    --store 127.0.0.1:19191 \
    --store 127.0.0.1:19192 && echo "Started Thanos Querier with access to both Veggie's and Fruit's data"
```{{execute}}

Withing short time we should be able to see "Tomato" view [when we open Querier UI](https://[[HOST_SUBDOMAIN]]-29090-[[KATACODA_HOST]].environments.katacoda.com/)

## Tenant Query Isolation

Undoubtedly, the main problem with this setup is that by default **every tenant will see each other data**, similar to what you have in Prometheus,
if single Prometheus scrapes data from multiple teams.

Both Prometeus and Thanos [follow UNIX philosopy](https://github.com/thanos-io/thanos#thanos-philosophy). **One of the principles is to ensure
each component is doing one thing and do it well**. Thanos Querier does not perform any authentication or authorization. 
This is because you probably already have consistent auth mechanism in your organization. So why not composing that with flexible 
flat label pairs identifing the data blocks and each individual series for data isolation? 

### Meet [prom-label-proxy](https://github.com/prometheus-community/prom-label-proxy)

At Red Hat we started [prom-label-proxy](https://github.com/prometheus-community/prom-label-proxy). This allows read tenancy for all the
resources that Prometheus and Thanos currently exposes, by enforcing certain `tenant` label to be used in mathers, as well resulted
data.

It works smoothly with Kubernetes Auth using [kube-rbac-proxy](https://github.com/brancz/kube-rbac-proxy) project

TBD..

