"""
Calcutta Analytics Pipeline - Local Development

Simplified DAG for local testing that uses the moneyball CLI directly.
Writes to both parquet (for debugging) and database (for API serving).
"""
from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.bash import BashOperator

default_args = {
    'owner': 'calcutta',
    'depends_on_past': False,
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=2),
}

dag = DAG(
    'calcutta_pipeline_local',
    default_args=default_args,
    description='Local Calcutta analytics pipeline with database writes',
    schedule_interval=None,  # Manual trigger
    start_date=datetime(2025, 1, 1),
    catchup=False,
    tags=['calcutta', 'local', 'development'],
)

# Configuration from DAG run
YEAR = '{{ dag_run.conf.get("year", "2025") }}'
N_SIMS = '{{ dag_run.conf.get("n_sims", "5000") }}'
SEED = '{{ dag_run.conf.get("seed", "42") }}'
STRATEGY = '{{ dag_run.conf.get("strategy", "minlp") }}'

# Base command with environment
BASE_CMD = """
export CALCUTTA_WRITE_TO_DB=true
export DB_HOST=calcutta-db-1
export DB_PORT=5432
export DB_NAME=calcutta
export DB_USER=calcutta
export DB_PASSWORD=calcutta
"""

# Run full pipeline with all stages
run_pipeline = BashOperator(
    task_id='run_full_pipeline',
    bash_command=f"""
{BASE_CMD}
# Create snapshot directory if it doesn't exist
mkdir -p /tmp/out/{YEAR}

# Run each stage of the pipeline
python -m moneyball.cli predicted-game-outcomes /tmp/out/{YEAR} --n-sims {N_SIMS} --seed {SEED} && \\
python -m moneyball.cli simulate-tournaments /tmp/out/{YEAR} --n-sims {N_SIMS} --seed {SEED} && \\
python -m moneyball.cli predicted-auction-share-of-pool /tmp/out/{YEAR} && \\
python -m moneyball.cli recommended-entry-bids /tmp/out/{YEAR} --strategy {STRATEGY} && \\
python -m moneyball.cli simulated-entry-outcomes /tmp/out/{YEAR} --n-sims {N_SIMS} --seed {SEED} && \\
python -m moneyball.cli investment-report /tmp/out/{YEAR}
    """,
    dag=dag,
)

# Optional: Verify database writes
verify_database = BashOperator(
    task_id='verify_database',
    bash_command=f"""
{BASE_CMD}
psql -h calcutta-db-1 -p 5432 -U calcutta -d calcutta -c "
SELECT
    'bronze_tournaments' as table_name,
    COUNT(*) as row_count
FROM bronze_tournaments
UNION ALL
SELECT
    'bronze_teams',
    COUNT(*)
FROM bronze_teams
UNION ALL
SELECT
    'silver_simulated_tournaments',
    COUNT(*)
FROM silver_simulated_tournaments
UNION ALL
SELECT
    'silver_predicted_game_outcomes',
    COUNT(*)
FROM silver_predicted_game_outcomes
UNION ALL
SELECT
    'gold_recommended_entry_bids',
    COUNT(*)
FROM gold_recommended_entry_bids;
"
    """,
    dag=dag,
)

run_pipeline >> verify_database
