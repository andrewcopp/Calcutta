-- Rollback lab schema creation

DROP VIEW IF EXISTS lab.entry_evaluations;
DROP VIEW IF EXISTS lab.model_leaderboard;

DROP TABLE IF EXISTS lab.evaluations;
DROP TABLE IF EXISTS lab.entries;
DROP TABLE IF EXISTS lab.investment_models;

DROP SCHEMA IF EXISTS lab;
