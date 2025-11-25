"""
  Tiltfile for microservices deployment in development environment.
"""

load("ext://helm_resource", "helm_repo", "helm_resource")
load("ext://restart_process", "docker_build_with_restart")

update_settings(
  max_parallel_updates = 3,
  k8s_upsert_timeout_secs = 360,
  suppress_unused_image_warnings = None,
)

helm_repo("hashicorp", "https://helm.releases.hashicorp.com")

# Consul
helm_resource(
  "consul",
  "hashicorp/consul",
  namespace = "consul",
  flags = [
    "--version=1.9.0",
    "--create-namespace",
    "--set=global.name=consul",
    "--values=./infrastructure/helm/consul/values.yaml",
    "--values=./infrastructure/helm/consul/values-dev.yaml",
  ],
  pod_readiness = "ignore",
  resource_deps = ["hashicorp"],
  labels = "tooling",
)

# Vault
helm_resource(
  "vault",
  "hashicorp/vault",
  namespace = "vault",
  flags = [
    "--create-namespace",
    "--set=server.dev.enabled=true",
    "--set=server.dev.devRootToken=root",
    "--set=injector.enabled=true",
    "--set=csi.enabled=true",
    "--set=ui.enabled=true",
    "--values=./infrastructure/helm/vault/values.yaml",
    "--values=./infrastructure/helm/vault/values-dev.yaml",
  ],
  resource_deps = ["hashicorp"],
  labels = "tooling",
)

# Umbrella chart
k8s_yaml(helm(
  "./infrastructure/helm/umbrella",
  name = "money-tracker-api",
  values = [
    "./infrastructure/helm/umbrella/values.yaml",
    "./infrastructure/helm/umbrella/values-dev.yaml",
  ],
))

k8s_resource(
  workload = "consul",
  port_forwards = ["8501:8500"],
  labels = "tooling",
  extra_pod_selectors = [{"component": "server"}],
  discovery_strategy = "selectors-only",
)

k8s_resource(
  workload = "vault",
  port_forwards = ["8200:8200"],
  labels = "tooling",
  extra_pod_selectors = [{"app.kubernetes.io/name": "vault"}],
  discovery_strategy = "selectors-only",
)

local_resource(
  "vault-config",
  "./scripts/configure_vault.sh",
  resource_deps = ["vault"],
)

services = [
  "api-gateway",
  "auth-service",
]

for service in services:
  docker_build(
    "vasapolrittideah/money-tracker-api-{}:tilt".format(service),
    ".",
    dockerfile = "./infrastructure/docker/{}/Dockerfile.dev".format(service),
    live_update = [
      sync("./services/{}".format(service), "/app/services/{}".format(service)),
      sync("./shared", "/app/shared"),
      sync("./go.mod", "/app/go.mod"),
      sync("./go.sum", "/app/go.sum"),
    ],
  )

  if service == "api-gateway":
    k8s_resource(
      workload = "money-tracker-api-{}".format(service),
      new_name = service,
      labels = "services",
      resource_deps = ["consul"],
      port_forwards = ["9000:9000"],
    )
  else:
    db_name = "{}-mongodb".format(service.split("-")[0])

    k8s_resource(
      workload = "money-tracker-api-{}".format(db_name),
      new_name = db_name,
      labels = "databases",
      resource_deps = ["consul"],
      port_forwards = ["27017:27017"],
    )

    k8s_resource(
      workload = "money-tracker-api-{}".format(service),
      new_name = service,
      labels = "services",
      resource_deps = ["consul", "vault-config", db_name],
    )
