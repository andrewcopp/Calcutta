-- name: CreateInvestmentSnapshot :exec
INSERT INTO core.investment_snapshots (portfolio_id, changed_by, reason, investments)
VALUES ($1, $2, $3, $4);
