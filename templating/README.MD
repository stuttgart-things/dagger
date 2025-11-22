# TEMPLATING

This module provides template rendering capabilities using Go's text/template package.
It supports loading templates from local files or HTTPS URLs, and accepts variables from YAML files or command-line parameters.

```bash
dagger call -m templating render \
--src=tests/templating \
--templates="config.yaml.tmpl,deployment.yaml.tmpl,service.yaml.tmpl" \   --variables="name=myapp,namespace=default,version=1.0.0,replicas=3,image=nginx,port=8080,serviceType=ClusterIP,debug=false" \
--variables-file=example-vars.yaml \
export --path=./rendered
```
