# resource-quota-check

The `resource-quota-check` inspects resource quotas across namespaces and reports when CPU or memory usage reaches the configured threshold.

## Configuration

Set these environment variables in the `HealthCheck` spec:

- `BLACKLIST` (optional): comma-separated namespaces to exclude.
- `WHITELIST` (optional): comma-separated namespaces to include.
- `THRESHOLD` (optional): usage threshold as a float (for example, `0.9` for 90%). Defaults to `0.9`.
- `DEBUG` (optional): set to `true` to enable debug logging.
- `KUBECONFIG` (optional): explicit kubeconfig path for local development.

The check timeout defaults to 5 minutes but is overridden by the Kuberhealthy run deadline when available.

## Build

- `just build` builds the container image locally.
- `just test` runs unit tests.
- `just binary` builds the binary in `bin/`.

## Example HealthCheck

Apply the example below or the provided `healthcheck.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: resource-quota-check
  namespace: kuberhealthy
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: resource-quota-check
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["resourcequotas"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: resource-quota-check
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: resource-quota-check
subjects:
  - kind: ServiceAccount
    name: resource-quota-check
    namespace: kuberhealthy
---
apiVersion: kuberhealthy.github.io/v2
kind: HealthCheck
metadata:
  name: resource-quota-check
  namespace: kuberhealthy
spec:
  runInterval: 5m
  timeout: 5m
  podSpec:
    spec:
      serviceAccountName: resource-quota-check
      containers:
        - name: resource-quota-check
          image: kuberhealthy/resource-quota-check:sha-<short-sha>
          imagePullPolicy: IfNotPresent
          env:
            - name: THRESHOLD
              value: "0.9"
      restartPolicy: Never
```
