#!/bin/bash
# Run investment report pipeline for all years 2017-2025

set -e

# Activate virtual environment if it exists
if [ -d ".venv" ]; then
    source .venv/bin/activate
fi

YEARS=(2017 2018 2019 2020 2021 2022 2023 2024 2025)

for year in "${YEARS[@]}"; do
    # Check if snapshot directory exists
    if [ ! -d "out/$year" ]; then
        echo "⚠️  Skipping $year (no snapshot data)"
        continue
    fi
    
    echo "========================================="
    echo "Running pipeline for year: $year"
    echo "========================================="
    
    PYTHONPATH=. python -m moneyball.cli investment-report \
        "out/$year" \
        --snapshot-name "$year" \
        --n-sims 5000 \
        --seed 123 \
        --budget-points 100
    
    echo ""
done

echo "========================================="
echo "All years complete!"
echo "========================================="
