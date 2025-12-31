"""
Database connection pooling for analytics database.

Provides thread-safe connection pooling for Python services to write
directly to the analytics Postgres database.
"""
import os
import logging
from typing import Optional
import psycopg2
import psycopg2.pool
from contextlib import contextmanager

logger = logging.getLogger(__name__)

_pool: Optional[psycopg2.pool.ThreadedConnectionPool] = None


def get_pool() -> psycopg2.pool.ThreadedConnectionPool:
    """Get or create the connection pool."""
    global _pool
    if _pool is None:
        logger.info("Creating database connection pool")
        _pool = psycopg2.pool.ThreadedConnectionPool(
            minconn=int(os.getenv("DB_MIN_CONN", "1")),
            maxconn=int(os.getenv("DB_MAX_CONN", "10")),
            host=os.getenv("DB_HOST", "localhost"),
            port=int(os.getenv("DB_PORT", "5432")),
            database=os.getenv("DB_NAME", "calcutta"),
            user=os.getenv("DB_USER", "calcutta"),
            password=os.getenv("DB_PASSWORD", "calcutta"),
        )
    return _pool


def get_connection():
    """Get a connection from the pool."""
    return get_pool().getconn()


def release_connection(conn):
    """Return a connection to the pool."""
    get_pool().putconn(conn)


@contextmanager
def get_db_connection():
    """
    Context manager for database connections.
    
    Usage:
        with get_db_connection() as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT * FROM table")
    """
    conn = get_connection()
    try:
        yield conn
    finally:
        release_connection(conn)


def close_pool():
    """Close all connections in the pool."""
    global _pool
    if _pool is not None:
        logger.info("Closing database connection pool")
        _pool.closeall()
        _pool = None
