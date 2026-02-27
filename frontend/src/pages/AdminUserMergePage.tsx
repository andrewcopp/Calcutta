import { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  getFilteredRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
  type RowSelectionState,
} from '@tanstack/react-table';
import { adminService } from '../services/adminService';
import type { AdminUserListItem } from '../schemas/admin';
import { toast } from '../lib/toast';
import { queryKeys } from '../queryKeys';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Badge } from '../components/ui/Badge';
import { Modal, ModalActions } from '../components/ui/Modal';
import { Alert } from '../components/ui/Alert';
import { Input } from '../components/ui/Input';
import { formatDate } from '../utils/format';

const PAGE_SIZE = 25;

const columns: ColumnDef<AdminUserListItem, unknown>[] = [
  {
    id: 'select',
    header: ({ table }) => (
      <input
        type="checkbox"
        checked={table.getIsAllPageRowsSelected()}
        onChange={table.getToggleAllPageRowsSelectedHandler()}
        className="h-4 w-4 rounded border-border"
      />
    ),
    cell: ({ row }) => (
      <input
        type="checkbox"
        checked={row.getIsSelected()}
        onChange={row.getToggleSelectedHandler()}
        className="h-4 w-4 rounded border-border"
      />
    ),
    enableSorting: false,
    enableGlobalFilter: false,
  },
  {
    accessorFn: (row) => `${row.firstName} ${row.lastName}`,
    id: 'name',
    header: 'Name',
  },
  {
    accessorKey: 'email',
    header: 'Email',
    cell: ({ row }) =>
      row.original.email ?? <span className="text-muted-foreground/60 italic">No email</span>,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => {
      const status = row.original.status;
      const variant = status === 'active' ? 'success' : status === 'stub' ? 'secondary' : 'default';
      return <Badge variant={variant}>{status}</Badge>;
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: ({ row }) => <span className="whitespace-nowrap">{formatDate(row.original.createdAt)}</span>,
  },
];

export function AdminUserMergePage() {
  const queryClient = useQueryClient();
  const [globalFilter, setGlobalFilter] = useState('');
  const [sorting, setSorting] = useState<SortingState>([]);
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const [showMergeModal, setShowMergeModal] = useState(false);
  const [primaryUserId, setPrimaryUserId] = useState<string | null>(null);
  const [merging, setMerging] = useState(false);
  const [mergeError, setMergeError] = useState('');

  const usersQuery = useQuery({
    queryKey: queryKeys.admin.users(),
    queryFn: () => adminService.listUsers(),
  });

  const users = useMemo(() => usersQuery.data?.items ?? [], [usersQuery.data]);

  const table = useReactTable({
    data: users,
    columns,
    state: { sorting, globalFilter, rowSelection },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    onRowSelectionChange: setRowSelection,
    enableRowSelection: true,
    getRowId: (row) => row.id,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    initialState: { pagination: { pageSize: PAGE_SIZE } },
  });

  const selectedRows = table.getSelectedRowModel().rows;
  const selectedUsers = selectedRows.map((r) => r.original);
  const selectedCount = selectedUsers.length;

  const openMergeModal = () => {
    setPrimaryUserId(null);
    setMergeError('');
    setShowMergeModal(true);
  };

  const handleMergeConfirm = async () => {
    if (!primaryUserId || selectedCount < 2) return;
    setMergeError('');
    setMerging(true);
    try {
      const sourceIds = selectedUsers.filter((u) => u.id !== primaryUserId).map((u) => u.id);
      const result = await adminService.batchMergeUsers(sourceIds, primaryUserId);
      const totalPortfolios = result.merges.reduce((s, m) => s + m.entriesMoved, 0);
      const totalInvitations = result.merges.reduce((s, m) => s + m.invitationsMoved, 0);
      const totalGrants = result.merges.reduce((s, m) => s + m.grantsMoved, 0);
      toast.success(
        `Merged ${result.merges.length} users: ${totalPortfolios} portfolios, ${totalInvitations} invitations, ${totalGrants} grants moved.`,
      );
      setShowMergeModal(false);
      setRowSelection({});
      await queryClient.invalidateQueries({ queryKey: queryKeys.admin.users() });
    } catch (err: unknown) {
      setMergeError(err instanceof Error ? err.message : 'Merge failed. Please try again.');
    } finally {
      setMerging(false);
    }
  };

  const primaryUser = selectedUsers.find((u) => u.id === primaryUserId);

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Admin', href: '/admin' }, { label: 'User Merge' }]} />
      <PageHeader
        title="Admin: User Merge"
        subtitle="Select two or more users to merge into a single account."
        actions={
          <Link to="/admin" className="text-primary hover:text-primary">
            Back to Admin Console
          </Link>
        }
      />

      <div className="mb-4">
        <Input
          placeholder="Search by name, email, or status..."
          value={globalFilter}
          onChange={(e) => setGlobalFilter(e.target.value)}
          className="max-w-sm"
        />
      </div>

      {usersQuery.isError && <ErrorState error={usersQuery.error} onRetry={() => usersQuery.refetch()} />}
      {usersQuery.isLoading && <LoadingState label="Loading users..." layout="inline" />}

      {!usersQuery.isLoading && (
        <div className="overflow-x-auto rounded-lg border border-border">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-accent">
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      className={`px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider ${
                        header.column.getCanSort() ? 'cursor-pointer select-none hover:text-foreground' : ''
                      }`}
                      onClick={header.column.getToggleSortingHandler()}
                    >
                      <div className="flex items-center gap-1">
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                        {header.column.getCanSort() && (
                          <span className="text-muted-foreground/60">
                            {{ asc: '\u2191', desc: '\u2193' }[header.column.getIsSorted() as string] ?? '\u2195'}
                          </span>
                        )}
                      </div>
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {table.getRowModel().rows.map((row) => (
                <tr
                  key={row.id}
                  className={`hover:bg-accent ${row.getIsSelected() ? 'bg-primary/5' : ''}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3 text-sm text-foreground">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
              {table.getRowModel().rows.length === 0 && (
                <tr>
                  <td colSpan={columns.length} className="px-4 py-8 text-center text-muted-foreground text-sm">
                    No users found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>

          {table.getPageCount() > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t border-border">
              <div className="text-sm text-muted-foreground">
                Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()} ({users.length} users)
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => table.previousPage()}
                  disabled={!table.getCanPreviousPage()}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => table.nextPage()}
                  disabled={!table.getCanNextPage()}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Bottom padding spacer when bar is visible */}
      {selectedCount >= 2 && <div className="h-16" />}

      {/* Sticky bottom merge bar */}
      {selectedCount >= 2 && (
        <div className="fixed bottom-0 left-0 right-0 bg-card border-t border-border px-6 py-3 flex items-center justify-between z-50 shadow-lg">
          <span className="text-sm font-medium">{selectedCount} users selected</span>
          <Button onClick={openMergeModal}>Merge {selectedCount} Users</Button>
        </div>
      )}

      {/* Merge modal */}
      {showMergeModal && (
        <Modal
          open={showMergeModal}
          onClose={() => {
            setShowMergeModal(false);
            setMergeError('');
          }}
          title="Select Primary User"
        >
          <p className="text-sm text-muted-foreground mb-4">
            All other users will be merged into the selected primary account. Their portfolios,
            invitations, and grants will be transferred, and they will be soft-deleted.
          </p>

          <div className="space-y-2 mb-4">
            {selectedUsers.map((u) => (
              <label
                key={u.id}
                className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer ${
                  primaryUserId === u.id ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50'
                }`}
              >
                <input
                  type="radio"
                  name="primaryUser"
                  checked={primaryUserId === u.id}
                  onChange={() => setPrimaryUserId(u.id)}
                  className="h-4 w-4"
                />
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium">
                    {u.firstName} {u.lastName}
                  </div>
                  <div className="text-xs text-muted-foreground truncate">
                    {u.email ?? 'No email'} &middot; {u.status}
                  </div>
                </div>
              </label>
            ))}
          </div>

          {mergeError && (
            <Alert variant="error" className="mb-4">
              {mergeError}
            </Alert>
          )}

          <ModalActions>
            <Button
              variant="secondary"
              onClick={() => {
                setShowMergeModal(false);
                setMergeError('');
              }}
              disabled={merging}
            >
              Cancel
            </Button>
            <Button onClick={handleMergeConfirm} disabled={merging || !primaryUserId}>
              {merging
                ? 'Merging...'
                : primaryUser
                  ? `Merge into ${primaryUser.firstName} ${primaryUser.lastName}`
                  : 'Select a primary user'}
            </Button>
          </ModalActions>
        </Modal>
      )}
    </PageContainer>
  );
}
