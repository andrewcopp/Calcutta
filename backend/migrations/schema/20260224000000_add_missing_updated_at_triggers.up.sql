CREATE TRIGGER trg_compute_prediction_batches_updated_at
  BEFORE UPDATE ON compute.prediction_batches
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();

CREATE TRIGGER trg_compute_predicted_team_values_updated_at
  BEFORE UPDATE ON compute.predicted_team_values
  FOR EACH ROW EXECUTE FUNCTION core.set_updated_at();
