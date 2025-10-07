⏺ 🧪 Comprehensive Testing Strategy

  🎯 Testing Goals

  1. ✅ Cloud Database Deployment - Verify PostgreSQL backend works
  2. ✅ Local Development - Ensure MemStore still works for dev
  3. ✅ Environment Switching - Test configuration toggles
  4. ✅ API Compatibility - Ensure no breaking changes
  5. ✅ Data Persistence - Verify workloads survive pod restarts

  📋 Phase 1: Infrastructure Setup

  Step 1.1: Deploy Cloud Infrastructure

  # Destroy existing infrastructure
  ./scripts/destroy.sh

  # Deploy new Terraform infrastructure
  cd terraform/
  cp terraform.tfvars.example terraform.tfvars

  # Edit terraform.tfvars:
  cat > terraform.tfvars << EOF
  aws_region  = "us-east-1"
  aws_profile = "myclaude"
  environment = "prod"
  cluster_name_prefix = "aegis-spoke"
  cluster_version = "1.33"
  cpu_desired_capacity = 1
  gpu_desired_capacity = 0
  use_spot_instances = true
  db_instance_class = "db.t3.micro"
  ecr_repositories = ["aegis/k8s-agent", "aegis/proxy", "aegis/platform-api"]
  EOF

  # Deploy infrastructure
  make dev-up

  Step 1.2: Build and Push Images

  # Get ECR login
  aws ecr get-login-password --region us-east-1 --profile myclaude | \
    docker login --username AWS --password-stdin \
    $(terraform output -raw ecr_registry_url)

  # Build and push platform-api (new)
  cd ../services/platform-api/
  docker build -t platform-api .
  docker tag platform-api:latest \
    $(cd ../../terraform && terraform output -raw ecr_registry_url)/aegis/platform-api:latest
  docker push $(cd ../../terraform && terraform output -raw ecr_registry_url)/aegis/platform-api:latest

  # Build and push existing images
  cd ../proxy/
  docker build -t proxy .
  docker tag proxy:latest \
    $(cd ../../terraform && terraform output -raw ecr_registry_url)/aegis/proxy:latest
  docker push $(cd ../../terraform && terraform output -raw ecr_registry_url)/aegis/proxy:latest

  cd ../../agents/k8s-agent/
  docker build -t k8s-agent .
  docker tag k8s-agent:latest \
    $(cd ../../terraform && terraform output -raw ecr_registry_url)/aegis/k8s-agent:latest
  docker push $(cd ../../terraform && terraform output -raw ecr_registry_url)/aegis/k8s-agent:latest

  📋 Phase 2: Cloud Database Testing

  Step 2.1: Deploy aegis-services with PostgreSQL

  cd terraform/

  # Create updated values file for cloud deployment
  cat > ../charts/aegis-services/values-test-cloud.yaml << EOF
  # Test values for cloud deployment with PostgreSQL
  platformApi:
    enabled: true
    replicaCount: 1  # Start with 1 for testing

    image:
      repository: $(terraform output -raw helm_values | jq -r '.platform_api.image_repository')
      tag: "latest"
      pullPolicy: Always

    env:
      LOG_level: "debug"  # More verbose for testing
      ENVIRONMENT: "test"
      AEGIS_STORE_BACKEND: "postgres"  # Key test: Use PostgreSQL
      DB_HOST: $(terraform output -raw helm_values | jq -r '.platform_api.db_host')
      DB_PORT: "5432"
      DB_NAME: "aegis"
      DB_USER: "aegis_api"
      DB_SSLMODE: "require"
      AEGIS_PROXY_BASE_URL: "https://proxy-test.example.com"
      AEGIS_PROXY_EXPECTED_AUDIENCE: "aegis-proxy"

    # We'll set secrets via command line
    secrets:
      db-password: ""
      proxy-jwt-secret: ""

  proxy:
    enabled: true
    replicaCount: 1

    image:
      repository: $(terraform output -raw helm_values | jq -r '.proxy.image_repository')
      tag: "latest"
      pullPolicy: Always

    jwtSecret: ""  # Will be set via command line
  EOF

  # Deploy with secrets from AWS Secrets Manager
  kubectl create namespace aegis-services --dry-run=client -o yaml | kubectl apply -f -

  # Get secrets and deploy
  DB_PASSWORD=$(aws secretsmanager get-secret-value \
    --secret-id "$(terraform output -raw secrets | jq -r '.db_password_secret_name')" \
    --query SecretString --output text \
    --profile myclaude --region us-east-1)

  JWT_SECRET=$(aws secretsmanager get-secret-value \
    --secret-id "$(terraform output -raw secrets | jq -r '.jwt_secret_secret_name')" \
    --query SecretString --output text \
    --profile myclaude --region us-east-1)

  helm upgrade --install aegis-services ../charts/aegis-services \
    -n aegis-services \
    -f ../charts/aegis-services/values-test-cloud.yaml \
    --set-string platformApi.secrets.db-password="$DB_PASSWORD" \
    --set-string proxy.jwtSecret="$JWT_SECRET" \
    --wait --timeout=600s

  Step 2.2: Test Database Connectivity

  # Check pod status
  kubectl get pods -n aegis-services

  # Check platform-api logs for database connection
  kubectl logs -l app.kubernetes.io/component=platform-api -n aegis-services --tail=50

  # Test database connection directly
  kubectl run -it --rm debug --image=postgres:15 \
    --env="PGPASSWORD=$DB_PASSWORD" \
    -- psql -h $(terraform output -raw rds_endpoint | cut -d: -f1) \
      -U aegis_api -d aegis -c "\dt"

  📋 Phase 3: API Functionality Testing

  Step 3.1: Port Forward and Test APIs

  # Port forward platform-api
  kubectl port-forward svc/aegis-services-platform-api 8080:8080 -n aegis-services &
  kubectl port-forward svc/aegis-services-platform-api 8081:8081 -n aegis-services &

  # Wait for port forwards
  sleep 5

  # Test health endpoints
  curl http://localhost:8080/healthz
  curl http://localhost:8080/readyz

  # Test API functionality
  curl -X POST http://localhost:8080/v1/projects \
    -H "Content-Type: application/json" \
    -d '{
      "project": {
        "id": "test-project-1",
        "display_name": "Test Project",
        "owner_group": "test-team"
      }
    }'

  # Test budget creation
  curl -X POST http://localhost:8080/v1/budgets \
    -H "Content-Type: application/json" \
    -d '{
      "budget": {
        "project_id": "test-project-1",
        "limit_usd": 100.0,
        "policy_mode": "SOFT"
      }
    }'

  # Test flavor creation
  curl -X POST http://localhost:8080/v1/flavors \
    -H "Content-Type: application/json" \
    -d '{
      "flavor": {
        "name": "test-gpu-flavor",
        "chip": "NVIDIA T4",
        "gpu_count": 1,
        "price_usd_per_gpu_hour": 0.50
      }
    }'

  Step 3.2: Test Data Persistence

  # Create test data
  curl -X POST http://localhost:8080/v1/projects \
    -H "Content-Type: application/json" \
    -d '{
      "project": {
        "id": "persistence-test",
        "display_name": "Persistence Test Project",
        "owner_group": "test-team"
      }
    }'

  # Restart platform-api pod to test persistence
  kubectl delete pod -l app.kubernetes.io/component=platform-api -n aegis-services

  # Wait for pod to restart
  kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=platform-api -n aegis-services --timeout=120s

  # Re-establish port forward
  kubectl port-forward svc/aegis-services-platform-api 8080:8080 -n aegis-services &
  sleep 5

  # Verify data persisted
  curl http://localhost:8080/v1/projects/persistence-test

  # This should return the project we created before the restart!

  📋 Phase 4: Local Development Testing

  Step 4.1: Test Local MemStore Mode

  cd ../services/platform-api/

  # Test with memory backend (default)
  export AEGIS_STORE_BACKEND=memory  # or just omit
  export GRPC_ADDR=":8082"  # Different port to avoid conflicts
  export HTTP_ADDR=":8083"

  # Run locally
  go run ./... &
  LOCAL_PID=$!

  sleep 3

  # Test local API
  curl http://localhost:8083/healthz
  curl -X POST http://localhost:8083/v1/projects \
    -H "Content-Type: application/json" \
    -d '{
      "project": {
        "id": "local-test",
        "display_name": "Local Test Project",
        "owner_group": "dev-team"
      }
    }'

  # Verify it's in memory (restart and check data is gone)
  kill $LOCAL_PID
  go run ./... &
  LOCAL_PID=$!
  sleep 3

  # This should return 404 (data not persisted)
  curl http://localhost:8083/v1/projects/local-test
  kill $LOCAL_PID

  Step 4.2: Test Local PostgreSQL Mode

  # Start local PostgreSQL with Docker
  docker run --name aegis-local-db -d \
    -e POSTGRES_DB=aegis_local \
    -e POSTGRES_USER=aegis_dev \
    -e POSTGRES_PASSWORD=dev_password \
    -p 5433:5432 \
    postgres:15

  # Wait for DB to start
  sleep 10

  # Run migrations locally
  cd ../../../terraform/
  export DATABASE_URL="postgres://aegis_dev:dev_password@localhost:5433/aegis_local?sslmode=disable"
  docker run --rm -v $(pwd)/../services/platform-api/migrations:/migrations \
    ghcr.io/golang-migrate/migrate:v4.16.2 \
    -path=/migrations -database="$DATABASE_URL" up

  # Test platform-api with local PostgreSQL
  cd ../services/platform-api/
  export AEGIS_STORE_BACKEND=postgres
  export DATABASE_URL="postgres://aegis_dev:dev_password@localhost:5433/aegis_local?sslmode=disable"
  export GRPC_ADDR=":8084"
  export HTTP_ADDR=":8085"

  go run ./... &
  LOCAL_PG_PID=$!
  sleep 3

  # Test local PostgreSQL API
  curl http://localhost:8085/healthz
  curl -X POST http://localhost:8085/v1/projects \
    -H "Content-Type: application/json" \
    -d '{
      "project": {
        "id": "local-pg-test",
        "display_name": "Local PostgreSQL Test",
        "owner_group": "dev-team"
      }
    }'

  # Test persistence (restart and verify data remains)
  kill $LOCAL_PG_PID
  go run ./... &
  LOCAL_PG_PID=$!
  sleep 3

  # This should return the project (data persisted)
  curl http://localhost:8085/v1/projects/local-pg-test
  kill $LOCAL_PG_PID

  # Cleanup
  docker stop aegis-local-db
  docker rm aegis-local-db

  📋 Phase 5: Integration Testing

  Step 5.1: Test k8s-agent Integration

  # Deploy aegis-spoke (k8s-agent) pointing to cloud platform-api
  cd terraform/

  # Get platform-api service endpoint
  PLATFORM_API_ENDPOINT=$(kubectl get svc aegis-services-platform-api -n aegis-services -o
  jsonpath='{.spec.clusterIP}'):8081

  # Deploy aegis-spoke
  kubectl create namespace aegis-spoke --dry-run=client -o yaml | kubectl apply -f -

  cat > ../charts/aegis-spoke/values-test-cloud.yaml << EOF
  k8sAgent:
    image:
      repository: $(terraform output -raw helm_values | jq -r '.k8s_agent.image_repository')
      tag: "latest"
    env:
      AEGIS_CLUSTER_ID: "$(terraform output -raw helm_values | jq -r '.k8s_agent.cluster_id')"
      AEGIS_CP_GRPC: "$PLATFORM_API_ENDPOINT"  # Point to platform-api service
      AEGIS_PROXY_INGRESS_HOST: "test-proxy.example.com"

  proxy:
    enabled: false  # We have proxy in aegis-services namespace
  EOF

  helm upgrade --install aegis-spoke ../charts/aegis-spoke \
    -n aegis-spoke \
    -f ../charts/aegis-spoke/values-test-cloud.yaml

  # Check k8s-agent logs for platform-api connectivity
  kubectl logs -l app.kubernetes.io/component=k8s-agent -n aegis-spoke --tail=50

  Step 5.2: End-to-End Workload Test

  # Create a test workload via platform-api
  curl -X POST http://localhost:8080/v1/workloads \
    -H "Content-Type: application/json" \
    -d '{
      "workload": {
        "project_id": "test-project-1",
        "queue": "",
        "kind": {
          "workspace": {
            "flavor": "test-gpu-flavor",
            "image": "nginx:latest",
            "interactive": true,
            "ports": [22, 8080]
          }
        }
      }
    }'

  # Check if workload was created and placed
  kubectl get aegisworkloads -n aegis-spoke

  # Verify workload appears in database
  curl http://localhost:8080/v1/workloads?project_id=test-project-1

  📋 Phase 6: Environment Toggle Testing

  Step 6.1: Test Helm Environment Switching

  # Test local Helm deployment with memory store
  helm upgrade --install aegis-services-local ../charts/aegis-services \
    -n aegis-services-local \
    --create-namespace \
    -f ../charts/aegis-services/values/common.yaml \
    -f ../charts/aegis-services/values/local.yaml \
    --set platformApi.env.AEGIS_STORE_BACKEND=memory

  # Test local Helm deployment with PostgreSQL subchart
  helm upgrade --install aegis-services-local-pg ../charts/aegis-services \
    -n aegis-services-local-pg \
    --create-namespace \
    -f ../charts/aegis-services/values/common.yaml \
    -f ../charts/aegis-services/values/local.yaml \
    --set postgresql.enabled=true \
    --set platformApi.env.AEGIS_STORE_BACKEND=postgres

  🧹 Cleanup Script

