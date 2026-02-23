import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Modal } from './Modal';

describe('Modal', () => {
  it('renders nothing when open is false', () => {
    // GIVEN open is false
    // WHEN rendering Modal
    const { container } = render(
      <Modal open={false} onClose={() => {}}>
        Content
      </Modal>,
    );

    // THEN nothing is rendered
    expect(container.innerHTML).toBe('');
  });

  it('renders children when open is true', () => {
    // GIVEN open is true
    // WHEN rendering Modal
    render(
      <Modal open={true} onClose={() => {}}>
        Modal content
      </Modal>,
    );

    // THEN children are displayed
    expect(screen.getByText('Modal content')).toBeInTheDocument();
  });

  it('renders the title when provided', () => {
    // GIVEN a title prop
    // WHEN rendering Modal
    render(
      <Modal open={true} onClose={() => {}} title="My Title">
        Content
      </Modal>,
    );

    // THEN the title is displayed
    expect(screen.getByText('My Title')).toBeInTheDocument();
  });

  it('calls onClose when Escape key is pressed', async () => {
    // GIVEN an open modal
    const onClose = vi.fn();
    render(
      <Modal open={true} onClose={onClose}>
        <button>Focus me</button>
      </Modal>,
    );

    // WHEN pressing Escape
    await userEvent.keyboard('{Escape}');

    // THEN onClose is called
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('calls onClose when clicking the overlay', async () => {
    // GIVEN an open modal
    const onClose = vi.fn();
    render(
      <Modal open={true} onClose={onClose}>
        Content
      </Modal>,
    );

    // WHEN clicking the overlay (the outer dialog wrapper)
    const overlay = screen.getByRole('dialog');
    await userEvent.click(overlay);

    // THEN onClose is called
    expect(onClose).toHaveBeenCalled();
  });

  it('does not call onClose when clicking inside the modal content', async () => {
    // GIVEN an open modal
    const onClose = vi.fn();
    render(
      <Modal open={true} onClose={onClose}>
        <p>Inner content</p>
      </Modal>,
    );

    // WHEN clicking inside the modal content
    await userEvent.click(screen.getByText('Inner content'));

    // THEN onClose is not called
    expect(onClose).not.toHaveBeenCalled();
  });

  it('traps focus within the modal on Tab', async () => {
    // GIVEN a modal with two focusable elements
    render(
      <Modal open={true} onClose={() => {}}>
        <button>First</button>
        <button>Last</button>
      </Modal>,
    );

    // Focus the last button
    screen.getByText('Last').focus();

    // WHEN pressing Tab from the last element
    await userEvent.tab();

    // THEN focus wraps to the first element
    expect(screen.getByText('First')).toHaveFocus();
  });

  it('traps focus within the modal on Shift+Tab', async () => {
    // GIVEN a modal with two focusable elements
    render(
      <Modal open={true} onClose={() => {}}>
        <button>First</button>
        <button>Last</button>
      </Modal>,
    );

    // Focus the first button
    screen.getByText('First').focus();

    // WHEN pressing Shift+Tab from the first element
    await userEvent.tab({ shift: true });

    // THEN focus wraps to the last element
    expect(screen.getByText('Last')).toHaveFocus();
  });
});
