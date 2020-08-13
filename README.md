## Installation

To install this application using Helm run the following commands: 

```bash
helm repo add jorritsalverda https://helm.jorritsalverda.com
kubectl create namespace tp-link-hs110-bigquery-exporter

helm upgrade \
  tp-link-hs110-bigquery-exporter \
  jorritsalverda/tp-link-hs110-bigquery-exporter \
  --install \
  --namespace tp-link-hs110-bigquery-exporter \
  --set config.bqProjectID=your-project-id \
  --set config.bqDataset=your-dataset \
  --set config.bqTable=your-table \
  --set secret.gcpServiceAccountKeyfile='{abc: blabla}' \
  --wait
```

If you later on want to upgrade without specifying all values again you can use

```bash
helm upgrade \
  tp-link-hs110-bigquery-exporter \
  jorritsalverda/tp-link-hs110-bigquery-exporter \
  --install \
  --namespace tp-link-hs110-bigquery-exporter \
  --reuse-values \
  --set cronjob.schedule='*/1 * * * *' \
  --wait
```