package httpserver

import "github.com/gorilla/mux"

func (s *Server) registerSuiteCalcuttaEvaluationRoutes(r *mux.Router) {
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulations",
		s.requirePermission("analytics.suite_calcutta_evaluations.write", s.createCohortSimulationHandler),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulations",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.listCohortSimulationsHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulations/{id}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getCohortSimulationHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulations/{id}/result",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getCohortSimulationResultHandler),
	).Methods("GET", "OPTIONS")
	r.HandleFunc(
		"/api/cohorts/{cohortId}/simulations/{id}/entries/{snapshotEntryId}",
		s.requirePermission("analytics.suite_calcutta_evaluations.read", s.getCohortSimulationSnapshotEntryHandler),
	).Methods("GET", "OPTIONS")
}
