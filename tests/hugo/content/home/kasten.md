+++
weight = 30
+++

{{< slide id=kasten background-color="#00FFA6" type="slide" transition="zoom" transition-speed="fast" >}}

{{% section %}}

# /KASTEN K10
- enterprise-grade platform designed to protect Kubernetes applications, including their data and configuration
- It supports full application backup/restore, disaster recovery (DR), and application migration (across clusters, clouds, and regions)

---
# 🏢 /Where Does It Come From?

- Company: Originally developed by Kasten Inc., a startup founded in 2017 focused on Kubernetes data management.

- Acquisition: Acquired by Veeam in 2020

---

## /FEATURES

| Storage Type       | Examples Supported                               |
|--------------------|--------------------------------------------------|
| Object Storage     | AWS S3, Azure Blob, GCP GCS, MinIO, Ceph RGW     |
| Snapshots          | AWS EBS, Azure Disk, GCE PD, CSI snapshots       |
| NFS via Kanister   | Supported for some workloads with custom blueprints |

---

## /Key Features

| Feature                     | Description                                                                 |
|-----------------------------|-----------------------------------------------------------------------------|
| Application-Aware Backups   | Integrates with databases like MySQL, PostgreSQL, MongoDB, Kafka, etc.     |
| Multi-Cluster Support       | Centralized dashboard and control across multiple Kubernetes clusters      |
| Policy-Driven               | Define backup frequency, retention, and export policies                    |
| Disaster Recovery           | Easy restore to another cluster, region, or cloud                          |
| Secure by Design            | Encrypts data in transit and at rest; integrates with KMS and Vault        |
| UI + CLI + APIs             | Offers a modern web UI, CLI (`k10tools`), and REST APIs                    |

---

## /How Kasten K10 Works


- Installed via Helm into your cluster.
- Scans Kubernetes resources and auto-detects workloads.
- Backs up apps and PVCs to a configured storage location (S3, GCS, etc.).
- Uses Kanister, an open-source framework developed by Kasten, to perform application-consistent operations.
- Enables scheduled backups, restores, export/import, compliance reporting, and more.

---

## /LOCATIONS

<img src="https://private-user-images.githubusercontent.com/47567770/440166371-e804a27f-1b54-43d3-99fa-8a8b279b3a52.png?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NDYzODc5MzYsIm5iZiI6MTc0NjM4NzYzNiwicGF0aCI6Ii80NzU2Nzc3MC80NDAxNjYzNzEtZTgwNGEyN2YtMWI1NC00M2QzLTk5ZmEtOGE4YjI3OWIzYTUyLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNTA1MDQlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjUwNTA0VDE5NDAzNlomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPWNlMGRjODA3N2NkYzA4ZWZkNWI4YjczYzZmYjJhZDkwZmFjNzQzMDViOWUxY2RmOGY3ZmU1ZDk1ZjliYmY2YjgmWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.4LD8AsvgBTt8ZM21-VzVEe4olmRKdhoI4FuCB9b10Vw" alt="Alt Text" width="1000"/>

---

## /APPS/NAMESPACES

<img src="https://private-user-images.githubusercontent.com/47567770/440166726-8b6c9f3a-675f-42be-ba61-525d952f7b2a.png?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NDYzODgwMzAsIm5iZiI6MTc0NjM4NzczMCwicGF0aCI6Ii80NzU2Nzc3MC80NDAxNjY3MjYtOGI2YzlmM2EtNjc1Zi00MmJlLWJhNjEtNTI1ZDk1MmY3YjJhLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNTA1MDQlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjUwNTA0VDE5NDIxMFomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPTRjZTZjMzQ1ZTFkY2I3NmZkN2I5MTBmZjMzZWYxYmJkZGViMjZmMzkwZTVhOWU0NmRkZDExZDM4NzNlYjhhZDImWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.mL1Wy2W1UqpjyOnfhTeL8ZK9CowfGz8bUxuzW2nuYPk" alt="Alt Text" width="1000"/>

---

## /BACKUP

<img src="https://private-user-images.githubusercontent.com/47567770/440166830-4638c77a-0604-4797-a071-c2de0fa8198f.png?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NDYzODc5MzIsIm5iZiI6MTc0NjM4NzYzMiwicGF0aCI6Ii80NzU2Nzc3MC80NDAxNjY4MzAtNDYzOGM3N2EtMDYwNC00Nzk3LWEwNzEtYzJkZTBmYTgxOThmLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNTA1MDQlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjUwNTA0VDE5NDAzMlomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPTI5MjE1MDU0ODdhZjgxZGNlY2Y3NmEyOWMyZDg2MWYwN2FiZTFkNzgwZmJiYzc2M2JjZjc5YWVkYzM5NDBjM2EmWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.ce7vRp08l_c-_9pRy5ST47iSCSV1UUZeZfWVvPR8Lbs" alt="Alt Text" width="450"/>

---

## /BLUEPRINT-BACKUP

```yaml
apiVersion: cr.kanister.io/v1alpha1
kind: Blueprint
metadata:
  name: postgresql-blueprint
actions:
  backup:
    kind: StatefulSet
    phases:
      - func: KubeExec
        name: dumpPostgres
        args:
          namespace: "{{ .StatefulSet.Namespace }}"
          pod: "{{ index .StatefulSet.Pods 0 }}"
          container: postgres
          command:
            - bash
            - -c
            - |
              pg_dump -U $POSTGRES_USER $POSTGRES_DB
              > /mnt/backup/backup.sql
```
---

## /BLUEPRINT-RESTORE

```yaml
apiVersion: cr.kanister.io/v1alpha1
kind: Blueprint
# ...
  restore:
    kind: StatefulSet
    phases:
      - func: KubeExec
        name: restorePostgres
        args:
          namespace: "{{ .StatefulSet.Namespace }}"
          pod: "{{ index .StatefulSet.Pods 0 }}"
          container: postgres
          command:
            - bash
            - -c
            - |
              psql -U $POSTGRES_USER $POSTGRES_DB < /mnt/backup/backup.sql
```

---

## /LICENSE

<img src="https://private-user-images.githubusercontent.com/47567770/440167089-ed5b9ab3-ce32-4fa7-8c7d-216324970e21.png?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NDYzODgyMjYsIm5iZiI6MTc0NjM4NzkyNiwicGF0aCI6Ii80NzU2Nzc3MC80NDAxNjcwODktZWQ1YjlhYjMtY2UzMi00ZmE3LThjN2QtMjE2MzI0OTcwZTIxLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNTA1MDQlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjUwNTA0VDE5NDUyNlomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPTRhZWVhMjhiMWMzZTQ1MmY1OTc0YTFhNjNlNDMzMDFiMDU1YzBlMzkwN2NiZTgzNDQwZjViYjcyZDNhNDQ0OTYmWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.zpYtTN5aFZXbq4EnFDZJgncPLGO9kNUsMjH-9hMEG0A" alt="Alt Text" width="1200"/>

{{% /section %}}