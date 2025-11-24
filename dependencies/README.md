# Dependencies Dagger Module

## RENOVATE

```bash
dagger call -m dependencies renovate-dry-run \
--repo="stuttgart-things/ansible" \
--github-token=env:GITHUB_TOKEN
```

## ANSIBLE

```bash
# CHECK REQUIREMENTS FILE FOR UPDATES
dagger call -m ./dependencies update-ansible-requirements \
--requirements-file ansible/requirements.yaml
```

```bash
# CHECK REQUIREMENTS FILE FOR UPDATES
dagger call -m ./dependencies apply-ansible-updates \
--requirements-file ansible/requirements.yaml \
export --path /tmp/requirements.yaml
```

```bash
# UPDATES ONLY SPECIFIC COLLECTIONS YOU SPECIFY
dagger call -m ./dependencies update-ansible-requirements-and-apply \
--requirements-file /home/sthings/projects/stuttgart-things/ansible/requirements.yaml \
--collections-to-update "community.general,kubernetes.core" \
export --path /tmp/requirements.yaml
```
