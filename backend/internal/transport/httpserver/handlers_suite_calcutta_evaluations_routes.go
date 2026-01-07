package httpserver

import "github.com/gorilla/mux"

func (s *Server) registerSuiteCalcuttaEvaluationRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/suite-calcutta-evaluations",
		s.requirePermission("analytics.suite_calcutta_evaluations.write", s.createSuiteCalcuttaEvaluationHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.listSuiteCalcuttaEvaluationsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations/{id}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations/{id}/result",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationResultHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/suite-calcutta-evaluations/{id}/entries/{snapshotEntryId}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getSuiteCalcuttaEvaluationSnapshotEntryHandler),
	).Methods("GET", "OPTIONS")
}
