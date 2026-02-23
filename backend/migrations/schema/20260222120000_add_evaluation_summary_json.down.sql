SET search_path = '';

ALTER TABLE lab.evaluations DROP COLUMN IF EXISTS summary_json;
