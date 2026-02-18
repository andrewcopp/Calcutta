from setuptools import setup, find_packages

setup(
    name="moneyball",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "pandas>=2.2.0",
        "numpy>=1.26.0",
        "pyarrow>=15.0.0",
        "psycopg2-binary>=2.9.0",
    ],
    python_requires=">=3.9",
    author="Calcutta",
    description="NCAA Tournament Calcutta Analytics Pipeline",
)
