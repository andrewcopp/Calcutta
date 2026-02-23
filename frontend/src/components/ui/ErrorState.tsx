import { Alert } from './Alert';
import { Button } from './Button';

type ErrorStateProps = {
  error: unknown;
  onRetry?: () => void;
};

function formatError(error: unknown): string {
  if (typeof error === 'string') return error;
  if (error instanceof Error) return error.message;
  return 'An unexpected error occurred';
}

export function ErrorState({ error, onRetry }: ErrorStateProps) {
  return (
    <div className="space-y-4">
      <Alert variant="error">{formatError(error)}</Alert>
      {onRetry ? <Button onClick={onRetry}>Retry</Button> : null}
    </div>
  );
}
