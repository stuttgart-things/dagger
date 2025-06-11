# stuttgart-things/dagger

collection of dagger modules

## MODULES

<details><summary><b>TERRAFORM</b></summary>

```bash
# RUN TERRAFORM APPLY
dagger call -m terraform \
execute \
--terraform-dir /home/sthings/projects/terraform/vms/dagger/ \
--operation apply \
-vv --progress plain \
export --path=~/projects/terraform/vms/dagger/
```

```bash
# DECRYPT SOPS SECRETS FILE
# + RUN APPLY
dagger call -m terraform \
execute \
--terraform-dir /home/sthings/projects/terraform/vms/dagger/ \
--operation apply \
--sops-key=env:SOPS_AGE_KEY \
--encrypted-file /tmp/terraform.tfvars.enc.json \
-vv --progress plain \
export --path=~/projects/terraform/vms/dagger/
```

```bash
# RUN TERRAFORM OUTPUT
dagger call -m terraform \
output \
--terraform-dir ~/projects/terraform/vms/dagger/ \
-vv --progress plain
```

```bash
# DECRYPT SOPS SECRET
dagger call -m terraform \
decrypt-sops \
--sops-key=env:SOPS_AGE_KEY \
--encrypted-file /tmp/tfvars.enc.json
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
# GIT
dagger call -m packer build \
--repo-url https://github.com/stuttgart-things/stuttgart-things.git \
--branch "feat/packer-hello" \
--token env:GITHUB_TOKEN \
--build-path packer/builds/hello \
--progress plain -vv
```

```bash
# LOCAL
dagger call -m packer build \
--local-dir "." \
--build-path tests/packer/u24/ubuntu24-base-os.pkr.hcl \
--progress plain -vv

# LOCAL - w/ VAULT AUTH
dagger call -m packer build \
--local-dir "." \
--build-path tests/packer/u24/ubuntu24-base-os.pkr.hcl \
--vault-addr https://vault-vsphere.example.com:8200 \
--vault-role-id 1d42d7e7-8c14-e5f9-801d-b3ecef416616 \
--vault-token env:VAULT_TOKEN \
--vault-secret-id env:VAULT_SECRET_ID \
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

### SCAN IMAGE

```bash
dagger call -m \
github.com/stuttgart-things/dagger/docker@v0.6.2 \
trivy-scan \
--image-ref nginx:latest \
--progress plain
```

### BUILD + PUSH TEMPORARY IMAGE w/o AUTH

```bash
dagger call -m \
github.com/stuttgart-things/dagger/docker@v0.6.2 \
build-and-push \
--source images/sthings-alpine \
--repository-name stuttgart-things/alpine \
--registry-url ttl.sh \
--version 1h \
--progress plain
```

### BUILD + PUSH IMAGE w/ AUTH

```bash
dagger call -m \
github.com/stuttgart-things/dagger/docker@v0.6.2 \
build-and-push \
--source tests/docker \
--registry-url ghcr.io \
--repository-name stuttgart-things/sthings-alpine \
--version 1.10 \
--with-registry-username=env:USER \
--with-registry-password=env:PASSWORD \
--progress plain
```

</details>

<details><summary><b>GO</b></summary>

### LINT PROJECT

```bash
dagger call -m \
"github.com/stuttgart-things/dagger/go@v0.2.2" \
lint --src "." --timeout 300s --progress plain
```

### BUILD PROJECT

```bash
dagger call -m github.com/stuttgart-things/dagger/go@v0.10.2 binary \
--src "." \
--os linux \
--arch amd64 \
--go-main-file main.go \
--bin-name k2 \
--go-version 1.24.2 \
export --path=/tmp/go/build/ \
--progress plain
```

### RUN-WORKFLOW-CONTAINER-STAGE

```bash
dagger call -m \
github.com/stuttgart-things/dagger/go@v0.4.2 \
run-workflow-container-stage --src tests/calculator/ \
--token=env:GITHUB_TOKEN --token-name GITHUB_TOKEN \
--repo ghcr.io/stuttgart-things/dagger \
--ko-version 3979dd70544adde24d336d5b605f4cf6f0ea9479 \
--output /tmp/calc-image.report.json --progress plain
```

</details>

<details><summary><b>ANSIBLE</b></summary>

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

RENDER A CHART w/ VALUES

```bash
# EXAMPLE MODULE
VERSION=v0.0.4
dagger call -m github.com/stuttgart-things/dagger/helm@${VERSION} template --chart ./Service --values this-env.yaml
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
* pr:                    Create pull request into main
* release:               push new version
* switch-local:          Switch to local branch
* switch-remote:         Switch to remote branch
* tasks:                 Select a task to run
* test:                  Select test to run
* test-ansible:          Test ansible modules
* test-crossplane:       Test crossplame modules
* test-go:               Test go modules
* test-helm:             Test helm modules
```

</details>

<details><summary>SELECT TASK</summary>

```bash
task=$(yq e '.tasks | keys' Taskfile.yaml | sed 's/^- //' | gum choose) && task ${task}
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

<details><summary><b>CALL FUNCTION FROM LOCAL</b></summary>

```bash
MODULE=golang #example
dagger call -m github.com/stuttgart-things/dagger/${MODULE} build --progress plain --src ./ export --path build
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