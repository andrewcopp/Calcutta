import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Alert } from './Alert';

describe('Alert', () => {
  it('renders children text', () => {
    // GIVEN children content
    // WHEN rendering Alert
    render(<Alert>Hello world</Alert>);

    // THEN the children are displayed
    expect(screen.getByText('Hello world')).toBeInTheDocument();
  });

  it('applies role="alert" for error variant', () => {
    // GIVEN the error variant
    // WHEN rendering Alert
    render(<Alert variant="error">Error occurred</Alert>);

    // THEN role="alert" is present
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('does not apply role="alert" for info variant', () => {
    // GIVEN the info variant (default)
    // WHEN rendering Alert
    render(<Alert variant="info">Info message</Alert>);

    // THEN role="alert" is not present
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('does not apply role="alert" for success variant', () => {
    // GIVEN the success variant
    // WHEN rendering Alert
    render(<Alert variant="success">Success</Alert>);

    // THEN role="alert" is not present
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('does not apply role="alert" for warning variant', () => {
    // GIVEN the warning variant
    // WHEN rendering Alert
    render(<Alert variant="warning">Warning</Alert>);

    // THEN role="alert" is not present
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
