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

# Consul
helm_repo("hashicorp", "https://helm.releases.hashicorp.com")
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

services = [
  "api-gateway",
  "auth-service",
]

for service in services:
  compile_cmd = "./scripts/compile_service.sh {}".format(service)
  local_resource(
    "{}-compile".format(service),
    compile_cmd,
    deps = ["./services/{}/".format(service), "./shared"],
    labels = "compiles",
  )

  docker_build_with_restart(
    "vasapolrittideah/money-tracker-api-{}:tilt".format(service),
    ".",
    entrypoint = ["/app/build/{}".format(service)],
    dockerfile = "./infrastructure/docker/{}/Dockerfile.dev".format(service),
    only = ["./build/{}".format(service), "./shared"],
    live_update = [
      sync("./build", "/app/build"),
      sync("./shared", "/app/shared"),
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
    )

    k8s_resource(
      workload = "money-tracker-api-{}".format(service),
      new_name = service,
      labels = "services",
      resource_deps = ["consul", db_name],
    )
