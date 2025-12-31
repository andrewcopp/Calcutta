"""
Initialize Calcutta Analytics Postgres database.
"""
import os
import psycopg2
from pathlib import Path


def get_db_connection():
    """Get database connection from environment variables."""
    return psycopg2.connect(
        host=os.getenv("CALCUTTA_ANALYTICS_DB_HOST", "localhost"),
        port=os.getenv("CALCUTTA_ANALYTICS_DB_PORT", "5432"),
        database=os.getenv("CALCUTTA_ANALYTICS_DB_NAME", "calcutta_analytics"),
        user=os.getenv("CALCUTTA_ANALYTICS_DB_USER", "postgres"),
        password=os.getenv("CALCUTTA_ANALYTICS_DB_PASSWORD", "postgres"),
    )


def init_database():
    """Initialize database schema."""
    schema_path = Path(__file__).parent / "schema.sql"
    
    with open(schema_path, 'r') as f:
        schema_sql = f.read()
    
    conn = get_db_connection()
    try:
        with conn.cursor() as cur:
            print("Creating database schema...")
            cur.execute(schema_sql)
            conn.commit()
            print("✓ Database schema created successfully")
    except Exception as e:
        conn.rollback()
        print(f"✗ Error creating schema: {e}")
        raise
    finally:
        conn.close()


def drop_all_tables():
    """Drop all tables (for development/testing)."""
    conn = get_db_connection()
    try:
        with conn.cursor() as cur:
            print("Dropping all tables...")
            
            # Drop views first
            cur.execute("""
                DROP VIEW IF EXISTS view_entry_rankings CASCADE;
                DROP VIEW IF EXISTS view_latest_optimization_runs CASCADE;
                DROP VIEW IF EXISTS view_tournament_sim_stats CASCADE;
            """)
            
            # Drop functions
            cur.execute("""
                DROP FUNCTION IF EXISTS get_entry_portfolio CASCADE;
            """)
            
            # Drop gold tables
            cur.execute("""
                DROP TABLE IF EXISTS gold_detailed_investment_report CASCADE;
                DROP TABLE IF EXISTS gold_entry_performance CASCADE;
                DROP TABLE IF EXISTS gold_entry_simulation_outcomes CASCADE;
                DROP TABLE IF EXISTS gold_recommended_entry_bids CASCADE;
                DROP TABLE IF EXISTS gold_optimization_runs CASCADE;
            """)
            
            # Drop silver tables
            cur.execute("""
                DROP TABLE IF EXISTS silver_team_tournament_value CASCADE;
                DROP TABLE IF EXISTS silver_predicted_market_share CASCADE;
                DROP TABLE IF EXISTS silver_predicted_game_outcomes CASCADE;
            """)
            
            # Drop bronze tables
            cur.execute("""
                DROP TABLE IF EXISTS bronze_payouts CASCADE;
                DROP TABLE IF EXISTS bronze_entry_bids CASCADE;
                DROP TABLE IF EXISTS bronze_calcuttas CASCADE;
                DROP TABLE IF EXISTS bronze_simulated_tournaments CASCADE;
                DROP TABLE IF EXISTS bronze_teams CASCADE;
                DROP TABLE IF EXISTS bronze_tournaments CASCADE;
            """)
            
            conn.commit()
            print("✓ All tables dropped successfully")
    except Exception as e:
        conn.rollback()
        print(f"✗ Error dropping tables: {e}")
        raise
    finally:
        conn.close()


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Initialize Calcutta Analytics database")
    parser.add_argument("--drop", action="store_true", help="Drop all tables first")
    args = parser.parse_args()
    
    if args.drop:
        drop_all_tables()
    
    init_database()
