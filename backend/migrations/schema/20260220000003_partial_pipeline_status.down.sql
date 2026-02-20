UPDATE lab.pipeline_runs SET status = 'succeeded' WHERE status = 'partial';
ALTER TABLE lab.pipeline_runs DROP CONSTRAINT ck_lab_pipeline_runs_status;
ALTER TABLE lab.pipeline_runs ADD CONSTRAINT ck_lab_pipeline_runs_status
    CHECK (status = ANY (ARRAY['pending','running','succeeded','failed','cancelled']));
