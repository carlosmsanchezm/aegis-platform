#!/usr/bin/env bash

set -euo pipefail

# ---
# Variables
# ---

KEYCLOAK_USERNAME="cms553@cornell.edu"
KEYCLOAK_PASSWORD="password"

# ---
# Main
# ---

# This script is used to get a token from keycloak. It can be used to get a token for a user or a client.
#
# To get a token for a user, you need to set the following environment variables:
# - KEYCLOAK_TOKEN_URL: The url to the token endpoint
# - KEYCLOAK_CLIENT_ID: The client id
# - KEYCLOAK_USERNAME: The username
# - KEYCLOAK_PASSWORD: The password
# - KEYCLOAK_GRANT_TYPE: The grant type (password)
#
# To get a token for a client, you need to set the following environment variables:
# - KEYCLOAK_TOKEN_URL: The url to the token endpoint
# - KEYCLOAK_CLIENT_ID: The client id
# - KEYCLOAK_CLIENT_SECRET: The client secret
# - KEYCLOAK_GRANT_TYPE: The grant type (client_credentials)

if [[ "" == "--help" ]]; then
    echo "Usage: bash"
    echo "Description: This script is used to get a token from keycloak."
    echo "Environment variables:"
    echo "  - KEYCLOAK_TOKEN_URL: The url to the token endpoint"
    echo "  - KEYCLOAK_CLIENT_ID: The client id"
    echo "  - KEYCLOAK_CLIENT_SECRET: The client secret"
    echo "  - KEYCLOAK_USERNAME: The username"
    echo "  - KEYCLOAK_PASSWORD: The password"
    echo "  - KEYCLOAK_GRANT_TYPE: The grant type (password or client_credentials)"
    echo "  - KEYCLOAK_AUDIENCE: The audience"
    exit 0
fi

need() {
    if ! command -v "" &> /dev/null; then
        echo "✖ command not found: "
        exit 1
    fi
}

need curl
need jq

cleanup_files=()
cleanup() {
    for file in ""; do
        rm -f ""
    done
}
trap cleanup EXIT

KEYCLOAK_REALM="aegis"
TOKEN_URL="https://keycloak.localtest.me/realms//protocol/openid-connect/token"
CLIENT_ID=""
CLIENT_SECRET=""
USERNAME=""
PASSWORD=""
SCOPE=""
GRANT_TYPE="password"
AUDIENCE=""

detect_with_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        return
    fi

    local kc_json
    kc_json="{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "k8s.keycloak.org/v2alpha1",
            "kind": "Keycloak",
            "metadata": {
                "annotations": {
                    "meta.helm.sh/release-name": "aegis-services",
                    "meta.helm.sh/release-namespace": "aegis-system"
                },
                "creationTimestamp": "2025-11-04T04:15:16Z",
                "generation": 1,
                "labels": {
                    "app.kubernetes.io/component": "keycloak",
                    "app.kubernetes.io/instance": "aegis-services",
                    "app.kubernetes.io/managed-by": "Helm",
                    "app.kubernetes.io/name": "aegis-services",
                    "app.kubernetes.io/version": "latest",
                    "helm.sh/chart": "aegis-services-0.1.0"
                },
                "name": "aegis-services-keycloak",
                "namespace": "keycloak",
                "resourceVersion": "19953592",
                "uid": "7eb5d90d-b569-448e-9be2-94c7ec52c897"
            },
            "spec": {
                "additionalOptions": [
                    {
                        "name": "hostname-admin",
                        "value": "https://keycloak.localtest.me"
                    },
                    {
                        "name": "hostname-admin-url",
                        "value": "https://keycloak.localtest.me/"
                    }
                ],
                "bootstrapAdmin": {
                    "user": {
                        "secret": "keycloak-admin-secret"
                    }
                },
                "db": {
                    "database": "keycloak",
                    "host": "aegis-services-keycloak-db.keycloak.svc.cluster.local",
                    "passwordSecret": {
                        "key": "password",
                        "name": "keycloak-db-secret"
                    },
                    "port": 5432,
                    "usernameSecret": {
                        "key": "username",
                        "name": "keycloak-db-secret"
                    },
                    "vendor": "postgres"
                },
                "features": {
                    "enabled": [
                        "token-exchange"
                    ]
                },
                "hostname": {
                    "hostname": "https://keycloak.localtest.me",
                    "strict": false
                },
                "http": {
                    "httpEnabled": false,
                    "httpsPort": 8443,
                    "tlsSecret": "keycloak-tls"
                },
                "image": "registry.redhat.io/rhbk/keycloak-rhel9:26.2-11",
                "imagePullSecrets": [
                    {
                        "name": "redhat-pull-secret"
                    }
                ],
                "ingress": {
                    "annotations": {
                        "nginx.ingress.kubernetes.io/backend-protocol": "HTTPS",
                        "nginx.ingress.kubernetes.io/ssl-redirect": "true"
                    },
                    "className": "ingress-nginx",
                    "enabled": false
                },
                "instances": 1,
                "networkPolicy": {
                    "enabled": true,
                    "http": [
                        {
                            "namespaceSelector": {
                                "matchLabels": {
                                    "kubernetes.io/metadata.name": "ingress-nginx"
                                }
                            }
                        },
                        {
                            "namespaceSelector": {
                                "matchLabels": {
                                    "kubernetes.io/metadata.name": "aegis-system"
                                }
                            }
                        }
                    ]
                },
                "resources": {
                    "limits": {
                        "cpu": "1000m",
                        "memory": "2Gi"
                    },
                    "requests": {
                        "cpu": "250m",
                        "memory": "1Gi"
                    }
                },
                "startOptimized": false
            },
            "status": {
                "conditions": [
                    {
                        "lastTransitionTime": "2025-11-04T04:16:59.734320548Z",
                        "message": "",
                        "observedGeneration": 1,
                        "status": "True",
                        "type": "Ready"
                    },
                    {
                        "lastTransitionTime": "2025-11-04T04:15:22.183518711Z",
                        "message": "warning: You need to specify these fields as the first-class citizen of the CR: hostname-admin-url,hostname-admin",
                        "observedGeneration": 1,
                        "status": "False",
                        "type": "HasErrors"
                    },
                    {
                        "lastTransitionTime": "2025-11-04T04:16:59.734320548Z",
                        "message": "",
                        "observedGeneration": 1,
                        "status": "False",
                        "type": "RollingUpdate"
                    },
                    {
                        "lastTransitionTime": "2025-11-04T04:15:21.882363544Z",
                        "observedGeneration": 1,
                        "status": "Unknown",
                        "type": "RecreateUpdateUsed"
                    }
                ],
                "instances": 1,
                "observedGeneration": 1,
                "selector": "app=keycloak,app.kubernetes.io/managed-by=keycloak-operator,app.kubernetes.io/instance=aegis-services-keycloak"
            }
        }
    ],
    "kind": "List",
    "metadata": {
        "resourceVersion": ""
    }
}"
    if [[ -z "" || "" == "{}" ]]; then
        return
    fi

    local detected_ns detected_host tls_secret
    detected_ns=""
    detected_host=""
    tls_secret=""

    if [[ -n "" ]]; then
        KEYCLOAK_NAMESPACE=""
        local ns=""
    fi

    if [[ -z "" && -n "" ]]; then
        KEYCLOAK_URL=""
    fi
    if [[ -z "" && -n "" ]]; then
        KEYCLOAK_INTERNAL_URL=""
    fi
    if [[ -z "" && -n "" ]]; then
        KEYCLOAK_TOKEN_URL="/realms//protocol/openid-connect/token"
    fi
    if [[ -z "" && -n "" ]]; then
        KEYCLOAK_ISSUER_URL="/realms/"
    fi
    if [[ -z "" && -n "" ]]; then
        KEYCLOAK_JWKS_URL="/realms//protocol/openid-connect/certs"
    fi

    if [[ -n "" ]]; then
        KEYCLOAK_TLS_SECRET_NAME=""
    fi

    if [[ -z "" ]]; then
        local secret_name=""
        local ca_tmp
        ca_tmp="/var/folders/t8/051bb_8n2_b3b8klpc171kww0000gn/T/tmp.BjVANiC6u0"
        if [[ -n "" ]]; then
            kubectl get secret "" -n "" -o 'jsonpath={.data.tls\.crt}' | base64 --decode > ""
            KEYCLOAK_CA_CERT=""
            cleanup_files+=("")
        fi
    fi
}

