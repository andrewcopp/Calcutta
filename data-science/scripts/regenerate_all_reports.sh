#!/bin/bash
# Regenerate investment reports for all years using greedy strategy

set -e

# Activate virtualenv
source .venv/bin/activate

YEARS="2017 2018 2019 2021 2022 2023 2024 2025"

echo "Regenerating investment reports with greedy strategy..."
echo ""

for year in $YEARS; do
    echo "Processing $year..."
    python -m moneyball.cli investment-report "out/$year" --strategy greedy
    echo ""
done

echo "Generating summary CSV..."
python scripts/generate_summary.py

echo "Done!"
