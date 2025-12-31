# Airflow Setup Guide

## Local Development

### Prerequisites
- Docker and Docker Compose
- Python 3.9+ (for DAG development/testing)

### Quick Start

1. **Start Airflow with Docker Compose**

```bash
cd data-science
docker-compose -f docker-compose.airflow.yml up -d
```

This starts:
- Airflow webserver (port 8080)
- Airflow scheduler
- Postgres metadata DB
- Redis (for Celery executor)

2. **Access Airflow UI**

```
http://localhost:8080
Username: airflow
Password: airflow
```

3. **Trigger a DAG**

Via UI: Navigate to DAGs → `calcutta_analytics_pipeline` → Trigger DAG

Via CLI:
```bash
docker exec -it airflow-scheduler airflow dags trigger calcutta_analytics_pipeline \
  --conf '{"year": 2025, "n_sims": 5000, "strategy": "minlp"}'
```

### Directory Structure

```
data-science/
  airflow/
    dags/                    # DAG definitions
      calcutta_analytics_pipeline.py
    plugins/                 # Custom operators/sensors
    config/
      airflow.cfg           # Airflow configuration
    logs/                   # Task logs
  docker-compose.airflow.yml
```

### Environment Variables

Create `.env` file in `data-science/`:

```bash
# Airflow Core
AIRFLOW__CORE__EXECUTOR=CeleryExecutor
AIRFLOW__CORE__SQL_ALCHEMY_CONN=postgresql+psycopg2://airflow:airflow@postgres-airflow/airflow
AIRFLOW__CELERY__RESULT_BACKEND=db+postgresql://airflow:airflow@postgres-airflow/airflow
AIRFLOW__CELERY__BROKER_URL=redis://:@redis:6379/0

# Analytics Database (separate from Airflow metadata)
CALCUTTA_ANALYTICS_DB_HOST=postgres-analytics
CALCUTTA_ANALYTICS_DB_PORT=5432
CALCUTTA_ANALYTICS_DB_NAME=calcutta_analytics
CALCUTTA_ANALYTICS_DB_USER=postgres
CALCUTTA_ANALYTICS_DB_PASSWORD=postgres

# Docker
AIRFLOW_UID=50000
```

### Development Workflow

1. **Edit DAG**: Modify `airflow/dags/calcutta_analytics_pipeline.py`
2. **Test DAG syntax**: 
   ```bash
   python airflow/dags/calcutta_analytics_pipeline.py
   ```
3. **Refresh in UI**: DAGs auto-refresh every 30 seconds
4. **View logs**: Check `airflow/logs/` or UI

### Building Docker Images

Before running the DAG, build the service images:

```bash
# Go simulator
cd ../go-service
docker build -t calcutta-go-simulator:latest -f Dockerfile.simulator .

# Go evaluator
docker build -t calcutta-go-evaluator:latest -f Dockerfile.evaluator .

# Python ML service
cd ../data-science
docker build -t calcutta-python-ml:latest -f Dockerfile.ml .

# Python optimizer
docker build -t calcutta-python-optimizer:latest -f Dockerfile.optimizer .
```

---

## Production Deployment

### Option 1: Managed Airflow (Recommended)

**AWS MWAA (Managed Workflows for Apache Airflow)**

Pros:
- Fully managed (no infrastructure maintenance)
- Auto-scaling
- Integrated with AWS services (S3, CloudWatch, IAM)
- High availability

Setup:
1. Upload DAGs to S3 bucket
2. Create MWAA environment via AWS Console/Terraform
3. Configure VPC, subnets, security groups
4. Set environment variables

Cost: ~$300-500/month for small environment

**Google Cloud Composer**

Similar to MWAA but on GCP. Good if already using GCP.

### Option 2: Self-Hosted on Kubernetes

**Helm Chart Deployment**

```bash
# Add Airflow Helm repo
helm repo add apache-airflow https://airflow.apache.org
helm repo update

# Install Airflow
helm install airflow apache-airflow/airflow \
  --namespace airflow \
  --create-namespace \
  --values airflow-values.yaml
```