⏺ Write(~/code/aegis/scripts/test-cleanup.sh)
  ⎿  Wrote 22 lines to ../../../../scripts/test-cleanup.sh
     #!/bin/bash
     # Cleanup script for testing environments
     echo "🧹 Cleaning up test environments..."
     # Kill any local port forwards
     pkill -f "kubectl port-forward" || true
     # Remove local test namespaces
     … +12 lines (ctrl+o to expand)

⏺ Bash(chmod +x /Users/carlossanchez/code/aegis/scripts/test-cleanup.sh)
  ⎿  Error: Error

⏺ 📊 Success Criteria Checklist

  ✅ Cloud Database Deployment

  - Infrastructure deploys via Terraform
  - Platform-api connects to RDS PostgreSQL
  - Database migrations run successfully
  - APIs respond correctly
  - Data persists after pod restarts

  ✅ Local Development

  - Memory store works for quick iteration
  - Local PostgreSQL works for database testing
  - Environment switching works via AEGIS_STORE_BACKEND
  - Local and cloud APIs have identical responses

  ✅ Integration Testing

  - k8s-agent connects to platform-api
  - Workloads can be submitted and placed
  - End-to-end workflow functions
  - No breaking changes to existing APIs

  ✅ Configuration Validation

  - Helm values work for both environments
  - Secrets management functions correctly
  - Environment variables switch backends properly
  - Health checks pass in all configurations

  🚀 Quick Start Testing

  To run the complete test suite:

  # 1. Deploy infrastructure and test cloud database
  cd terraform/
  make dev-up

  # 2. Run the complete test suite
  # Follow Phase 1-6 above step by step

  # 3. Cleanup when done
  ./scripts/test-cleanup.sh
  make dev-down  # Destroys cloud infrastructure

  This comprehensive testing plan will validate that:
  1. Your database implementation works correctly
  2. Local development remains fast and easy
  3. Cloud deployment is production-ready
  4. Environment switching is seamless
  5. No existing functionality is broken

  The tests are designed to be incremental - you can stop at any phase and still have confidence in that level of
  functionality. Start with Phase 1-2 to validate the core database functionality first!
