# sailtrim

A minimal deployment tool for [Amazon Lightsail Container](https://aws.amazon.com/jp/blogs/news/lightsail-containers-an-easy-way-to-run-your-containers-in-the-cloud/).

## Usage

```
usage: sailtrim [<flags>] <command> [<args> ...]

Flags:
  --help                  Show context-sensitive help (also try --help-long and --help-man).
  --log-level=info        log level (trace, debug, info, warn, error)
  --config="config.yaml"  configuration file path

Commands:
  help [<command>...]
    Show help.

  version
    show version

  deploy
    create new deployment

  update
    update container service

  status [<flags>]
    show container service status

  init --service-name=SERVICE-NAME
    initialize a container service

  logs [<flags>]
    show logs
```

## Configuration

```yaml
# config.yaml
service: service.json
deployment: deployment.json
```


`service.json` represents container service attributes.
```json
{
  "containerServiceName": "container-service-1",
  "power": "micro",
  "scale": 1
}
```


`deployment.json` represents a deployments of container service.
```json
{
  "containers": {
    "nginx": {
      "image": "nginx:latest",
      "command": [],
      "environment": {
        "FOO": "BAR"
      },
      "ports": {
        "80": "HTTP"
      }
    }
  },
  "publicEndpoint": {
    "containerName": "nginx",
    "containerPort": 80,
    "healthCheck": {
      "healthyThreshold": 2,
      "unhealthyThreshold": 2,
      "timeoutSeconds": 2,
      "intervalSeconds": 5,
      "path": "/",
      "successCodes": "200-499"
    }
  }
}
```

## LICENSE

MIT License

Copyright (c) 2020 FUJIWARA Shunichiro
