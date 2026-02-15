package lab

import (
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// Type aliases for backwards compatibility. Canonical types live in internal/models.
type InvestmentModel = models.InvestmentModel
type Entry = models.LabEntry
type EntryDetail = models.LabEntryDetail
type Prediction = models.LabPrediction
type EnrichedPrediction = models.LabEnrichedPrediction
type EntryBid = models.LabEntryBid
type EnrichedBid = models.LabEnrichedBid
type EntryDetailEnriched = models.LabEntryDetailEnriched
type Evaluation = models.LabEvaluation
type EvaluationDetail = models.LabEvaluationDetail
type EvaluationEntryResult = models.LabEvaluationEntryResult
type EvaluationEntryBid = models.LabEvaluationEntryBid
type EvaluationEntryProfile = models.LabEvaluationEntryProfile
type LeaderboardEntry = models.LabLeaderboardEntry
type ListModelsFilter = models.LabListModelsFilter
type ListEntriesFilter = models.LabListEntriesFilter
type ListEvaluationsFilter = models.LabListEvaluationsFilter
type Pagination = models.LabPagination
type GenerateEntriesRequest = models.LabGenerateEntriesRequest
type GenerateEntriesResponse = models.LabGenerateEntriesResponse

// Interface aliases for backwards compatibility. Canonical interfaces live in internal/ports.
type Repository = ports.LabRepository
type PipelineRepository = ports.LabPipelineRepository
