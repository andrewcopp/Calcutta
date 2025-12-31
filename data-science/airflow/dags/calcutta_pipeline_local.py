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
export CALCUTTA_ANALYTICS_DB_HOST=localhost
export CALCUTTA_ANALYTICS_DB_PORT=5433
export CALCUTTA_ANALYTICS_DB_NAME=calcutta_analytics
export CALCUTTA_ANALYTICS_DB_USER=postgres
export CALCUTTA_ANALYTICS_DB_PASSWORD=postgres
cd /Users/andrewcopp/Developer/Calcutta/data-science
"""

# Run full pipeline with all stages
run_pipeline = BashOperator(
    task_id='run_full_pipeline',
    bash_command=f"""
{BASE_CMD}
python -m moneyball.cli report \\
    --year {YEAR} \\
    --n-sims {N_SIMS} \\
    --seed {SEED} \\
    --strategy {STRATEGY} \\
    --stages predicted_game_outcomes \\
             predicted_auction_share_of_pool \\
             recommended_entry_bids \\
             simulated_tournaments \\
             simulated_entry_outcomes \\
             investment_report
    """,
    dag=dag,
)

# Optional: Verify database writes
verify_database = BashOperator(
    task_id='verify_database',
    bash_command=f"""
{BASE_CMD}
psql -h localhost -p 5433 -U postgres -d calcutta_analytics -c "
SELECT 
    'bronze_simulated_tournaments' as table_name,
    COUNT(*) as row_count 
FROM bronze_simulated_tournaments
UNION ALL
SELECT 
    'silver_predicted_game_outcomes',
    COUNT(*) 
FROM silver_predicted_game_outcomes
UNION ALL
SELECT 
    'gold_optimization_runs',
    COUNT(*) 
FROM gold_optimization_runs;
"
    """,
    dag=dag,
)

run_pipeline >> verify_database
