import React, { useState, useMemo, useCallback } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { ValidationPanel } from './Tournament/ValidationPanel';
import { RegionSeedForm } from './Tournament/RegionSeedForm';

const REGIONS = ['East', 'West', 'South', 'Midwest'] as const;
type Region = (typeof REGIONS)[number];

interface TeamSlot {
  schoolId: string;
  searchText: string;
}

type RegionState = Record<number, TeamSlot[]>;

function createEmptyRegion(): RegionState {
  const state: RegionState = {};
  for (let seed = 1; seed <= 16; seed++) {
    state[seed] = [{ schoolId: '', searchText: '' }];
  }
  return state;
}

function createInitialState(): Record<Region, RegionState> {
  return {
    East: createEmptyRegion(),
    West: createEmptyRegion(),
    South: createEmptyRegion(),
    Midwest: createEmptyRegion(),
  };
}

export const TournamentSetupTeamsPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [regions, setRegions] = useState<Record<Region, RegionState>>(createInitialState);
  const [errors, setErrors] = useState<string[]>([]);

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournament(id!),
  });

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    queryFn: () => schoolService.getSchools(),
  });

  const schools = schoolsQuery.data || [];
  const schoolOptions = useMemo(
    () => schools.map((s) => ({ id: s.id, label: s.name })),
    [schools]
  );

  // Collect all used school IDs for exclusion
  const usedSchoolIds = useMemo(() => {
    const ids = new Set<string>();
    for (const region of REGIONS) {
      const regionState = regions[region];
      for (let seed = 1; seed <= 16; seed++) {
        for (const slot of regionState[seed]) {
          if (slot.schoolId) ids.add(slot.schoolId);
        }
      }
    }
    return ids;
  }, [regions]);

  // Validation stats
  const stats = useMemo(() => {
    let total = 0;
    let playIns = 0;
    const perRegion: Record<string, number> = {};
    const duplicates: string[] = [];
    const schoolCounts: Record<string, number> = {};

    for (const region of REGIONS) {
      let regionCount = 0;
      const regionState = regions[region];
      for (let seed = 1; seed <= 16; seed++) {
        const slots = regionState[seed];
        const filledSlots = slots.filter((s) => s.schoolId);
        regionCount += filledSlots.length;
        total += filledSlots.length;
        if (filledSlots.length === 2) playIns++;
        for (const slot of filledSlots) {
          schoolCounts[slot.schoolId] = (schoolCounts[slot.schoolId] || 0) + 1;
        }
      }
      perRegion[region] = regionCount;
    }

    for (const [schoolId, count] of Object.entries(schoolCounts)) {
      if (count > 1) {
        const school = schools.find((s) => s.id === schoolId);
        duplicates.push(school?.name || schoolId);
      }
    }

    return { total, playIns, perRegion, duplicates };
  }, [regions, schools]);

  const updateSlot = useCallback(
    (region: Region, seed: number, slotIndex: number, update: Partial<TeamSlot>) => {
      setRegions((prev) => {
        const regionState = { ...prev[region] };
        const slots = [...regionState[seed]];
        slots[slotIndex] = { ...slots[slotIndex], ...update };
        regionState[seed] = slots;
        return { ...prev, [region]: regionState };
      });
    },
    []
  );

  const addPlayIn = useCallback((region: Region, seed: number) => {
    setRegions((prev) => {
      const regionState = { ...prev[region] };
      const slots = [...regionState[seed]];
      if (slots.length < 2) {
        slots.push({ schoolId: '', searchText: '' });
      }
      regionState[seed] = slots;
      return { ...prev, [region]: regionState };
    });
  }, []);

  const removePlayIn = useCallback((region: Region, seed: number) => {
    setRegions((prev) => {
      const regionState = { ...prev[region] };
      const slots = [...regionState[seed]];
      if (slots.length > 1) {
        slots.pop();
      }
      regionState[seed] = slots;
      return { ...prev, [region]: regionState };
    });
  }, []);

  const replaceTeamsMutation = useMutation({
    mutationFn: async () => {
      const teams: { schoolId: string; seed: number; region: string }[] = [];
      for (const region of REGIONS) {
        const regionState = regions[region];
        for (let seed = 1; seed <= 16; seed++) {
          for (const slot of regionState[seed]) {
            if (slot.schoolId) {
              teams.push({ schoolId: slot.schoolId, seed, region });
            }
          }
        }
      }
      return tournamentService.replaceTeams(id!, teams);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.teams(id) });
      navigate(`/admin/tournaments/${id}`);
    },
    onError: (err: unknown) => {
      // Handle bracket validation errors
      const apiErr = err as { body?: { errors?: string[] } };
      if (apiErr?.body?.errors) {
        setErrors(apiErr.body.errors);
      } else {
        setErrors([err instanceof Error ? err.message : 'Failed to save teams']);
      }
    },
  });

  const handleSubmit = () => {
    setErrors([]);
    replaceTeamsMutation.mutate();
  };

  if (!id) {
    return (
      <PageContainer>
        <Alert variant="error">Missing tournament ID</Alert>
      </PageContainer>
    );
  }

  if (tournamentQuery.isLoading || schoolsQuery.isLoading) {
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

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Tournaments', href: '/admin/tournaments' },
          { label: tournament.name, href: `/admin/tournaments/${id}` },
          { label: 'Setup Teams' },
        ]}
      />
      <PageHeader
        title="Setup Teams"
        subtitle={tournament.name}
        actions={
          <Button variant="outline" onClick={() => navigate(`/admin/tournaments/${id}`)}>
            Cancel
          </Button>
        }
      />

      {errors.length > 0 && (
        <Alert variant="error" className="mb-4">
          <div className="font-semibold mb-1">Validation Errors</div>
          <ul className="list-disc list-inside text-sm space-y-1">
            {errors.map((err, i) => (
              <li key={i}>{err}</li>
            ))}
          </ul>
        </Alert>
      )}

      <ValidationPanel stats={stats} />

      {/* Region Tabs */}
      <Tabs defaultValue="East">
        <TabsList>
          {REGIONS.map((region) => (
            <TabsTrigger key={region} value={region}>
              {region} ({stats.perRegion[region] || 0})
            </TabsTrigger>
          ))}
        </TabsList>

        {REGIONS.map((region) => (
          <TabsContent key={region} value={region}>
            <Card>
              <RegionSeedForm
                region={region}
                regionState={regions[region]}
                schoolOptions={schoolOptions}
                schools={schools}
                usedSchoolIds={usedSchoolIds}
                updateSlot={updateSlot}
                addPlayIn={addPlayIn}
                removePlayIn={removePlayIn}
              />
            </Card>
          </TabsContent>
        ))}
      </Tabs>

      <div className="flex justify-end mt-6">
        <Button
          onClick={handleSubmit}
          disabled={replaceTeamsMutation.isPending || stats.total === 0}
          loading={replaceTeamsMutation.isPending}
        >
          {replaceTeamsMutation.isPending ? 'Saving...' : 'Save Teams'}
        </Button>
      </div>
    </PageContainer>
  );
};