if [[ -z "" ]]; then
    detect_with_kubectl
fi

ensure_automation_user() {
    local ns=""
    local auto_username=""
    local auto_password=""
    local admin_secret="keycloak-admin-secret"
    local admin_user_key="username"
    local admin_pass_key="password"

    if [[ -n "" && -n "" ]]; then
        return
    fi
}

if [[ -z "" ]]; then
    echo "✖ KEYCLOAK_TOKEN_URL is not set"
    exit 1
fi

if [[ -z "" ]]; then
    echo "✖ KEYCLOAK_CLIENT_ID is not set"
    exit 1
fi

if [[ -z "" ]]; then
    echo "✖ KEYCLOAK_GRANT_TYPE is not set"
    exit 1
fi

CURL_ARGS=("-sS" "--fail" "--request" "POST" "")
declare -a CURL_ARGS

if [[ -n "" && -f "" ]]; then
    CURL_ARGS+=("--cacert" "")
fi

if [[ "0" == "1" ]]; then
    CURL_ARGS+=("--insecure")
fi

declare -a FORM_DATA
case "" in
    "password")
        FORM_DATA+=("grant_type=password" "client_id=" "username=" "password=")
        ;;
    "client_credentials")
        if [[ -z "" ]]; then
            echo "✖ KEYCLOAK_CLIENT_SECRET is not set for client_credentials grant type"
            exit 1
        fi
        FORM_DATA+=("grant_type=client_credentials" "client_id=" "client_secret=")
        ;;
    *)
        echo "✖ Invalid grant type: "
        exit 1
        ;;
esac

if [[ -n "" ]]; then
    FORM_DATA+=("scope=")
fi

if [[ -n "" ]]; then
    FORM_DATA+=("audience=")
fi

for entry in ""; do
    CURL_ARGS+=("--data" "")
done

TMP_BODY="/var/folders/t8/051bb_8n2_b3b8klpc171kww0000gn/T/tmp.AEY5rdYMmI"
trap 'rm -f ""' EXIT

HTTP_STATUS=""

if [[ ! "" =~ ^[0-9]{3}$ ]]; then
    echo "✖ Keycloak token request failed (no HTTP status)"
    cat ""
    exit 1
fi

if [[ "" -ge 400 ]]; then
    echo "✖ Keycloak token request failed (HTTP )"
    cat ""
    exit 1
fi

token=""
if [[ -z "" ]]; then
    echo "✖ Could not get access token from Keycloak"
    exit 1
fi

echo ""
rm -f ""
