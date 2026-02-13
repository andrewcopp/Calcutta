import React from 'react';
import { cn } from '../../lib/cn';

type PipelineProgressBarProps = {
  hasEntries: boolean;
  hasEvaluations: boolean;
  className?: string;
};

type StageStatus = 'complete' | 'pending' | 'inactive';

type StageConfig = {
  bgClass: string;
  borderClass: string;
  iconColor: string;
  lineColor: string;
};

const stageConfigs: Record<StageStatus, StageConfig> = {
  complete: {
    bgClass: 'bg-green-500',
    borderClass: 'border-green-500',
    iconColor: 'text-white',
    lineColor: 'bg-green-500',
  },
  pending: {
    bgClass: 'bg-white',
    borderClass: 'border-amber-500',
    iconColor: 'text-amber-500',
    lineColor: 'bg-gray-300',
  },
  inactive: {
    bgClass: 'bg-white',
    borderClass: 'border-gray-300',
    iconColor: 'text-gray-300',
    lineColor: 'bg-gray-300',
  },
};

function getStageStatus(isComplete: boolean, isPending: boolean): StageStatus {
  if (isComplete) return 'complete';
  if (isPending) return 'pending';
  return 'inactive';
}

function StageNode({
  status,
  label,
  ariaLabel,
}: {
  status: StageStatus;
  label: string;
  ariaLabel: string;
}) {
  const config = stageConfigs[status];

  return (
    <div className="flex flex-col items-center" aria-label={ariaLabel}>
      <div
        className={cn(
          'w-6 h-6 rounded-full border-2 flex items-center justify-center',
          config.bgClass,
          config.borderClass
        )}
      >
        {status === 'complete' ? (
          <svg className={cn('w-3.5 h-3.5', config.iconColor)} fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        ) : (
          <div
            className={cn(
              'w-2 h-2 rounded-full',
              status === 'pending' ? 'bg-amber-500' : 'bg-gray-300'
            )}
          />
        )}
      </div>
      <span className="text-[10px] text-gray-500 mt-1 whitespace-nowrap">{label}</span>
    </div>
  );
}

function ConnectingLine({ status }: { status: StageStatus }) {
  const config = stageConfigs[status];
  return <div className={cn('flex-1 h-0.5 mx-1', config.lineColor)} />;
}

export function PipelineProgressBar({ hasEntries, hasEvaluations, className }: PipelineProgressBarProps) {
  // Stage 1: Model Registered - always true (model exists)
  const stage1Status: StageStatus = 'complete';

  // Stage 2: Entries Generated - pending if no entries, complete if has entries
  const stage2Status = getStageStatus(hasEntries, !hasEntries);

  // Stage 3: Evaluated - pending if has entries but no evaluations, complete if has evaluations
  const stage3Status = getStageStatus(hasEvaluations, hasEntries && !hasEvaluations);

  // Line statuses - line is complete only if the next stage is complete
  const line1Status: StageStatus = hasEntries ? 'complete' : 'inactive';
  const line2Status: StageStatus = hasEvaluations ? 'complete' : 'inactive';

  // Calculate progress for aria
  const stagesCompleted = 1 + (hasEntries ? 1 : 0) + (hasEvaluations ? 1 : 0);

  return (
    <div
      className={cn('flex items-start', className)}
      role="progressbar"
      aria-valuenow={stagesCompleted}
      aria-valuemin={0}
      aria-valuemax={3}
      aria-label={`Pipeline progress: ${stagesCompleted} of 3 stages complete`}
    >
      <StageNode status={stage1Status} label="Registered" ariaLabel="Model registered: complete" />
      <ConnectingLine status={line1Status} />
      <StageNode status={stage2Status} label="Entries" ariaLabel={`Entries generated: ${stage2Status}`} />
      <ConnectingLine status={line2Status} />
      <StageNode status={stage3Status} label="Evaluated" ariaLabel={`Performance evaluated: ${stage3Status}`} />
    </div>
  );
}
