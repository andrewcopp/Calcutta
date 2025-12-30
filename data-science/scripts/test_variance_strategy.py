"""
Quick test to verify variance-aware strategy works correctly.
"""
import sys
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))

from moneyball.cli import investment_report


def test_variance_strategy():
    """Test that variance_aware_light strategy runs without errors."""
    print("Testing variance_aware_light strategy on 2024 data...")
    
    try:
        result = investment_report(
            snapshot_path="out/2024",
            snapshot_name="2024_variance_test",
            n_sims=100,  # Small number for quick test
            seed=123,
            budget_points=100,
            strategy="variance_aware_light",
        )
        
        print("\n✓ Strategy ran successfully!")
        print(f"\nResults:")
        print(f"  Portfolio teams: {result.get('portfolio_team_count', 'N/A')}")
        print(f"  Mean payout: ${result.get('mean_expected_payout_cents', 0)/100:.2f}")
        print(f"  Mean points: {result.get('mean_expected_points', 0):.1f}")
        print(f"  P(top 1): {result.get('p_top1', 0)*100:.1f}%")
        print(f"  P(in money): {result.get('p_in_money', 0)*100:.1f}%")
        
        return True
        
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = test_variance_strategy()
    sys.exit(0 if success else 1)
