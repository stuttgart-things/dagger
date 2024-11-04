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

## INSTAL EXTERNAL DAGGER MODULE

```bash
dagger install github.com/purpleclay/daggerverse/golang@v0.5.0
```

```bash
MODULE=example #example
dagger functions -m ${MODULE}
```


## CALL FUNCTION (FROM DIFFERENT REPO / FROM EVERYWHERE)

```bash
MODULE=golang #example
dagger call -m github.com/stuttgart-things/dagger/${MODULE} build --progress plain --src ./ export --path build
```