# Stuttgart-Things Dagger Modules

A comprehensive collection of Dagger modules for infrastructure automation, container management, security scanning, and DevOps workflows.

## 🚀 Available Modules

| Module | Purpose | Key Features |
|--------|---------|--------------|
| [**Ansible**](./ansible/README.md) | Automation & Configuration | Playbook execution, collection building, GitHub releases |
| [**Go**](./go/README.md) | Go Development | Linting, building, Ko containers, security scanning |
| [**Helm**](./helm/README.md) | Kubernetes Package Management | Chart operations, Helmfile, validation, registry publishing |
| [**Terraform**](./terraform/README.md) | Infrastructure as Code | Plan/apply automation, Vault integration, state management |
| [**Docker**](./docker/README.md) | Container Management | Building, registry operations, multi-platform support |
| [**Hugo**](./hugo/README.md) | Static Site Generation | Site building, MinIO integration, development server |
| [**Crossplane**](./crossplane/README.md) | Cloud Infrastructure | Package management, OCI publishing, custom packages |
| [**GitLab**](./gitlab/README.md) | GitLab Integration | API operations, merge requests, repository management |
| [**Packer**](./packer/README.md) | Image Building | VM templates, vCenter operations, Vault authentication |
| [**Kyverno**](./kyverno/README.md) | Policy Management | Policy validation, resource compliance testing |
| [**Trivy**](./trivy/README.md) | Security Scanning | Vulnerability scanning, container security, compliance |
| [**SOPS**](./sops/README.md) | Secret Management | Encryption/decryption, key management, K8s secrets |
| [**Release**](./release/README.md) | Release Automation | Semantic versioning, changelog generation, GitHub releases |
| [**Git**](./git/README.md) | Git Operations | Repository management, branching, tagging, remote sync |
| [**Crane**](./crane/README.md) | Registry Operations | Image copying, manifest inspection, multi-arch support |

## 🔧 Quick Start

