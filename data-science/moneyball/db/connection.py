"""
Database connection pooling.

Provides thread-safe connection pooling for Python services to write
directly to the Postgres database.
"""
import atexit
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
        database_url = os.getenv("DATABASE_URL", "").strip()

        logger.info("Creating database connection pool")
        minconn = int(os.getenv("DB_MIN_CONN", "1"))
        maxconn = int(os.getenv("DB_MAX_CONN", "10"))

        connect_timeout = int(os.getenv("DB_CONNECT_TIMEOUT", "5"))

        if database_url:
            _pool = psycopg2.pool.ThreadedConnectionPool(
                minconn=minconn,
                maxconn=maxconn,
                dsn=database_url,
                connect_timeout=connect_timeout,
            )
        else:
            password = os.getenv("DB_PASSWORD", "").strip()
            if not password:
                raise RuntimeError(
                    "DB_PASSWORD or DATABASE_URL must be set"
                )

            _pool = psycopg2.pool.ThreadedConnectionPool(
                minconn=minconn,
                maxconn=maxconn,
                host=os.getenv("DB_HOST", "localhost"),
                port=int(os.getenv("DB_PORT", "5432")),
                database=os.getenv("DB_NAME", "calcutta"),
                user=os.getenv("DB_USER", "calcutta"),
                password=password,
                connect_timeout=connect_timeout,
            )
    return _pool


def _get_connection():
    """Get a connection from the pool."""
    return get_pool().getconn()


def _release_connection(conn):
    """Return a connection to the pool."""
    get_pool().putconn(conn)


@contextmanager
def get_db_connection():
    """
    Context manager for database connections.

    Automatically commits on success and rolls back on exception.

    Usage:
        with get_db_connection() as conn:
            with conn.cursor() as cur:
                cur.execute("INSERT INTO table ...")
    """
    conn = _get_connection()
    try:
        yield conn
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        _release_connection(conn)


def _close_pool():
    """Shut down the connection pool at interpreter exit."""
    global _pool
    if _pool is not None:
        _pool.closeall()
        _pool = None


atexit.register(_close_pool)
