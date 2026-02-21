"""
Unit tests for _build_team_dataset_query() in moneyball.db.readers_ridge.

Tests the SQL string construction logic as a pure function, verifying
that CTEs, joins, select columns, and parameter placeholders are
conditionally included based on include_target and exclude_clause.
"""

from __future__ import annotations

from moneyball.db.readers_ridge import _build_team_dataset_query


class TestThatBuildTeamDatasetQueryIncludesTeamBidsCteWhenTargetEnabled:
    """When include_target=True the query must contain the team_bids CTE."""

    def test_that_team_bids_cte_is_present_when_include_target_is_true(self) -> None:
        # GIVEN include_target is True
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True)

        # THEN the SQL contains the team_bids CTE
        assert "team_bids AS" in sql


class TestThatBuildTeamDatasetQueryIncludesTotalCteWhenTargetEnabled:
    """When include_target=True the query must contain the total CTE."""

    def test_that_total_cte_is_present_when_include_target_is_true(self) -> None:
        # GIVEN include_target is True
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True)

        # THEN the SQL contains the total CTE
        assert "total AS" in sql


class TestThatBuildTeamDatasetQueryIncludesTargetColumnWhenTargetEnabled:
    """When include_target=True the query must select observed_team_share_of_pool."""

    def test_that_observed_team_share_of_pool_column_is_present(self) -> None:
        # GIVEN include_target is True
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True)

        # THEN the SQL selects the target column
        assert "observed_team_share_of_pool" in sql


class TestThatBuildTeamDatasetQueryIncludesTeamBidsJoinWhenTargetEnabled:
    """When include_target=True the query must LEFT JOIN team_bids."""

    def test_that_left_join_team_bids_is_present(self) -> None:
        # GIVEN include_target is True
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True)

        # THEN the SQL contains a LEFT JOIN on team_bids
        assert "LEFT JOIN team_bids tb ON tb.team_id = tt.id" in sql


class TestThatBuildTeamDatasetQueryUsesCalcuttaKeyPlaceholderWhenTargetEnabled:
    """When include_target=True, calcutta_key uses a parameter placeholder."""

    def test_that_calcutta_key_uses_parameter_placeholder(self) -> None:
        # GIVEN include_target is True
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True)

        # THEN calcutta_key uses a parameterized placeholder, not NULL
        assert "%s::text AS calcutta_key" in sql


class TestThatBuildTeamDatasetQueryOmitsTeamBidsCteWhenTargetDisabled:
    """When include_target=False the query must not contain the team_bids CTE."""

    def test_that_team_bids_cte_is_absent_when_include_target_is_false(self) -> None:
        # GIVEN include_target is False
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=False)

        # THEN the SQL does not contain the team_bids CTE
        assert "team_bids" not in sql


class TestThatBuildTeamDatasetQueryOmitsTargetColumnWhenTargetDisabled:
    """When include_target=False the query must not select observed_team_share_of_pool."""

    def test_that_observed_team_share_of_pool_is_absent(self) -> None:
        # GIVEN include_target is False
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=False)

        # THEN the SQL does not select the target column
        assert "observed_team_share_of_pool" not in sql


class TestThatBuildTeamDatasetQueryUsesNullCalcuttaKeyWhenTargetDisabled:
    """When include_target=False, calcutta_key should be NULL::text."""

    def test_that_calcutta_key_is_null(self) -> None:
        # GIVEN include_target is False
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=False)

        # THEN calcutta_key is NULL
        assert "NULL::text AS calcutta_key" in sql


class TestThatBuildTeamDatasetQueryIncludesExcludeClause:
    """When an exclude_clause is provided, it appears in the CTE."""

    def test_that_exclude_clause_is_embedded_in_cte(self) -> None:
        # GIVEN an exclude clause that filters by entry name
        exclude_clause = " AND ce.name <> ALL(%s::text[]) "

        # WHEN building the query with the exclude clause
        sql = _build_team_dataset_query(
            include_target=True,
            exclude_clause=exclude_clause,
        )

        # THEN the exclude clause appears in the generated SQL
        assert "AND ce.name <> ALL(%s::text[])" in sql


