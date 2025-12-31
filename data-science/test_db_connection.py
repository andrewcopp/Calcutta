"""Test database connection and writers."""
import os
os.environ["CALCUTTA_WRITE_TO_DB"] = "true"
os.environ["CALCUTTA_ANALYTICS_DB_HOST"] = "localhost"
os.environ["CALCUTTA_ANALYTICS_DB_PORT"] = "5433"
os.environ["CALCUTTA_ANALYTICS_DB_NAME"] = "calcutta_analytics"
os.environ["CALCUTTA_ANALYTICS_DB_USER"] = "postgres"
os.environ["CALCUTTA_ANALYTICS_DB_PASSWORD"] = "postgres"

from moneyball.pipeline.db_writer import get_db_writer

print("Testing database connection...")
db_writer = get_db_writer()

if db_writer.enabled:
    print("✓ Database writer is enabled")
    print(f"  Host: {os.getenv('CALCUTTA_ANALYTICS_DB_HOST')}")
    print(f"  Port: {os.getenv('CALCUTTA_ANALYTICS_DB_PORT')}")
    print(f"  Database: {os.getenv('CALCUTTA_ANALYTICS_DB_NAME')}")
else:
    print("✗ Database writer is disabled")
    print("  Check environment variables and database connection")
