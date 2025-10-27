+++
weight = 40
+++

{{< slide id=comparison background-color="#C740FF" type="slide" transition="zoom" transition-speed="fast" >}}

{{% section %}}

## /COMPARISON

|          | **Velero**                                                                 | **Kasten K10**                                                                                          |
|----------------------------|-----------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| **Vendor / Origin**        | Open source project by VMware                                              | Commercial product by Kasten (acquired by Veeam)                                                        |
| **License**                | Apache 2.0                                                                 | Proprietary (Free for up to 5 nodes)                                                                     |
| **UI Availability**        | CLI-based only                                                             | Web UI, CLI (k10tools), and REST API                                                                     |
| **Backup Targets**         | Object storage (S3, GCS, Azure Blob, etc.)                                 | Object storage + Snapshots + NFS via Kanister                                                            |

---

|        | **Velero**                                                                 | **Kasten K10**                                                                                          |
|----------------------------|-----------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| **Application-Awareness**  | Limited (via plugins or hooks)                                             | Native support for MySQL, PostgreSQL, MongoDB, Kafka, etc. via Kanister                                 |
| **Restore Capabilities**   | Namespaced restores, full cluster restores                                 | Granular restores, cross-namespace, cross-cluster, and even cloud-to-cloud restores                     |
| **Multi-Cluster Support**  | No native support, requires manual context switching                       | Built-in centralized multi-cluster management                                                            |
| **Policy Engine**          | Basic with Velero Schedules                                                | Advanced, policy-driven automation with SLAs                                                             |

---

| **Feature/Capability**     | **Velero**                                                                 | **Kasten K10**                                                                                          |
|----------------------------|-----------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| **Security**               | • RBAC support<br>• No built-in encryption                                | • RBAC-aware<br>• Encryption (at rest & in transit)<br>• Vault/KMS integration                          |
| **Extensibility**          | • Custom hooks via scripts                                               | • Kanister blueprints<br>• App-specific automation                                                      |
| **Disaster Recovery**      | • Manual cluster-to-cluster restore                                      | • Automated workflows<br>• Cross-cloud/region migration                                                 |

---

{{% /section %}}
