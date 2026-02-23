import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ErrorState } from './ErrorState';

describe('ErrorState', () => {
  it('renders a string error message', () => {
    // GIVEN a string error
    // WHEN rendering ErrorState
    render(<ErrorState error="Something went wrong" />);

    // THEN the string is displayed
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });

  it('renders an Error object message', () => {
    // GIVEN an Error object
    // WHEN rendering ErrorState
    render(<ErrorState error={new Error('Network failure')} />);

    // THEN the error message is displayed
    expect(screen.getByText('Network failure')).toBeInTheDocument();
  });

  it('renders a fallback message for unknown error types', () => {
    // GIVEN an unknown error type
    // WHEN rendering ErrorState
    render(<ErrorState error={42} />);

    // THEN the fallback message is displayed
    expect(screen.getByText('An unexpected error occurred')).toBeInTheDocument();
  });

  it('does not render a retry button when onRetry is not provided', () => {
    // GIVEN no onRetry callback
    // WHEN rendering ErrorState
    render(<ErrorState error="Oops" />);

    // THEN no retry button is shown
    expect(screen.queryByRole('button', { name: /retry/i })).not.toBeInTheDocument();
  });

  it('renders a retry button when onRetry is provided', () => {
    // GIVEN an onRetry callback
    // WHEN rendering ErrorState
    render(<ErrorState error="Oops" onRetry={() => {}} />);

    // THEN the retry button is visible
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
  });

  it('calls onRetry when the retry button is clicked', async () => {
    // GIVEN an onRetry callback
    const onRetry = vi.fn();
    render(<ErrorState error="Oops" onRetry={onRetry} />);

    // WHEN clicking the retry button
    await userEvent.click(screen.getByRole('button', { name: /retry/i }));

    // THEN the callback is invoked
    expect(onRetry).toHaveBeenCalledOnce();
  });
});
