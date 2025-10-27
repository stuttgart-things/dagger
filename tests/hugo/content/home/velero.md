+++
weight = 20
+++

{{< slide id=velero background-color="#D4B9FF" type="slide" transition="zoom" transition-speed="fast" >}}

{{% section %}}

# /VELERO
- open-source kubernetes backup/restore tool maintained by vmware
- complete application backups, disaster recovery, and kubernetes migration

---

## /FEATURES

- Backups: Snapshot cluster resources and persistent volumes.
- Restores: Restore from backups, including specific resources or namespaces.
- Schedules: Automate backups on a schedule (e.g., daily, weekly).
- Volume Snapshots: Integrates with cloud provider snapshot APIs (or Restic for generic PV backup).
- Cluster Migration: Move workloads between clusters or across cloud regions/providers.

---

## /PLATFORMS

| Platform         | Notes                                                         |
| ---------------- | ------------------------------------------------------------- |
| **Azure AKS**    | Uses Azure Disk snapshots and Azure Blob snapshots               |
| **VMware Tanzu** | Full support with vSphere CSI snapshots (vSphere 7+)          |
| **OpenShift**    | Fully supported (Red Hat OpenShift is Kubernetes-compatible)  |
| **Rancher**      | Fully supported                                               |

---

## /On-Prem / Self-Hosted Kubernetes
Works on any bare-metal or on-prem cluster using:

- CSI volume snapshots (if configured)
- Restic for generic volume backup
- Object storage backend (e.g., MinIO, Ceph RGW)

---

## /Is NFS supported?
🔴 Not directly supported as a backup target.
Velero requires an object storage backend (e.g. S3, GCS, Azure Blob). NFS is not an object store.

🟡 Workaround:
You can use NFS with Velero indirectly by Restic integration (Velero can back up volumes to object storage like MinIO, while volumes use NFS).

---

## BackupStorageLocation

```yaml
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: azure-backup
  namespace: velero
spec:
  provider: azure
  objectStorage:
    bucket: velero-container  # The Azure Blob container name
  config:
    resourceGroup: velero-rg
    storageAccount: velerostorage
```

---

## /USAGE CLI

```bash
velero backup create demo-backup \
--include-namespaces demo-ns
```

```bash
velero restore create demo-restore \
--from-backup demo-backup \
--namespace-mappings demo-ns:new-demo-ns
```

```bash
velero schedule create pgsched \
--schedule="0 4 * * *" \
--include-namespaces postgres \
--ttl 72h
```

---

## /USAGE CUSTOM-RESOURCES (YAML)

```yaml
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: cluster-backup
spec:
  includedNamespaces:
    - kube-system
    - my-app
  storageLocation: azure-backup
  snapshotVolumes: true
  ttl: 168h # 7 days
  hooks:
    resources:
      - name: pre-backup-hook
        includedNamespaces:
          - my-app
        includedResources:
          - pods
```

---

## /BACKUP/RESTORE HOOKS

```yaml
podAnnotations:
  backup.velero.io/backup-volumes: backup
  pre.hook.backup.velero.io/timeout: 5m
  pre.hook.restore.velero.io/timeout: 5m
  post.hook.restore.velero.io/command: |
    '["/bin/bash", "-c", "sleep 1m && PGPASSWORD=${POSTGRES_PASSWORD} \
    pg_restore -U postgres -d postgres --clean
    < /scratch/backup.psql"]'
  pre.hook.backup.velero.io/command: |
    '["/bin/bash", "-c", "export PGPASSWORD=${POSTGRES_PASSWORD} \
    && sleep 1m && pg_dump -U postgres -d postgres -F c
    -f /scratch/backup.psql"]'
```

{{% /section %}}