class TestThatBuildTeamDatasetQueryOmitsExcludeClauseWhenEmpty:
    """When exclude_clause is empty, no extra filtering appears in the CTE."""

    def test_that_no_entry_name_filter_is_present(self) -> None:
        # GIVEN no exclude clause (default)
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True, exclude_clause="")

        # THEN the SQL does not contain the entry name exclusion filter
        assert "ce.name <>" not in sql


class TestThatBuildTeamDatasetQueryAlwaysContainsBaseColumns:
    """Regardless of include_target, the base SELECT columns are always present."""

    def test_that_snapshot_column_is_always_present(self) -> None:
        # GIVEN include_target is False
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=False)

        # THEN the SQL selects the snapshot column
        assert "AS snapshot" in sql

    def test_that_seed_column_is_always_present(self) -> None:
        # GIVEN include_target is True
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True)

        # THEN the SQL selects the seed column
        assert "tt.seed::int AS seed" in sql

    def test_that_kenpom_net_column_is_always_present(self) -> None:
        # GIVEN include_target is False
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=False)

        # THEN the SQL selects the kenpom_net column
        assert "AS kenpom_net" in sql


class TestThatBuildTeamDatasetQueryHasCorrectParameterPlaceholders:
    """The number of %s placeholders differs based on include_target."""

    def test_that_target_query_has_more_placeholders_than_non_target(self) -> None:
        # GIVEN both query variants
        # WHEN building both
        sql_with_target = _build_team_dataset_query(include_target=True)
        sql_without_target = _build_team_dataset_query(include_target=False)

        # THEN the target query has more parameter placeholders
        count_with = sql_with_target.count("%s")
        count_without = sql_without_target.count("%s")
        assert count_with > count_without

    def test_that_non_target_query_has_five_placeholders(self) -> None:
        # GIVEN include_target is False
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=False)

        # THEN there are exactly 5 parameter placeholders
        # (snapshot, tournament_key, tournament_id, tournament_key prefix, tournament_id WHERE)
        assert sql.count("%s") == 5

    def test_that_target_query_without_exclude_has_seven_placeholders(self) -> None:
        # GIVEN include_target is True with no exclude clause
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True, exclude_clause="")

        # THEN there are exactly 7 parameter placeholders:
        #   CTE: calcutta_id (1)
        #   SELECT: snapshot, tournament_key, calcutta_key, tournament_id, tournament_key prefix (5)
        #   WHERE: tournament_id (1)
        assert sql.count("%s") == 7

    def test_that_target_query_with_exclude_has_eight_placeholders(self) -> None:
        # GIVEN include_target is True with an exclude clause containing one placeholder
        exclude_clause = " AND ce.name <> ALL(%s::text[]) "

        # WHEN building the query
        sql = _build_team_dataset_query(
            include_target=True,
            exclude_clause=exclude_clause,
        )

        # THEN there are exactly 8 parameter placeholders (7 base + 1 from exclude)
        assert sql.count("%s") == 8


class TestThatBuildTeamDatasetQueryAlwaysOrdersBySeedThenName:
    """The ORDER BY clause is always seed ASC, name ASC regardless of variant."""

    def test_that_order_by_is_present_when_target_disabled(self) -> None:
        # GIVEN include_target is False
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=False)

        # THEN the SQL orders by seed then name
        assert "ORDER BY tt.seed ASC, s.name ASC" in sql

    def test_that_order_by_is_present_when_target_enabled(self) -> None:
        # GIVEN include_target is True
        # WHEN building the query
        sql = _build_team_dataset_query(include_target=True)

        # THEN the SQL orders by seed then name
        assert "ORDER BY tt.seed ASC, s.name ASC" in sql
