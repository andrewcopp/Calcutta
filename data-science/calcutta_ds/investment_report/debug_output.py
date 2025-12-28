from __future__ import annotations

from typing import Dict, List

import pandas as pd


def build_debug_output(
    *,
    usable: pd.DataFrame,
    chosen: pd.DataFrame,
    market_bids: pd.DataFrame,
    min_bid: float,
    n_entries: int,
    predicted_total_pool_bids: float,
) -> Dict[str, object]:
    market_totals = (
        market_bids.groupby(["team_key"], dropna=False)
        .agg(market_total_team_bids=("bid_amount", "sum"))
        .reset_index()
    )
    market_totals["market_total_team_bids"] = pd.to_numeric(
        market_totals["market_total_team_bids"],
        errors="coerce",
    ).fillna(0.0)

    debug_df = usable.copy()
    debug_df = debug_df.merge(market_totals, on="team_key", how="left")
    debug_df["market_total_team_bids"] = pd.to_numeric(
        debug_df.get("market_total_team_bids"),
        errors="coerce",
    ).fillna(0.0)

    if "expected_team_points" in debug_df.columns:
        total_exp = float(
            pd.to_numeric(
                debug_df["expected_team_points"],
                errors="coerce",
            )
            .fillna(0.0)
            .sum()
        )
        if total_exp > 0:
            debug_df["expected_points_share"] = (
                pd.to_numeric(
                    debug_df["expected_team_points"],
                    errors="coerce",
                )
                .fillna(0.0)
                / total_exp
            )
        else:
            debug_df["expected_points_share"] = 0.0

        debug_df["bid_share"] = pd.to_numeric(
            debug_df["predicted_team_share_of_pool"],
            errors="coerce",
        ).fillna(0.0)

        debug_df["value_ratio"] = debug_df.apply(
            lambda r: (
                float(r.get("expected_points_share") or 0.0)
                / float(r.get("bid_share") or 0.0)
                if float(r.get("bid_share") or 0.0) > 0
                else 0.0
            ),
            axis=1,
        )

        debug_df["implied_ownership_at_min_bid"] = debug_df.apply(
            lambda r: (
                float(min_bid)
                / (
                    float(r.get("market_total_team_bids") or 0.0)
                    + float(min_bid)
                )
                if (
                    float(r.get("market_total_team_bids") or 0.0)
                    + float(min_bid)
                )
                > 0
                else 0.0
            ),
            axis=1,
        )
        debug_df["implied_points_at_min_bid"] = debug_df.apply(
            lambda r: float(r.get("expected_team_points") or 0.0)
            * float(r.get("implied_ownership_at_min_bid") or 0.0),
            axis=1,
        )

        if "predicted_team_total_bids" in debug_df.columns:
            debug_df["predicted_total_team_bids"] = pd.to_numeric(
                debug_df.get("predicted_team_total_bids"),
                errors="coerce",
            ).fillna(0.0)

            debug_df["predicted_ownership_at_min_bid"] = debug_df.apply(
                lambda r: (
                    float(min_bid)
                    / (
                        float(r.get("predicted_total_team_bids") or 0.0)
                        + float(min_bid)
                    )
                    if (
                        float(r.get("predicted_total_team_bids") or 0.0)
                        + float(min_bid)
                    )
                    > 0
                    else 0.0
                ),
                axis=1,
            )
            debug_df["predicted_points_at_min_bid"] = debug_df.apply(
                lambda r: float(r.get("expected_team_points") or 0.0)
                * float(r.get("predicted_ownership_at_min_bid") or 0.0),
                axis=1,
            )
            debug_df["predicted_points_per_dollar_at_min_bid"] = debug_df.apply(
                lambda r: (
                    float(r.get("predicted_points_at_min_bid") or 0.0)
                    / float(min_bid)
                    if float(min_bid) > 0
                    else 0.0
                ),
                axis=1,
            )

        def _own(m: float, b: float) -> float:
            denom = float(m) + float(b)
            if denom <= 0:
                return 0.0
            return float(b) / denom

        def _marginal_own(m: float, b: float) -> float:
            base = _own(m, b)
            after = _own(m, float(b) + 1.0)
            return float(after - base)

        for b in [0.0, 1.0, 5.0, 10.0, 25.0]:
            key = f"marginal_points_per_plus1_at_bid_{int(b)}"
            debug_df[key] = debug_df.apply(
                lambda r: float(r.get("expected_team_points") or 0.0)
                * _marginal_own(
                    float(r.get("market_total_team_bids") or 0.0),
                    float(b),
                ),
                axis=1,
            )

    portfolio_debug = chosen.copy()
    portfolio_debug = portfolio_debug.merge(
        market_totals,
        on="team_key",
        how="left",
    )
    portfolio_debug["market_total_team_bids"] = pd.to_numeric(
        portfolio_debug.get("market_total_team_bids"),
        errors="coerce",
    ).fillna(0.0)
    portfolio_debug["final_total_team_bids"] = (
        portfolio_debug["market_total_team_bids"]
        + pd.to_numeric(
            portfolio_debug.get("bid_amount"),
            errors="coerce",
        ).fillna(0.0)
    )
    portfolio_debug["final_ownership"] = portfolio_debug.apply(
        lambda r: (
            float(r.get("bid_amount") or 0.0)
            / float(r.get("final_total_team_bids") or 0.0)
            if float(r.get("final_total_team_bids") or 0.0) > 0
            else 0.0
        ),
        axis=1,
    )

    if "predicted_team_total_bids" in portfolio_debug.columns:
        portfolio_debug["predicted_total_team_bids"] = pd.to_numeric(
            portfolio_debug["predicted_team_total_bids"],
            errors="coerce",
        ).fillna(0.0)

        def _predicted_ownership_row(r: pd.Series) -> float:
            b = float(r.get("bid_amount") or 0.0)
            m = float(r.get("predicted_total_team_bids") or 0.0)
            denom = m + b
            return (b / denom) if denom > 0 else 0.0

        pred_key = "predicted_ownership"
        portfolio_debug[pred_key] = portfolio_debug.apply(
            _predicted_ownership_row,
            axis=1,
        )

        def _rvp_ratio_row(r: pd.Series) -> float:
            f = float(r.get("final_ownership") or 0.0)
            p = float(r.get("predicted_ownership") or 0.0)
            return (f / p) if p > 0 else 0.0

        ratio_key = "realized_vs_predicted_ownership_ratio"
        portfolio_debug[ratio_key] = portfolio_debug.apply(
            _rvp_ratio_row,
            axis=1,
        )

    if "expected_team_points" in portfolio_debug.columns:
        portfolio_debug["expected_points_contribution"] = (
            portfolio_debug.apply(
                lambda r: float(r.get("expected_team_points") or 0.0)
                * float(r.get("final_ownership") or 0.0),
                axis=1,
            )
        )

    def _cols(df: pd.DataFrame, cols: List[str]) -> List[str]:
        return [c for c in cols if c in df.columns]

    def _ranked(
        *,
        df: pd.DataFrame,
        by: str,
        n: int,
        ascending: bool,
        cols: List[str],
    ) -> List[Dict[str, object]]:
        if by not in df.columns:
            return []
        out = df.copy()
        out[by] = pd.to_numeric(out[by], errors="coerce")
        out = out[out[by].notna()].copy()
        if out.empty:
            return []
        out = out.sort_values(by=by, ascending=bool(ascending))
        return out.head(int(n))[_cols(out, cols)].to_dict(orient="records")

    base_cols = [
        "team_key",
        "school_name",
        "seed",
        "region",
        "kenpom_net",
        "predicted_team_share_of_pool",
        "predicted_team_total_bids",
        "market_total_team_bids",
        "expected_team_points",
        "expected_points_share",
        "value_score",
        "value_ratio",
        "predicted_ownership_at_min_bid",
        "predicted_points_at_min_bid",
        "predicted_points_per_dollar_at_min_bid",
        "implied_ownership_at_min_bid",
        "implied_points_at_min_bid",
        "marginal_points_per_plus1_at_bid_0",
        "marginal_points_per_plus1_at_bid_1",
        "marginal_points_per_plus1_at_bid_5",
        "marginal_points_per_plus1_at_bid_10",
        "marginal_points_per_plus1_at_bid_25",
    ]

    return {
        "n_entries": int(n_entries),
        "predicted_total_pool_bids": float(predicted_total_pool_bids),
        "market_totals_by_team": market_totals.to_dict(orient="records"),
        "usable_row_order_preview": debug_df.head(30)[
            _cols(
                debug_df,
                [
                    "team_key",
                    "school_name",
                    "seed",
                    "predicted_team_share_of_pool",
                    "value_score",
                    "market_total_team_bids",
                ],
            )
        ].to_dict(orient="records"),
        "ranked": {
            "by_value_score": _ranked(
                df=debug_df,
                by="value_score",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_predicted_points_per_dollar_at_min_bid": _ranked(
                df=debug_df,
                by="predicted_points_per_dollar_at_min_bid",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_predicted_team_share_of_pool": _ranked(
                df=debug_df,
                by="predicted_team_share_of_pool",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_market_total_team_bids": _ranked(
                df=debug_df,
                by="market_total_team_bids",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_expected_team_points": _ranked(
                df=debug_df,
                by="expected_team_points",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_implied_points_at_min_bid": _ranked(
                df=debug_df,
                by="implied_points_at_min_bid",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_marginal_points_per_plus1_at_bid_0": _ranked(
                df=debug_df,
                by="marginal_points_per_plus1_at_bid_0",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_marginal_points_per_plus1_at_bid_10": _ranked(
                df=debug_df,
                by="marginal_points_per_plus1_at_bid_10",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
            "by_value_ratio": _ranked(
                df=debug_df,
                by="value_ratio",
                n=25,
                ascending=False,
                cols=base_cols,
            ),
        },
        "team_table": debug_df[
            _cols(
                debug_df,
                [
                    "team_key",
                    "school_name",
                    "seed",
                    "region",
                    "kenpom_net",
                    "predicted_team_share_of_pool",
                    "value_score",
                    "market_total_team_bids",
                    "predicted_team_total_bids",
                    "expected_team_points",
                    "expected_points_share",
                    "bid_share",
                    "value_ratio",
                    "predicted_ownership_at_min_bid",
                    "predicted_points_at_min_bid",
                    "predicted_points_per_dollar_at_min_bid",
                    "implied_ownership_at_min_bid",
                    "implied_points_at_min_bid",
                ],
            )
        ].to_dict(orient="records"),
        "portfolio_table": portfolio_debug[
            _cols(
                portfolio_debug,
                [
                    "team_key",
                    "school_name",
                    "seed",
                    "bid_amount",
                    "market_total_team_bids",
                    "predicted_total_team_bids",
                    "final_total_team_bids",
                    "predicted_ownership",
                    "final_ownership",
                    "realized_vs_predicted_ownership_ratio",
                    "expected_team_points",
                    "expected_points_contribution",
                ],
            )
        ].to_dict(orient="records"),
    }
