import { useQuery } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { TournamentSetupTeamsForm } from './Tournament/TournamentSetupTeamsForm';

export function TournamentSetupTeamsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournament(id!),
  });

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    queryFn: () => schoolService.getSchools(),
  });

  const teamsQuery = useQuery({
    queryKey: queryKeys.tournaments.teams(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournamentTeams(id!),
  });

  if (!id) {
    return (
      <PageContainer>
        <Alert variant="error">Missing tournament ID</Alert>
      </PageContainer>
    );
  }

  if (tournamentQuery.isLoading || schoolsQuery.isLoading || teamsQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading..." />
      </PageContainer>
    );
  }

  const tournament = tournamentQuery.data;
  if (!tournament) {
    return (
      <PageContainer>
        <Alert variant="error">Tournament not found</Alert>
      </PageContainer>
    );
  }

  const schools = schoolsQuery.data || [];
  const schoolOptions = schools.map((s) => ({ id: s.id, label: s.name }));
  const existingTeams = teamsQuery.data || [];
  const isEditing = existingTeams.length > 0;
  const pageTitle = isEditing ? 'Edit Field' : 'Setup Teams';

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Tournaments', href: '/admin/tournaments' },
          { label: tournament.name, href: `/admin/tournaments/${id}` },
          { label: pageTitle },
        ]}
      />
      <PageHeader
        title={pageTitle}
        subtitle={tournament.name}
        actions={
          <Button variant="outline" onClick={() => navigate(`/admin/tournaments/${id}`)}>
            Cancel
          </Button>
        }
      />

      <TournamentSetupTeamsForm
        tournament={tournament}
        schools={schools}
        schoolOptions={schoolOptions}
        initialTeams={existingTeams}
      />
    </PageContainer>
  );
}
