package models

// DefaultTotalPoolBudget is the fallback total pool budget used when
// the actual value cannot be computed from database entries.
// Roughly 42 entries * 100 budget_points per entry.
const DefaultTotalPoolBudget = 4200
