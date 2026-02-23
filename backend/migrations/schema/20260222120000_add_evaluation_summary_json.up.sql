SET search_path = '';

ALTER TABLE lab.evaluations ADD COLUMN summary_json jsonb;
