+++
weight = 10
+++

{{< slide id=backup background-color="#FFB3D1" type="slide" transition="zoom" transition-speed="fast" >}}

{{% section %}}

## /WHAT TO BACKUP

<img src="https://miro.medium.com/v2/resize:fit:934/0*xgCTrk6NbQGdv_qF.jpeg" alt="Alt Text" width="700"/>

- Stateful application data e.g. databases using persistent volumes
- Cluster state e.g. etcd backup
- Namespace/resource-level e.g. deployments, services ... (resource definitions)

---

# /EXECUTION

<img src="https://www.zerto.com/wp-content/uploads/2021/03/2021-meme-contest-andy-2-300x296.jpeg" alt="Alt Text" width="400"/>

* on-demand backup/restore
* scheduled backup/restore
* kubernetes custom resources (yaml/git)

---

## /STACKEHOLDER

<img src="https://miro.medium.com/v2/resize:fit:578/1*X5utmik_Ch4hMLyeuy-0LA.png" alt="Alt Text" width="450"/>

- dev: application
- platform: backup-sw
- infrastructure: storage-systems

---

## /USE-CASES

- application/namespace
- pvcs
- volumesnapshots
- cross namespace/cluster restores (rewrites)
- schedules
- configuration as code

---

## /VOLUME-SNAPSHOTS

<img src="https://miro.medium.com/v2/resize:fit:1400/1*dmFkYADFaijhRDa37p1RTg.png" alt="Alt Text" width="800"/>

{{% /section %}}