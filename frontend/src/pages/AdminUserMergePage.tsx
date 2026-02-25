import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { adminService } from '../services/adminService';
import type { StubUser, MergeCandidate } from '../schemas/admin';
import { toast } from '../lib/toast';
import { queryKeys } from '../queryKeys';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';
import { Badge } from '../components/ui/Badge';
import { Modal, ModalActions } from '../components/ui/Modal';
import { Alert } from '../components/ui/Alert';
import { formatDate } from '../utils/format';

export function AdminUserMergePage() {
  const queryClient = useQueryClient();
  const [selectedStub, setSelectedStub] = useState<StubUser | null>(null);
  const [mergeTarget, setMergeTarget] = useState<MergeCandidate | null>(null);
  const [merging, setMerging] = useState(false);
  const [mergeError, setMergeError] = useState('');

  const stubsQuery = useQuery({
    queryKey: queryKeys.admin.stubs(),
    queryFn: () => adminService.listStubUsers(),
  });

  const candidatesQuery = useQuery({
    queryKey: queryKeys.admin.mergeCandidates(selectedStub?.id ?? ''),
    queryFn: () => adminService.findMergeCandidates(selectedStub!.id),
    enabled: !!selectedStub,
  });

  const stubs = stubsQuery.data?.items ?? [];
  const candidates = candidatesQuery.data?.items ?? [];

  const handleFindMatches = (stub: StubUser) => {
    setSelectedStub(stub);
    setMergeTarget(null);
  };

  const handleMergeConfirm = async () => {
    if (!selectedStub || !mergeTarget) return;
    setMergeError('');
    setMerging(true);
    try {
      const result = await adminService.mergeUsers(selectedStub.id, mergeTarget.id);
      toast.success(
        `Merged successfully: ${result.entriesMoved} entries, ${result.invitationsMoved} invitations, ${result.grantsMoved} grants moved.`,
      );
      setMergeTarget(null);
      setSelectedStub(null);
      await queryClient.invalidateQueries({ queryKey: queryKeys.admin.stubs() });
    } catch (err: unknown) {
      setMergeError(err instanceof Error ? err.message : 'Merge failed. Please try again.');
    } finally {
      setMerging(false);
    }
  };

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Admin', href: '/admin' }, { label: 'User Merge' }]} />
      <PageHeader
        title="Admin: User Merge"
        subtitle="Consolidate stub users created during historical imports with real user accounts."
        actions={
          <Link to="/admin" className="text-primary hover:text-primary">
            Back to Admin Console
          </Link>
        }
      />

      <Card>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Stub Users</h2>
          <Button onClick={() => stubsQuery.refetch()} disabled={stubsQuery.isFetching} variant="secondary">
            Refresh
          </Button>
        </div>

        {stubsQuery.isError && <ErrorState error={stubsQuery.error} onRetry={() => stubsQuery.refetch()} />}
        {stubsQuery.isLoading && <LoadingState label="Loading stub users..." layout="inline" />}

        <Table className="text-sm">
          <TableHead>
            <TableRow>
              <TableHeaderCell>Name</TableHeaderCell>
              <TableHeaderCell>Status</TableHeaderCell>
              <TableHeaderCell>Created</TableHeaderCell>
              <TableHeaderCell>Actions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {stubs.map((s) => (
              <TableRow key={s.id} className={selectedStub?.id === s.id ? 'bg-muted/50' : ''}>
                <TableCell className="whitespace-nowrap">
                  {s.firstName} {s.lastName}
                </TableCell>
                <TableCell>
                  <Badge variant="secondary">{s.status}</Badge>
                </TableCell>
                <TableCell className="whitespace-nowrap">{formatDate(s.createdAt)}</TableCell>
                <TableCell>
                  <Button
                    size="sm"
                    variant={selectedStub?.id === s.id ? 'primary' : 'outline'}
                    onClick={() => handleFindMatches(s)}
                  >
                    Find Matches
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {stubs.length === 0 && !stubsQuery.isLoading && (
              <TableRow>
                <TableCell className="text-muted-foreground" colSpan={4}>
                  No stub users found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </Card>

      {selectedStub && (
        <Card className="mt-6">
          <h2 className="text-xl font-semibold mb-4">
            Merge Candidates for {selectedStub.firstName} {selectedStub.lastName}
          </h2>

          {candidatesQuery.isError && (
            <ErrorState error={candidatesQuery.error} onRetry={() => candidatesQuery.refetch()} />
          )}
          {candidatesQuery.isLoading && <LoadingState label="Finding matches..." layout="inline" />}

          {!candidatesQuery.isLoading && candidates.length === 0 && (
            <p className="text-muted-foreground text-sm">No matching users found with the same name.</p>
          )}

          {candidates.length > 0 && (
            <Table className="text-sm">
              <TableHead>
                <TableRow>
                  <TableHeaderCell>Name</TableHeaderCell>
                  <TableHeaderCell>Email</TableHeaderCell>
                  <TableHeaderCell>Status</TableHeaderCell>
                  <TableHeaderCell>Created</TableHeaderCell>
                  <TableHeaderCell>Actions</TableHeaderCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {candidates.map((c) => (
                  <TableRow key={c.id}>
                    <TableCell className="whitespace-nowrap">
                      {c.firstName} {c.lastName}
                    </TableCell>
                    <TableCell>{c.email ?? <span className="text-muted-foreground/60 italic">No email</span>}</TableCell>
                    <TableCell>
                      <Badge variant={c.status === 'active' ? 'success' : 'secondary'}>{c.status}</Badge>
                    </TableCell>
                    <TableCell className="whitespace-nowrap">{formatDate(c.createdAt)}</TableCell>
                    <TableCell>
                      <Button size="sm" variant="outline" onClick={() => setMergeTarget(c)}>
                        Merge Into
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </Card>
      )}

      {mergeTarget && selectedStub && (
        <Modal
          open={!!mergeTarget}
          onClose={() => {
            setMergeTarget(null);
            setMergeError('');
          }}
          title="Confirm Merge"
        >
          <p className="text-sm text-muted-foreground mb-4">
            Merge stub user <strong>{selectedStub.firstName} {selectedStub.lastName}</strong> into{' '}
            <strong>
              {mergeTarget.firstName} {mergeTarget.lastName}
              {mergeTarget.email ? ` (${mergeTarget.email})` : ''}
            </strong>
            ?
          </p>
          <p className="text-sm text-muted-foreground mb-4">
            All entries, invitations, and grants will be transferred to the target user. The stub user will be
            soft-deleted.
          </p>

          {mergeError && (
            <Alert variant="error" className="mb-4">
              {mergeError}
            </Alert>
          )}

          <ModalActions>
            <Button
              variant="secondary"
              onClick={() => {
                setMergeTarget(null);
                setMergeError('');
              }}
              disabled={merging}
            >
              Cancel
            </Button>
            <Button onClick={handleMergeConfirm} disabled={merging}>
              {merging ? 'Merging...' : 'Confirm Merge'}
            </Button>
          </ModalActions>
        </Modal>
      )}
    </PageContainer>
  );
}
