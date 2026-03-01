import yaml

name = option("name") or "nginx"
namespace = option("namespace") or "default"
image = option("image") or "nginx:1.21"
port = int(option("port") or "80")
replicas = int(option("replicas") or "2")

deployment = {
    apiVersion = "apps/v1"
    kind = "Deployment"
    metadata = {
        name = name
        namespace = namespace
        labels = {
            app = name
        }
    }
    spec = {
        replicas = replicas
        selector.matchLabels = {
            app = name
        }
        template = {
            metadata.labels = {
                app = name
            }
            spec.containers = [
                {
                    name = name
                    image = image
                    ports = [
                        {
                            containerPort = port
                        }
                    ]
                }
            ]
        }
    }
}

service = {
    apiVersion = "v1"
    kind = "Service"
    metadata = {
        name = name
        namespace = namespace
        labels = {
            app = name
        }
    }
    spec = {
        selector = {
            app = name
        }
        ports = [
            {
                port = port
                targetPort = port
                protocol = "TCP"
            }
        ]
        type = "ClusterIP"
    }
}

print("---")
print(yaml.encode(deployment))
print("---")
print(yaml.encode(service))
