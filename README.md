# stuttgart-things/dagger

collection of dagger modules.

## MODULES

<details><summary><b>DOCKER</b></summary>

### SCAN IMAGE

```bash
dagger call -m \
github.com/stuttgart-things/dagger/docker@v0.6.0 \
trivy-scan \
--image-ref nginx:latest \
--progress plain
```

### BUILD + PUSH TEMPORARY IMAGE w/o AUTH

```bash
dagger call -m \
github.com/stuttgart-things/dagger/docker@v0.6.0 \
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
./docker build-and-push \
--source tests/docker \
--registry-url ghcr.io \
--repository-name stuttgart-things/sthings-alpine \
--version 1.10 \
--with-registry-username=env:USER \
--with-registry-password=env:PASSWORD \
--progress plain
```

</details>

<details><summary><b>GOLANG</b></summary>

### LINT PROJECT

```bash
dagger call -m \
"github.com/stuttgart-things/dagger/go@v0.2.2" \
lint --src "." --timeout 300s --progress plain
```

### BUILD PROJECT

```bash
dagger call -m \
"github.com/stuttgart-things/dagger/go@v0.4.4" \
binary --src "." --os linux --arch amd64 --goMainFile main.go --binName calc \
export --path=/tmp/go/build/ --progress plain
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