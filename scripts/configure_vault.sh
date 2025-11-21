#!/bin/bash
set -e

export VAULT_ADDR='http://localhost:8200'
export VAULT_TOKEN='root'

# Wait for Vault to be ready
echo "Waiting for Vault to be ready..."
until vault status > /dev/null 2>&1; do
  echo "Vault not ready yet, retrying..."
  sleep 2
done

echo "Vault is ready. Starting configuration..."

# Enable Kubernetes auth
vault auth enable kubernetes 2>/dev/null || echo "Kubernetes auth already enabled"

# Configure Kubernetes auth
vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc:443"

# =====================================================================
# Configure secrets for microservices

SECRETS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../secrets" && pwd)"
load_env() {
  local service_name=$1
  local env_file="${SECRETS_DIR}/.env.${service_name}"

  if [ ! -f "$env_file" ]; then
    echo "Environment file $env_file not found for service $service_name"
    exit 1
  fi

  echo "[${service_name}] Loading environment variables from $env_file..."

  set -a
  source "$env_file"
  set +a
}

# Auth Service
load_env "auth-service"

echo "[auth-service] Creating policy..."
vault policy write auth-service - <<EOF
path "secret/data/auth-service/*" {
  capabilities = ["read"]
}
EOF

echo "[auth-service] Creating Kubernetes role..."
vault write auth/kubernetes/role/auth-service \
  bound_service_account_names=money-tracker-api-auth-service \
  bound_service_account_namespaces=default \
  policies=auth-service \
  audience=vault \
  ttl=24h

echo "[auth-service] Creating secrets in Vault..."
vault kv put secret/auth-service/mongodb \
  MONGO_URI="${MONGO_URI}" \
  MONGO_DB="${MONGO_DB}"

vault kv put secret/auth-service/jwt \
  ACCESS_TOKEN_SECRET="${ACCESS_TOKEN_SECRET}" \
  REFRESH_TOKEN_SECRET="${REFRESH_TOKEN_SECRET}" \
  ACCESS_TOKEN_EXPIRES_IN="${ACCESS_TOKEN_EXPIRES_IN}" \
  REFRESH_TOKEN_EXPIRES_IN="${REFRESH_TOKEN_EXPIRES_IN}" \
  TOKEN_ISSUER="${TOKEN_ISSUER}"

vault kv put secret/auth-service/smtp \
  SMTP_HOST="${SMTP_HOST}" \
  SMTP_PORT="${SMTP_PORT}" \
  SMTP_USERNAME="${SMTP_USERNAME}" \
  SMTP_PASSWORD="${SMTP_PASSWORD}" \
  SMTP_FROM="${SMTP_FROM}"

echo "Vault configured successfully!"
