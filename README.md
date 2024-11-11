# stuttgart-things/dagger
collection of dagger modules

## TASKS

```bash
task: Available tasks for this project:
* branch:          Create branch from main
* commit:          Commit + push code into branch
* pr:              Create pull request into main
* test-go:         Test go modules
* test-helm:       Test helm modules
```

<details><summary><b>CREATE NEW MODULE</b></summary>

```bash
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

Author Information
------------------
Patrick Hermann, stuttgart-things 11/2024