### Prerequisites
- **Dagger CLI** installed ([installation guide](https://docs.dagger.io/install))
- **Docker** runtime available
- **Git** for repository operations

### Using a Module

```bash
# Example: Lint Go code
dagger call -m go lint --src ./my-go-project

# Example: Build Helm chart
dagger call -m helm build-chart --src ./my-helm-chart

# Example: Scan for vulnerabilities
dagger call -m trivy scan-filesystem --src ./my-project
```

### Module Development

Each module includes comprehensive documentation:
- **Features** and capabilities
- **Prerequisites** and setup
- **Quick Start** examples
- **API Reference** with all available functions
- **Testing** instructions

## 📋 Task Automation

Interactive task selection using [gum](https://github.com/charmbracelet/gum):

```bash
# Run interactive task menu
task

# Available tasks include:
# - Module testing and validation
# - Release management
# - Documentation generation
# - Code quality checks
```

## 🔍 Detailed Examples

For comprehensive usage examples and advanced configurations, see the individual module documentation linked above. Each module provides:

- Step-by-step tutorials
- Real-world use cases
- Integration patterns
- Best practices
- Troubleshooting guides

## 📖 Legacy Examples

<details><summary><b>RELEASE</b></summary>

```bash
# SEMANTIC RELEASE
dagger call -m release semantic \
--src ~/projects/k2n/ \
--token env:GITHUB_TOKEN \
--progress plain -vv \
```

```bash
# DELETE EXISTING TAG
dagger call -m release delete-tag \
--src ~/projects/k2n/ \
--release-tag v1.0.0 \
--git-config file://~/.gitconfig \
-vv --progress plain \
export --path ~/projects/k2n/
```

</details>

<details><summary><b>TRIVY</b></summary>

```bash
# FILESYSTEM SCAN LOCAL
dagger call -m trivy scan-filesystem \
--src /home/sthings/projects/stuttgart-things \
--progress plain -vv \
export --path=/tmp/trivy-fs.json
cat /tmp/trivy-fs.json
```

```bash
# FILESYSTEM SCAN FROM REMOTE GIT REPO
dagger call -m trivy scan-filesystem \
--src git://github.com/stuttgart-things/ansible.git \
--progress plain -vv \
export --path=/tmp/trivy-fs.json
cat /tmp/trivy-fs.json
```

```bash
# IMAGE SCAN (w/ REG LOGIN)
export REG_USER=""
export REG_PW=""

dagger call -m trivy scan-image \
--image-ref nginx:latest \
--registry-user=env:REG_USER \
--registry-password=env:REG_PW \
--progress plain -vv \
export --path=/tmp/image-nginx.json
```

</details>

<details><summary><b>GIT</b></summary>

```bash
dagger call -m git clone-git-hub \
--repository stuttgart-things/stuttgart-things \
--ref main \
--token env:GITHUB_TOKEN \
-vv --progress plain \
export --path=/tmp/repo
```

</details>

<details><summary><b>SOPS</b></summary>

```bash
# ENCRYPT SOPS SECRET
export AGE=age1g438n4l..

dagger call -m sops \
encrypt --age-key env:AGE \
--sops-config ~/.sops.yaml \
--plaintext-file tests/sops/tfvars.json \
--file-extension json \
export --path=/tmp/tfvars.enc.json
```

```bash
# DECRYPT SOPS SECRET
dagger call -m sops \
decrypt-sops \
--sops-key=env:SOPS_AGE_KEY \
--encrypted-file /tmp/tfvars.enc.json
```

</details>

<details><summary><b>TERRAFORM</b></summary>

```bash
# RUN TERRAFORM APPLY AND EXPORTS DIR w/ STATE
dagger call -m terraform \
execute \
--terraform-dir tests/terraform \
--variables "name=patrick,food=kaiserschmarrn" \
--operation apply \
-vv --progress plain \
export --path=~/tmp/dagger/tests/terraform/
```

```bash
# RUN TERRAFORM APPLY AND EXPORTS DIR w/ STATE + MOUNT SECRETS FILE
# SECRETS FILE MUST EXIST UNECRYPTED ON FS
dagger call -m terraform \
execute \
--terraform-dir tests/terraform \
--variables "name=patrick" \ # HIGHEST VAR PRIORITY
--secret-json-variables file://tests/terraform/terraform.tfvars.json \
--operation apply \
-vv --progress plain \
export --path=~/tmp/dagger/tests/terraform/
```

```bash
# RUN TERRAFORM APPLY w/ VAULT LOOKUPS
dagger call -m terraform \
execute \
--terraform-dir /home/sthings/projects/blueprints/tests/vmtemplate/tfvaulttest \
--vault-secret-id env:VAULT_SECRET_ID \
--vault-role-id env:VAULT_ROLE_ID \
--variables "vault_addr=https://vault.example.com:8200" \
--operation apply \
-vv --progress plain \
export --path=~/tmp/dagger/tests/terraform/
```

```bash
# RUN TERRAFORM OUTPUT
dagger call -m terraform \
output \
--terraform-dir ~/projects/terraform/vms/dagger/ \
-vv --progress plain
```

</details>

<details><summary><b>HUGO</b></summary>

```bash
# INIT HUGO FOLDER STRUCTURE (INCLUDING THEME)
dagger call -m hugo \
init-site \
--name test \
--config tests/hugo/hugo.toml \
--content tests/hugo/content \
export --path /tmp/hugo/test

# BUILD AND SERVE
dagger call -m hugo serve \
--config tests/hugo/hugo.toml \
--content tests/hugo/content \
--port 4144 \
up --progress plain

# BUILD + EXPORT STATIC CONTENT (INCLUDING THEME)
dagger call -m hugo \
build-and-export \
--name blog \
--config tests/hugo/hugo.toml \
--content tests/hugo/content \
export --path /tmp/blog/static

# JUST SYNC MINIO BUCKET
export MINIO_USER=""
export MINIO_PASSWORD=""

dagger call -m hugo \
sync-minio-bucket \
--endpoint https://artifacts.automation.example.com \
--bucket-name images \
--insecure=true \
--access-key=env:MINIO_USER \
--secret-key=env:MINIO_PASSWORD \
--alias-name artifacts \
export --path /tmp/images

# BUILD + EXPORT STATIC CONTENT (INCLUDING THEME+BUCKET)
export MINIO_USER=""
export MINIO_PASSWORD=""

dagger call -m hugo \
build-sync-export \
--name blog \
--config tests/hugo/hugo.toml \
--content tests/hugo/content/ \
--endpoint https://artifacts.automation.sthings-vsphere.example.com \
--bucket-name idp \
--insecure=true \
--access-key=env:MINIO_USER \
--secret-key=env:MINIO_PASSWORD \
--alias-name artifacts \
-vv export \
--path=/tmp/bucket
```

# SERVE EXPORTED STATIC CONTENT

```bash
# WORKAROUND FOR NOW
chmod -R o+rX /tmp/blog/static
docker run --rm -p 8080:80 \
-v "/tmp/blog/static:/usr/share/nginx/html" nginx
```

</details>

<details><summary><b>PACKER</b></summary>

```bash
dagger call -m packer bake \
--local-dir "." \
--build-path tests/packer/hello/hello.pkr.hcl \
--progress plain -vv
```

```bash
# w/ VAULT AUTH (PACKER ONLY WORKS WITH VAULT TOKEN, FOR ANSIBLE WE'RE USING APPROLE AUTH)
export VAULT_ROLE_ID=<>
export VAULT_TOKEN=<>
export VAULT_SECRET_ID<>

dagger call -m packer bake \
--local-dir "/home/sthings/projects/stuttgart-things/packer/builds/ubuntu24-labda-vsphere/" \
--build-path ubuntu24-base-os.pkr.hc\
--vault-addr https://vault-vsphere.example.com:8200 \
--vault-role-id env:VAULT_ROLE_ID \
--vault-token env:VAULT_TOKEN \
--vault-secret-id env:VAULT_SECRET_ID \
--progress plain -vv
```

```bash
# MOVE VM TEMPLATE
export VCENTER_FQDN=https://10.100.135.50/sdk
export VCENTER_USER=<>
export VCENTER_PASSWORD<>

dagger call -m packer vcenteroperation \
--operation move \
--vcenter env:VCENTER_FQDN \
--username env:VCENTER_USER \
--password env:VCENTER_PASSWORD \
--source /Datacenter/vm/stuttgart-things/rancher-things/sthings-app-4 \
--target /Datacenter/vm/stuttgart-things/testing/ \
--progress plain -vv
```

```bash
# RENAME VM TEMPLATE
export VCENTER_FQDN=https://10.100.135.50/sdk
export VCENTER_USER=<>
export VCENTER_PASSWORD<>

dagger call -m packer vcenteroperation \
--operation rename \
--vcenter env:VCENTER_FQDN \
--username env:VCENTER_USER \
--password env:VCENTER_PASSWORD \
--source /Datacenter/vm/stuttgart-things/vm-templates/u22-rke2-ipi  \
--target u22-rke2-old \
--progress plain -vv
```

</details>

<details><summary><b>KYVERNO</b></summary>

```bash
# VALIDATE RESOURCES AGAINST POLICIES
dagger call -m kyverno validate \
--policy tests/kyverno/policies/ \
--resource tests/kyverno/resource-good/ \
--progress plain
```

```bash
# OUTPUT KYVERNO VERSION
dagger call -m kyverno version \
--progress plain
```

</details>

<details><summary><b>GITLAB</b></summary>

```bash
# GET PROJECT ID BY PROJECT NAME
dagger call -m gitlab get-project-id \
--token env:GITLAB_TOKEN \
--server gitlab.com \
--project-name "docs" \
--group-path "Lab/stuttgart-things/idp"
```

```bash
# GET MERGE REQUEST ID BY PROJECT ID
dagger call -m gitlab list-merge-requests \
--token env:GITLAB_TOKEN \
--server gitlab.com \
--project-id 14160 \
--progress plain
```

```bash
# GET MERGE REQUEST ID BY PROJECT ID
dagger call -m gitlab get-merge-request-id \
--token env:GITLAB_TOKEN \
--server gitlab.com \
--project-id 14466 \
--merge-request-title "RFC- -" \
--progress plain
```

```bash
# LIST ALL CHANGES FROM MR INTO (USUALY) MAIN
dagger call -m gitlab list-merge-request-changes \ --token env:GITLAB_TOKEN \
--server gitlab.com \
--project-id="14466" \
--merge-request-id="1" \
--progress plain
```

```bash
# LIST ALL CHANGES FROM MR INTO (USUALY) MAIN
dagger call -m gitlab clone \
--repo-url https://gitlab.com/Lab/stuttgart-things/idp/resource-engines.git
--token env:GITLAB_TOKEN \
--branch=main \
#export --path /tmp/repo \ # IF YOU WANT TO EXPORT TO LOCAL FS
--progress plain
```

```bash
# PRINT ALL FILES CHANGED BY A MR
dagger call -m gitlab print-merge-request-file-changes \
--repo-url https://gitlab.com/Lab/stuttgart-things/idp/resource-engines.git \
--server gitlab.com \
--token env:GITLAB_TOKEN \
--merge-request-id="1" \
--project-id="14466" \
--branch "RFC-_" \
--progress plain
```

```bash
# LIST ALL PROJECTS IN A GROUP
dagger call -m gitlab list-projects \
--server gitlab.com \
--token env:GITLAB_TOKEN \
--group-path "Lab%2Fstuttgart-things"
--progress plain
```

```bash
# PRINT ALL FILES CHANGED BY A MR
dagger call -m gitlab update-merge-request-state \
--server gitlab.com \
--token env:GITLAB_TOKEN \
--merge-request-id="1" \
--project-id="14466" \
--action merge \ # or 'close'
--progress plain
```

</details>

<details><summary><b>CRANE</b></summary>

```bash
# REG AUTH FOR SOURCE AND TARGET REG
dagger call -m crane copy \
--sourceRegistry ghcr.io \
--sourceUsername patrick-hermann-sva \
--sourcePassword env:GITHUB_TOKEN \
--targetRegistry harbor.example.com \
--targetUsername admin \
--targetPassword env:HARBOR \
--platform linux/amd64 \
--insecure=true \
--source ghcr.io/stuttgart-things/backstage:2025-04-22 \
--target harbor.example.com/test/backstage:2025-04-22 \
--progress plain
```

```bash
# REG AUTH FOR TARGET REG ONLY
dagger call -m crane copy \
--targetUsername admin \
--targetPassword env:HARBOR \
--source redis:latest \
--target harbor.example.com/test/redis:2025-04-22 \
--targetRegistry harbor.example.com \
--insecure=true \
--platform linux/amd64 \
--progress plain
```

</details>

<details><summary><b>DOCKER</b></summary>

### LINT

```bash
dagger call -m docker \
lint \
--src tests/docker \
-vv --progress plain
```

### BUILD

```bash
dagger call -m docker \
build \
--src tests/docker \
-vv --progress plain
```

### BUILD + PUSH TEMPORARY IMAGE w/o AUTH

```bash
dagger call -m docker \
build-and-push \
--source tests/docker \
--repository-name stuttgart-things/test \
--registry-url ttl.sh \
--tag 1.2.3 \
-vv --progress plain
```

### BUILD + PUSH IMAGE w/ AUTH

```bash
dagger call -m docker \
build-and-push \
--source tests/docker \
--repository-name stuttgart-things/test \
--registry-url ghcr.io \
--tag 1.2.3 \
--registry-username=env:GITHUB_USER \
--registry-password=env:GITHUB_TOKEN \
-vv --progress plain
```

</details>

<details><summary><b>GO</b></summary>

### LINT PROJECT

```bash
dagger call -m go \
lint --src "." \
--timeout 300s \
--progress plain \
-vv
```

### BUILD PROJECT

```bash
dagger call -m go \
build-binary \
--src "." \
--os linux \
--arch amd64 \
--ldflags "cmd.version=1.278910; cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
--package-name github.com/stuttgart-things/k2n \
--go-main-file main.go \
--bin-name k2 \
--go-version 1.24.4 \
export --path=/tmp/go/build/ \
--progress plain \
-vv
```

### KO BUILD

```bash
# BUILD JUST LOCAL
dagger call -m go ko-build \
--src tests/go/calculator/ \
--push="false" \
--progress plain -vv
```

```bash
# BUILD + PUSH
dagger call -m go ko-build \
--src tests/go/calculator/ \
--token=env:GITHUB_TOKEN \
--repo ghcr.io/stuttgart-things/machineshop \
--progress plain -vv
```

</details>

<details><summary><b>ANSIBLE</b></summary>

### EXECUTE ANSIBLE

```bash
dagger call -m ansible execute \
--src . \
--playbooks tests/ansible/hello.yaml,tests/ansible/hello2.yaml \
-vv --progress plain
```

```bash
dagger call -m ansible execute \
--requirements tests/ansible/requirements.yaml \
--src . \
--playbooks tests/ansible/hello.yaml,tests/ansible/hello2.yaml \
-vv --progress plain
```

```bash
export SSH_USER=sthings
export SSH_PASSWORD=<REPLACEME>

dagger call -m ansible execute \
--requirements tests/ansible/requirements.yaml \
--src . \
--playbooks tests/ansible/hello.yaml,sthings.baseos.setup \
--inventory /home/sthings/projects/terraform/vms/sthings-runner/rke2.ini \
--ssh-user=env:SSH_USER \
--ssh-password=env:SSH_PASSWORD \
--parameters "send_to_homerun=false" \
-vv --progress plain
```

the idea of this module is to create versioned collection artifcat 'on the fly' -
this module can work with a file structure like this:

### CREATE A COLLECTION PACKAGE

```bash
dagger call --progress plain -m ansible run-collection-build-pipeline \
--src ansible/collections/baseos \
--progress plain \
export --path=/tmp/ansible/output/
```

### BUILD A GITHUB RELEASE FROM FILES

```bash
dagger call --progress plain -m ansible github-release \
--token=env:GITHUB_TOKEN \
--group stuttgart-things \
--repo dagger  \
--files "tests/test-values.yaml,tests/registry/README.md" \
--notes "test" \
--tag 09.1.6 \
--title hello
```

</details>

<details><summary><b>HELM</b></summary>

```bash
# LINT
dagger call -m helm \
lint \
--src tests/helm/test-chart \
-vv --progress plain
```

```bash
# RENDER A CHART w/ VALUES
dagger call -m helm \
render \
--src tests/helm/test-chart \
--valuesFile tests/helm/test-values.yaml \
-vv --progress plain
```

```bash
# PACKAGE + EXPORT CHART AS TGZ
dagger call -m helm \
package \
--src tests/helm/test-chart \
-vv --progress plain \
export --path=/tmp/chart.tgz
```

```bash
# PUSH CHART TO REGISTRY
dagger call -m helm \
push \
--src tests/helm/test-chart \
--registry ghcr.io \
--repository stuttgart-things \
--username patrick-hermann-sva \
--password env:GITHUB_TOKEN \
-vv --progress plain
```

```bash
# RENDER HELMFILE (w/ REG AUTH)
dagger call -m helm \
render-helmfile \
--src tests/helm/ \
--registry-secret file://~/.docker/config.json
```

```bash
# APPLY HELMFILE (w/ KUBECONFIG)
dagger call -m helm \
helmfile-operation \
--src tests/helm/ \
--kube-config file://~/.kube/labda-sthings-infra \
-vv --progress plain
```

```bash
# DESTROY RELEASES w/ HELMFILE (w/ KUBECONFIG DOWNLOADED FROM VAULT)
dagger call -m helm \
helmfile-operation \
--operation destroy \
--src tests/helm/ \
--vault-url env:VAULT_ADDR \
--vault-secret-id env:VAULT_SECRET_ID \
--vault-app-role-id env:VAULT_ROLE_ID \
--secretPathKubeconfig kubeconfigs/test2/kubeconfig \
-vv --progress plain
```

```bash
# DESTROY HELMFILE (w/ KUBECONFIG)
dagger call -m helm \
helmfile-operation \
--operation destroy \
--src tests/helm/ \
--kube-config file://~/.kube/labda-sthings-infra \
-vv --progress plain
```

```bash
# MANIFEST VALIDATION w/ POLARIS
dagger call -m helm \
validate-chart \
--severity danger \
--src tests/helm/test-chart/ \
-vv --progress plain \
export --path=/tmp/polaris.json
```

</details>

## DEV

<details><summary>ALL TASKS</summary>

```bash
task: Available tasks for this project:
* branch:                Create branch from main
* check:                 Run pre-commit hooks
* commit:                Commit + push code into branch
* create:                Create new dagger module
* do:                    Select a task to run
* pr:                    Create pull request into main
* release:               push new version
* switch-local:          Switch to local branch
* switch-remote:         Switch to remote branch
* test:                  Select test to run
* test-ansible:          Test ansible functions
* test-crossplane:       Test crossplame functions
* test-docker:           Test docker module
* test-gitlab:           Test gitlab functions
* test-go:               Test go functions
* test-helm:             Test helm functions
* test-hugo:             Test hugo
* test-packer:           Test packer functions
* test-terraform:        Test terraform functions
```

</details>

<details><summary><b>CREATE NEW MODULE</b></summary>

```bash
# EXAMPLE MODULE
MODULE=crossplane task create
```

</details>

## TESTING

<details><summary><b>.env FILE</b></summary>

```bash
cat <<EOF > .env
gitlab_server="#TOBESET"
gitlab_project=docs # example
gitlab_group="Lab/stuttgart-things/idp" # example
gitlab_group_escaped="Lab%2Fstuttgart-things%2Fidp" # example
EOF
```

</details>

## DAGGER

<details><summary><b>LIST FUNCTIONS</b></summary>

```bash
MODULE=golang #example
dagger functions -m ${MODULE}/
```

</details>

<details><summary><b>CREATE NEW FUNCTION</b></summary>

```bash
MODULE=example #example
dagger init --sdk=go --source=./${MODULE} --name=${MODULE}
```

</details>

<details><summary><b>INSTAL EXTERNAL DAGGER MODULE</b></summary>

```bash
dagger install github.com/purpleclay/daggerverse/golang@v0.5.0


https://github.com/disaster37/dagger-library-go@v0.0.24
```

</details>

<details><summary><b>UPDATE MODULE DEPS</b></summary>

```bash
cd docker  # your module directory

# Remove the cached module files
rm -rf dagger.gen.go go.sum

# Update the go.mod dependency
go get dagger.io/dagger@v0.19.2
go mod tidy

# Regenerate the Dagger SDK files
dagger develop
```

</details>

<details><summary><b>CALL FUNCTION FROM LOCAL</b></summary>

```bash
MODULE=example #example
dagger functions -m ${MODULE}
```

```bash
MODULE=helm #example
dagger call -m ./${MODULE} \
lint --source tests/test-chart/ \
--progress plain
```

</details>

<details><summary><b>CALL FUNCTION FROM GIT</b></summary>

```bash
MODULE=golang #example
dagger call -m github.com/stuttgart-things/dagger/${MODULE} build  \
--progress plain \
--src ./ \
export --path build
```

</details>

## LICENSE

<details><summary><b>APACHE 2.0</b></summary>

Copyright 2023 patrick hermann.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

</details>

```yaml
Author Information
------------------
Patrick Hermann, stuttgart-things 11/2024
```
