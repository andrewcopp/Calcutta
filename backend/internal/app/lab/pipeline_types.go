package lab

import "github.com/andrewcopp/Calcutta/backend/internal/models"

// Type aliases for backwards compatibility. Canonical types live in internal/models.
type PipelineRun = models.LabPipelineRun
type PipelineCalcuttaRun = models.LabPipelineCalcuttaRun
type StartPipelineRequest = models.LabStartPipelineRequest
type StartPipelineResponse = models.LabStartPipelineResponse
type CalcuttaProgressResponse = models.LabCalcuttaProgressResponse
type PipelineProgressSummary = models.LabPipelineProgressSummary
type PipelineProgressResponse = models.LabPipelineProgressResponse
type ModelPipelineProgress = models.LabModelPipelineProgress
