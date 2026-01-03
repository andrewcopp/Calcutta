"""
Calcutta Analytics Pipeline DAG

Orchestrates the end-to-end analytics pipeline:
1. Go: Simulate tournaments (compute-intensive)
2. Python: Predict game outcomes (ML)
3. Python: Predict market share (ML)
4. Python: Optimize portfolio (MINLP)
5. Go: Evaluate all entries (compute-intensive)
"""
from datetime import datetime, timedelta
from airflow import DAG
from airflow.providers.docker.operators.docker import DockerOperator


default_args = {
    'owner': 'calcutta',
    'depends_on_past': False,
    'email_on_failure': True,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

dag = DAG(
    'calcutta_analytics_pipeline',
    default_args=default_args,
    description='End-to-end Calcutta tournament analytics pipeline',
    schedule_interval=None,  # Manual trigger only
    start_date=datetime(2025, 1, 1),
    catchup=False,
    tags=['calcutta', 'analytics', 'tournament'],
)

# Configuration
YEAR = '{{ dag_run.conf.get("year", 2025) }}'
N_SIMS = '{{ dag_run.conf.get("n_sims", 5000) }}'
SEED = '{{ dag_run.conf.get("seed", 42) }}'
STRATEGY = '{{ dag_run.conf.get("strategy", "minlp") }}'

# Database connection (shared across all tasks)
DB_ENV = {
    'DB_HOST': 'postgres',
    'DB_PORT': '5432',
    'DB_NAME': 'calcutta_analytics',
    'DB_USER': 'postgres',
    'DB_PASSWORD': 'postgres',
}

# ============================================================================
# BRONZE LAYER: Raw Data Collection
# ============================================================================

load_tournament_data = DockerOperator(
    task_id='load_tournament_data',
    image='calcutta-go-loader:latest',
    command=f'./loader --year {YEAR}',
    environment=DB_ENV,
    network_mode='bridge',
    auto_remove=True,
    dag=dag,
)

# ============================================================================
# BRONZE LAYER: Tournament Simulation (Go - Compute Intensive)
# ============================================================================

simulate_tournaments = DockerOperator(
    task_id='simulate_tournaments',
    image='calcutta-go-simulator:latest',
    command=f'./simulator --year {YEAR} --n-sims {N_SIMS} --seed {SEED}',
    environment=DB_ENV,
    network_mode='bridge',
    auto_remove=True,
    docker_url='unix://var/run/docker.sock',
    mount_tmp_dir=False,
    dag=dag,
)

# ============================================================================
# SILVER LAYER: ML Predictions (Python - sklearn)
# ============================================================================

predict_game_outcomes = DockerOperator(
    task_id='predict_game_outcomes',
    image='calcutta-python-ml:latest',
    command=f'python -m moneyball.predict_games --year {YEAR}',
    environment={
        **DB_ENV,
        'PYTHONPATH': '/app',
    },
    network_mode='bridge',
    auto_remove=True,
    dag=dag,
)

predict_market_share = DockerOperator(
    task_id='predict_market_share',
    image='calcutta-python-ml:latest',
    command=f'python -m moneyball.predict_market --year {YEAR}',
    environment={
        **DB_ENV,
        'PYTHONPATH': '/app',
    },
    network_mode='bridge',
    auto_remove=True,
    dag=dag,
)

calculate_team_value = DockerOperator(
    task_id='calculate_team_value',
    image='calcutta-python-ml:latest',
    command=f'python -m moneyball.calculate_team_value --year {YEAR}',
    environment={
        **DB_ENV,
        'PYTHONPATH': '/app',
    },
    network_mode='bridge',
    auto_remove=True,
    dag=dag,
)

# ============================================================================
# GOLD LAYER: Portfolio Optimization (Python - MINLP)
# ============================================================================

optimize_portfolio = DockerOperator(
    task_id='optimize_portfolio',
    image='calcutta-python-optimizer:latest',
    command=f'python -m moneyball.optimize --year {YEAR} --strategy {STRATEGY} --n-sims {N_SIMS} --seed {SEED}',
    environment={
        **DB_ENV,
        'PYTHONPATH': '/app',
    },
    network_mode='bridge',
    auto_remove=True,
    dag=dag,
)

# ============================================================================
# GOLD LAYER: Entry Evaluation (Go - Compute Intensive)
# ============================================================================

evaluate_all_entries = DockerOperator(
    task_id='evaluate_all_entries',
    image='calcutta-go-evaluator:latest',
    command=f'./evaluator --year {YEAR} --n-sims {N_SIMS}',
    environment=DB_ENV,
    network_mode='bridge',
    auto_remove=True,
    dag=dag,
)

# ============================================================================
# Task Dependencies
# ============================================================================

# Bronze layer
load_tournament_data >> simulate_tournaments

# Silver layer (parallel after simulation)
simulate_tournaments >> [predict_game_outcomes, predict_market_share, calculate_team_value]

# Gold layer (optimizer needs all predictions)
[predict_game_outcomes, predict_market_share, calculate_team_value] >> optimize_portfolio

# Gold layer (evaluation needs optimizer output)
optimize_portfolio >> evaluate_all_entries

# Final aggregation
# Go evaluation is the final step.
