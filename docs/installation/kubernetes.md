# Install on Kubernetes

A [Helm](https://helm.sh) chart is available to install QingyuGist on a Kubernetes cluster.
Check the [Helm documentation](https://helm.sh/docs/) for more information on how to use Helm.

A non-customized installation of QingyuGist can be done with:

```bash
helm repo add qingyugist https://helm.qingyugist.io

helm install qingyugist qingyugist/qingyugist
```

Refer to the [QingyuGist chart](https://github.com/LiuShen-Fork/Qingyu-Gist/tree/master/helm/qingyugist) for more information
about the chart and to customize your installation.