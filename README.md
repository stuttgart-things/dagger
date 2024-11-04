# stuttgart-things/dagger
collection of dagger modules

## LIST FUNCTIONS

```bash
MODULE=golang #example
dagger functions -m ${MODULE}/
```

## CREATE NEW FUNCTION

```bash
MODULE=example #example
dagger init --sdk=go --source=./${MODULE} --name=${MODULE}
```

## CALL FUNCTION (FROM DIFFERENT REPO / FROM EVERYWHERE)

```bash
MODULE=golang #example
dagger call -m github.com/stuttgart-things/dagger/${MODULE} build --progress plain --src ./ export --path build
```