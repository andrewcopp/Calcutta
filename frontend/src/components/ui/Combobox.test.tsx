import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Combobox } from './Combobox';

const defaultOptions = [
  { id: '1', label: 'Alpha' },
  { id: '2', label: 'Beta' },
  { id: '3', label: 'Gamma' },
];

function renderCombobox(overrides: Record<string, unknown> = {}) {
  const props = {
    options: defaultOptions,
    value: '',
    onChange: vi.fn(),
    onSelect: vi.fn(),
    ...overrides,
  };
  return { ...render(<Combobox {...props} />), ...props };
}

describe('Combobox', () => {
  it('renders an input with combobox role', () => {
    // GIVEN default props
    // WHEN rendering Combobox
    renderCombobox();

    // THEN the input has role="combobox"
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('sets aria-expanded to false when the listbox is closed', () => {
    // GIVEN a closed combobox
    // WHEN rendering
    renderCombobox();

    // THEN aria-expanded is false
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-expanded', 'false');
  });

  it('sets aria-autocomplete to list', () => {
    // GIVEN default props
    // WHEN rendering
    renderCombobox();

    // THEN aria-autocomplete is "list"
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-autocomplete', 'list');
  });

  it('sets aria-haspopup to listbox', () => {
    // GIVEN default props
    // WHEN rendering
    renderCombobox();

    // THEN aria-haspopup is "listbox"
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-haspopup', 'listbox');
  });

  it('opens the listbox on focus', async () => {
    // GIVEN a closed combobox
    renderCombobox();

    // WHEN focusing the input
    await userEvent.click(screen.getByRole('combobox'));

    // THEN the listbox appears
    expect(screen.getByRole('listbox')).toBeInTheDocument();
  });

  it('shows filtered options when typing', async () => {
    // GIVEN a combobox with options
    const onChange = vi.fn();
    renderCombobox({ onChange });

    // WHEN typing "Al"
    await userEvent.type(screen.getByRole('combobox'), 'Al');

    // THEN onChange is called for each character
    expect(onChange).toHaveBeenCalledWith('A');
  });

  it('selects an option when clicking it', async () => {
    // GIVEN an open combobox with options
    const onSelect = vi.fn();
    const onChange = vi.fn();
    renderCombobox({ onSelect, onChange });

    // Open the listbox
    await userEvent.click(screen.getByRole('combobox'));

    // WHEN clicking "Beta"
    await userEvent.click(screen.getByText('Beta'));

    // THEN onSelect is called with the option ID
    expect(onSelect).toHaveBeenCalledWith('2');
  });

  it('navigates options with ArrowDown key', async () => {
    // GIVEN an open combobox
    renderCombobox();
    await userEvent.click(screen.getByRole('combobox'));

    // WHEN pressing ArrowDown
    await userEvent.keyboard('{ArrowDown}');

    // THEN the second option is highlighted (aria-selected)
    const options = screen.getAllByRole('option');
    expect(options[1]).toHaveAttribute('aria-selected', 'true');
  });

  it('navigates options with ArrowUp key', async () => {
    // GIVEN an open combobox with second item highlighted
    renderCombobox();
    await userEvent.click(screen.getByRole('combobox'));
    await userEvent.keyboard('{ArrowDown}');

    // WHEN pressing ArrowUp
    await userEvent.keyboard('{ArrowUp}');

    // THEN the first option is highlighted
    const options = screen.getAllByRole('option');
    expect(options[0]).toHaveAttribute('aria-selected', 'true');
  });

  it('selects highlighted option with Enter key', async () => {
    // GIVEN an open combobox with first item highlighted
    const onSelect = vi.fn();
    renderCombobox({ onSelect });
    await userEvent.click(screen.getByRole('combobox'));

    // WHEN pressing Enter
    await userEvent.keyboard('{Enter}');

    // THEN the first option is selected
    expect(onSelect).toHaveBeenCalledWith('1');
  });

  it('closes the listbox on Escape', async () => {
    // GIVEN an open combobox
    renderCombobox();
    await userEvent.click(screen.getByRole('combobox'));
    expect(screen.getByRole('listbox')).toBeInTheDocument();

    // WHEN pressing Escape
    await userEvent.keyboard('{Escape}');

    // THEN the listbox is closed
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('excludes options in the excludeIds set', async () => {
    // GIVEN excludeIds containing option "2"
    renderCombobox({ excludeIds: new Set(['2']) });
    await userEvent.click(screen.getByRole('combobox'));

    // THEN "Beta" is not in the list
    expect(screen.queryByText('Beta')).not.toBeInTheDocument();
  });

  it('shows "No results found" when no options match', async () => {
    // GIVEN a combobox with value that matches nothing
    renderCombobox({ value: 'zzz' });
    await userEvent.click(screen.getByRole('combobox'));

    // THEN the no results message is shown
    expect(screen.getByText('No results found')).toBeInTheDocument();
  });
});
