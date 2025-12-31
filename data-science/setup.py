from setuptools import setup, find_packages

setup(
    name="moneyball",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "pandas>=2.0.0",
        "numpy>=1.24.0",
        "scipy>=1.10.0",
        "pyarrow>=12.0.0",
        "psycopg2-binary>=2.9.0",
    ],
    python_requires=">=3.11",
    author="Calcutta",
    description="NCAA Tournament Calcutta Analytics Pipeline",
)
