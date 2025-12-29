"""
Old algorithms from calcutta_ds preserved for testing and comparison.

These algorithms have been replaced by newer implementations in moneyball,
but are preserved here for:
- Baseline comparison testing
- Validating that new implementations outperform old ones
- Historical replay and backtesting
"""
from typing import List


def waterfill_equal(
    k: int,
    budget: float,
    max_per_team: float,
) -> List[float]:
    """
    Old allocation algorithm: equal distribution with max-per-team caps.
    
    Replaced by: Greedy optimizer in recommended_entry_bids.py
    
    Algorithm:
    - Initialize all bids to 0
    - Iterate through teams, incrementing each by 1 until budget exhausted
    - Skip teams that have reached max_per_team cap
    
    Use case: Baseline for testing that greedy optimizer outperforms
    naive equal allocation.
    """
    if k <= 0:
        return []
    if budget < 0:
        raise ValueError("budget must be non-negative")
    if max_per_team <= 0:
        raise ValueError("max_per_team must be positive")

    b_int = int(round(float(budget)))
    m_int = int(round(float(max_per_team)))
    if abs(float(budget) - float(b_int)) > 1e-9:
        raise ValueError("budget must be an integer number of dollars")
    if abs(float(max_per_team) - float(m_int)) > 1e-9:
        raise ValueError("max_per_team must be an integer number of dollars")
    if b_int > int(k) * int(m_int):
        raise ValueError("budget exceeds k * max_per_team")

    bids_int: List[int] = [0 for _ in range(int(k))]
    remaining = int(b_int)

    while remaining > 0:
        made_progress = False
        for i in range(int(k)):
            if remaining <= 0:
                break
            if bids_int[i] >= int(m_int):
                continue
            bids_int[i] += 1
            remaining -= 1
            made_progress = True
        if not made_progress:
            raise ValueError("cannot allocate budget under max_per_team caps")

    return [float(x) for x in bids_int]
