# Integrate Hashicorp Vault with Microservices in Kubernetes

Vault has a concept called `Database Secret Engine` which supports widely used DBs. The database secrets engine is highly configurable & generates database credentials dynamically based on configured `roles` for a given `TTL`.

<p class="has-text-centered">
  <img src="integration/static/full.png" alt="full" style="border: none">
</p>

---

Things we're concerned about:
- Deploying Vault on Kubernetes
- Microservice authentication to Vault
- Configuring authentication mechanism
- Configuring Database Secret Engine for Dynamic creds generation
- Consuming those generated creds into K8s workloads, our microservices
- Providing fine grained privileges to Microservices
- Periodically rotating credentials

---

In terms of available options on consuming Vault generated creds into microservices:
- Using Vault Agent Sidecar Injector
  - Sidecar Injection
  - Deployment change not required
  - Secrets rotation supported
  - Requires Helm

- Using Vault Agent Init Container (Preferable)
  - Uses Init Container
  - Deployment change required (only add some annotations)
  - Supports templating from secrets
  - Lightweight
  - Doesn't require Helm
  - Manual/Automated intervention for Secret Rotation
  
- Using Secrets Store CSI Driver
  - Injects Daemonset to each K8s nodes
  - Deploy change required
  - Secrets rotation not supported
  - Secrets templating not supported
  - Requires Helm, CSI Driver, Vault CSI provider

- Re-writing Microservices to leverage on Vault APIs (Services must be Vault aware)

---

We're going to deploying Vault with `Raft` HA Storage Backend with a self written auto unsealer & consuming DB creds using `Vault agent init container`. In production some Vault provider Auto Unseal mechanism should be used. 

*Note:* The directory `integration` contains all the necessary manifest files for this demo, *which is not part of the microservice itself.*

---

We'll create separate `namespaces` for the workloads:

```bash
NAME                 STATUS   AGE
app                  Active   5d2h 
db                   Active   5d3h 
monitoring           Active   3d3h
vault                Active   5d5h
```

# Deploy Vault on K8s - (Persona - Vault Admin)

```bash
# deploy vault on vault namespace

$ kubectl apply -f integration/vault/vault.yaml

serviceaccount/vault-agent-injector created
serviceaccount/vault created
configmap/vault-config created
clusterrole.rbac.authorization.k8s.io/vault-agent-injector-clusterrole created
clusterrolebinding.rbac.authorization.k8s.io/vault-agent-injector-binding created
clusterrolebinding.rbac.authorization.k8s.io/vault-server-binding created
role.rbac.authorization.k8s.io/vault-discovery-role created
rolebinding.rbac.authorization.k8s.io/vault-discovery-rolebinding created
service/vault-agent-injector-svc created
service/vault-active created
service/vault-standby created
service/vault-internal created
service/vault created
deployment.apps/vault-agent-injector created
statefulset.apps/vault created
poddisruptionbudget.policy/vault created
mutatingwebhookconfiguration.admissionregistration.k8s.io/vault-agent-injector-cfg created

# additionally we'll create the auto-unsealer to initialize & unseal the vault cluster
# note that: this is not required in case we use gcpckms auto unseal in gke

$ kubectl apply -f integration/vault/unsealer.yaml
```

Once the Vault resources are up & running:

```bash
$ kubectl get all -n vault

NAME                                       READY   STATUS    RESTARTS   AGE
pod/vault-0                                1/1     Running   0          3m6s
pod/vault-1                                1/1     Running   0          3m6s
pod/vault-2                                1/1     Running   0          3m6s
pod/vault-agent-injector-fbf44fc45-qcth7   1/1     Running   0          3m6s
pod/vault-unsealer-7b6998cdb5-6wz5g        1/1     Running   0          80s

NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/vault                      ClusterIP   10.96.32.225    <none>        8200/TCP,8201/TCP   3m6s
service/vault-active               ClusterIP   10.96.116.55    <none>        8200/TCP,8201/TCP   3m6s
service/vault-agent-injector-svc   ClusterIP   10.96.160.167   <none>        443/TCP             3m6s
service/vault-internal             ClusterIP   None            <none>        8200/TCP,8201/TCP   3m6s
service/vault-standby              ClusterIP   10.96.239.109   <none>        8200/TCP,8201/TCP   3m6s

NAME                                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/vault-agent-injector   1/1     1            1           3m6s
deployment.apps/vault-unsealer         1/1     1            1           80s

NAME                                             DESIRED   CURRENT   READY   AGE
replicaset.apps/vault-agent-injector-fbf44fc45   1         1         1       3m6s
replicaset.apps/vault-unsealer-7b6998cdb5        1         1         1       80s

NAME                     READY   AGE
statefulset.apps/vault   3/3     3m6s
```

---

# Deploy DB(Postgres) - (Persona - DB Admin)

```bash
# deploy DB on db namespace 

$ kubectl apply -f integration/postgres/postgres.yaml

service/postgres created
deployment.apps/postgres created
```

Once the postgres is up & running:

```bash
$ kubectl get all -n db

NAME                            READY   STATUS    RESTARTS   AGE
pod/postgres-56c58c445d-tx85b   1/1     Running   0          87s

NAME               TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/postgres   ClusterIP   10.96.54.28   <none>        5432/TCP   87s

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/postgres   1/1     1            1           87s

NAME                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/postgres-56c58c445d   1         1         1       87s

```

At this point, let's create a `table` & write some dummy data in our `DB` (check the pg DB commands at the bottom for reference)

---

# Configure Authentication in Vault for App (Persona - Vault Admin)

