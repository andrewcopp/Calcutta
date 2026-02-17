"""
Lab module for simplified R&D iteration on investment models.

This module provides a clean interface for:
1. Creating/managing investment models
2. Generating entries (optimized bids)
3. Recording evaluation results
4. Comparing model performance
5. Portfolio optimization research (maxmin)

The lab schema replaces the over-complicated derived.algorithms,
derived.candidates, and derived.suite_* tables with just 3 tables:
- lab.investment_models
- lab.entries
- lab.evaluations

For production portfolio optimization, use the Go DP allocator in
backend/internal/app/recommended_entry_bids/allocator.go which provides
exact optimal solutions.
"""