**airflow-values.yaml**:
```yaml
executor: "KubernetesExecutor"  # Each task runs in its own pod

# DAG storage
dags:
  gitSync:
    enabled: true
    repo: https://github.com/your-org/calcutta
    branch: main
    subPath: data-science/airflow/dags

# Database (use managed RDS/CloudSQL)
postgresql:
  enabled: false
externalDatabase:
  type: postgres
  host: your-rds-endpoint.amazonaws.com
  port: 5432
  database: airflow
  user: airflow
  passwordSecret: airflow-postgres-secret

# Resources
workers:
  replicas: 2
  resources:
    requests:
      cpu: 1
      memory: 2Gi
    limits:
      cpu: 2
      memory: 4Gi

# Monitoring
prometheus:
  enabled: true
```

Pros:
- Full control
- Cost-effective at scale
- Can run on existing K8s cluster

Cons:
- More operational overhead
- Need to manage upgrades, scaling, monitoring

### Option 3: Docker Compose (Small Production)

For small-scale production (single server):

```bash
# Use production docker-compose
docker-compose -f docker-compose.airflow.prod.yml up -d

# Behind nginx reverse proxy with SSL
# Use systemd or supervisord for auto-restart
```

Not recommended for high availability or scale.

---

## Best Practices

### 1. DAG Design

**Idempotency**: Tasks should be rerunnable without side effects
```python
# Good: Upsert with conflict handling
INSERT INTO table VALUES (...) 
ON CONFLICT (id) DO UPDATE SET ...

# Bad: Append-only without deduplication
INSERT INTO table VALUES (...)
```

**Atomicity**: Each task should be atomic
```python
# Good: Single responsibility
simulate_tournaments = DockerOperator(...)

# Bad: Multiple responsibilities
simulate_and_optimize = DockerOperator(...)
```

**Failure Handling**: Set retries and alerts
```python
default_args = {
    'retries': 2,
    'retry_delay': timedelta(minutes=5),
    'email_on_failure': True,
    'email': ['alerts@example.com'],
}
```

### 2. Resource Management

**Task Concurrency**: Limit parallel tasks
```python
dag = DAG(
    'calcutta_analytics_pipeline',
    max_active_runs=1,  # Only one DAG run at a time
    concurrency=4,      # Max 4 tasks running concurrently
)
```

**Docker Resource Limits**:
```python
simulate_tournaments = DockerOperator(
    task_id='simulate_tournaments',
    image='calcutta-go-simulator:latest',
    docker_url='unix://var/run/docker.sock',
    mem_limit='4g',
    cpu_quota=200000,  # 2 CPUs
)
```

### 3. Monitoring

**Metrics to Track**:
- DAG run duration
- Task failure rate
- Resource utilization
- Data quality checks

**Alerting**:
- Slack/PagerDuty integration
- Email on failure
- SLA misses

### 4. Security

**Secrets Management**:
```python
# Use Airflow Connections/Variables, not hardcoded secrets
from airflow.models import Variable

db_password = Variable.get("analytics_db_password")
```

**Network Isolation**:
- Run tasks in isolated networks
- Use VPC peering for database access
- Restrict outbound internet access

---

## Troubleshooting

### DAG not appearing in UI
- Check DAG syntax: `python airflow/dags/your_dag.py`
- Check scheduler logs: `docker logs airflow-scheduler`
- Verify DAG is in `dags/` folder

### Task failing with "Image not found"
- Build Docker image: `docker build -t image:tag .`
- Check image name matches DAG definition
- Verify Docker daemon is accessible

### Database connection errors
- Check environment variables
- Verify network connectivity
- Check database credentials

### Out of memory errors
- Increase Docker memory limits
- Reduce batch sizes in tasks
- Use pagination for large datasets

---

## Migration Path

**Phase 1: Local Development** (Current)
- Run Airflow locally with docker-compose
- Develop and test DAGs
- Build Docker images for each service

**Phase 2: Staging Environment**
- Deploy to AWS MWAA or K8s staging cluster
- Test with production-like data volumes
- Validate monitoring and alerting

**Phase 3: Production**
- Deploy to production MWAA/K8s
- Set up CI/CD for DAG deployments
- Configure production monitoring and alerting
- Set up backup and disaster recovery

**Phase 4: Optimization**
- Add data quality checks
- Implement incremental processing
- Optimize resource allocation
- Add caching layers
