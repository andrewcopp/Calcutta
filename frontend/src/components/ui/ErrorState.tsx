import { Alert } from './Alert';
import { Button } from './Button';

type ErrorStateProps = {
  error: string;
  onRetry?: () => void;
};

export function ErrorState({ error, onRetry }: ErrorStateProps) {
  return (
    <div className="space-y-4">
      <Alert variant="error">{error}</Alert>
      {onRetry ? (
        <Button onClick={onRetry}>Retry</Button>
      ) : null}
    </div>
  );
}