## Enable Kubernets Authentication

```bash
# enable kubernetes auth 

$ vault auth enable kubernetes

Success! Enabled kubernetes auth method at: kubernetes/


# configure kubernetes authentication

$ vault write auth/kubernetes/config \
    token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
    kubernetes_host=https://${KUBERNETES_PORT_443_TCP_ADDR}:443 \
    kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

Success! Data written to: auth/kubernetes/config

# Create Policy/Permission for the Microservice to use
# We'll provide fine grained permissions, only those required by our services

$ vault policy write app-policy integration/vault/app-policy.hcl

Success! Uploaded policy: app-policy

# Create a specifil Role for the Microservice 
# This will bind the service account name, namespace with the Policy created above

$ vault write auth/kubernetes/role/app-creds-reader-role \
    bound_service_account_names=app-auth \
    bound_service_account_namespaces=app \
    policies=app-policy \
    token_ttl=24h

Success! Data written to: auth/kubernetes/role/app-creds-reader-role
```

At this point, our `app` is ready to be authenticated to `Vault`. We're going to create the `service account` mentioned here during app deployment.

<p class="has-text-centered">
  <img src="integration/static/auth.png" alt="auth" style="border: none">
</p>

---

# Setup Database Secret Engine (Persona - Vault Admin)

```bash
# Enable the database secret engine 

$ vault secrets enable database

Success! Enabled the database secrets engine at: database/


# configure vault with plugin & connection information
# provided username, password must be privileged enough to create roles in DB
# We'll rotate this username, password instantly 

$ vault write database/config/postgres \
    plugin_name="postgresql-database-plugin" \
    allowed_roles="db-reader-role" \
    connection_url="postgresql://{{username}}:{{password}}@postgres.db.svc:5432/postgres?sslmode=disable" \
    username="postgres" \
    password="password"

Success! Data written to: database/config/postgres

# Create role that will be bounded to the generated username, password
# creation statements is DB specific valid statement based on the privileges microservice requires
# To be specific, we're providing only DB select permission to our microservice

$ vault write database/roles/db-reader-role \
    db_name="postgres" \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
    default_ttl="1h" \
    max_ttl="1h"

Success! Data written to: database/roles/db-reader-role

# Rotate the root password, we won't need this anymore
# You can verify by trying to login using the root creds provided to configure, it should fail.

$ vault write -force database/rotate-root/postgres

Success! Data written to: database/rotate-root/postgres

```

At this point, our `DB Secret Engine` is `enabled` & `configured` to generated dynamic creds on demand.


<p class="has-text-centered">
  <img src="integration/static/creds.png" alt="creds" style="border: none">
</p>

---

# Deploy the Microservice (Persona - Service Owner)

We're going to use the Vault `Agent Init container` to inject creds into microservices.
This requires only minimal changes in deployment, lightweight & works with K8s annotations intercepting `MutatingWebhook` configuration. Check out the deploy/app.yaml for more details.

```bash
# deploy the microservice on app namespace

$ kubectl apply -f deploy/app.yaml

serviceaccount/app-auth created
deployment.apps/app created

```

Once the app is up & running, we can check the app log to verify that it's using the Vault generated creds.

```bash
2023/01/30 20:59:01 loading app
2023/01/30 20:59:01 loading db
2023/01/30 20:59:01 username:  v-kubernet-db-reade-zhOd64Scg1xFe2ORWQLj-1675112318
2023/01/30 20:59:01 password:  xREZ6QD28HaJL-x8P8Rb
2023/01/30 20:59:01 HTTP: Listening on port 8080

```

## Verify

The service itself exposes the port `8080` & couple of endpoints to `GET` & `UPDATE` user. To verify that `DB` privilege binding is working, `port-forward` from the app:

`kubectl port-forward -n app app-647984bb8-9dhgf 8080`. Using Postman you can verify that, `GET` requests succeds while the `POST` doesn't, since the privilege we binded allows
our service only to `GET` from `DB`.


### Make a GET request

<p class="has-text-centered">
  <img src="integration/static/get.png" alt="get-success" style="border: none">
</p>

### Make a POST request

<p class="has-text-centered">
  <img src="integration/static/post.png" alt="post-fail" style="border: none">
</p>

---

# Useful Postgres Commands

```
# creating a table called users

CREATE TABLE users
(
id          serial       not null unique,
title       varchar(255) not null,
name varchar(255)
);

# exect into postgres pod, connect using psql, authenticate with username, password

# inserting items into users table

INSERT INTO users(id, title, name)
VALUES (1, 'sre', 'sakib');

INSERT INTO users(id, title, name)
VALUES (2, 'manager', 'hanifa');

response should be in both case:

INSERT 0 1

# select all items from users table
select * from users;

postgres=# select * from users;
 id |  title  |  name  
----+---------+--------
  1 | sre     | sakib
  2 | manager | hanifa
(2 rows)

```

---

# Resources:

- https://developer.hashicorp.com/vault/docs/secrets/databases
- https://developer.hashicorp.com/vault/api-docs/secret/databases
- https://developer.hashicorp.com/vault/docs/platform/k8s/injector
- https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-sidecar
- https://developer.hashicorp.com/vault/docs/platform/k8s/injector/annotations
- https://www.hashicorp.com/blog/refresh-secrets-for-kubernetes-applications-with-vault-agent
- https://developer.hashicorp.com/vault/tutorials/raft/raft-storage
- https://developer.hashicorp.com/vault/tutorials/auto-unseal/autounseal-gcp-kms
- https://developer.hashicorp.com/vault/tutorials/operations/production-hardening

