# Install on Kubernetes

QingyuGist ships as a single container image
(`ghcr.io/liushen-fork/qingyugist`) with no configuration file: boot-time values
come from environment variables, everything else is set in the admin panel.

A minimal Deployment and Service. The PVC keeps the repositories, database,
logs and keys under `/qingyugist`:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: qingyugist-data
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: 5Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qingyugist
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: qingyugist
  template:
    metadata:
      labels:
        app: qingyugist
    spec:
      securityContext:
        fsGroup: 1000
      containers:
        - name: qingyugist
          image: ghcr.io/liushen-fork/qingyugist:latest
          ports:
            - containerPort: 6157
          env:
            - name: QYG_EXTERNAL_URL
              value: "https://gist.example.com"
            - name: QYG_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: qingyugist-secret
                  key: secret-key
          volumeMounts:
            - name: data
              mountPath: /qingyugist
          readinessProbe:
            httpGet:
              path: /
              port: 6157
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: qingyugist-data
---
apiVersion: v1
kind: Service
metadata:
  name: qingyugist
spec:
  selector:
    app: qingyugist
  ports:
    - port: 80
      targetPort: 6157
```

Create the secret first (the key encrypts sessions and MFA secrets — keep it
stable across restarts and backups):

```shell
kubectl create secret generic qingyugist-secret \
  --from-literal=secret-key="$(openssl rand -hex 32)"
```

Then `kubectl apply -f qingyugist.yaml` and expose the Service through your
Ingress controller. Set `QYG_EXTERNAL_URL` to the public URL so OAuth callbacks
and clone links are correct.

Keep `replicas: 1` and the `Recreate` strategy: the SQLite database and the git
repositories are on a single `ReadWriteOnce` volume. To scale horizontally, point
`QYG_DB_URI` at PostgreSQL or MySQL and give every replica its own volume — git
operations are still filesystem-bound, so a shared RWX volume is required.

The full list of environment variables is in the
[configuration reference](/configuration/configure).
