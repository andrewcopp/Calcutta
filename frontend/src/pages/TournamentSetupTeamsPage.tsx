import React, { useState, useMemo, useCallback, useEffect, useRef } from 'react';
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

function createInitialRegions(regionNames: string[]): Record<string, RegionState> {
  const result: Record<string, RegionState> = {};
  for (const name of regionNames) {
    result[name] = createEmptyRegion();
  }
  return result;
}

export const TournamentSetupTeamsPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [regions, setRegions] = useState<Record<string, RegionState>>({});
  const [errors, setErrors] = useState<string[]>([]);
  const [flashingSlots, setFlashingSlots] = useState<Record<string, boolean>>({});
  const [initialized, setInitialized] = useState(false);
  const [activeTab, setActiveTab] = useState<string>('');

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournament(id!),
  });

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    queryFn: () => schoolService.getSchools(),
  });

  // Derive region names from tournament data (read-only)
  const regionList = useMemo(() => {
    const t = tournamentQuery.data;
    if (!t) return ['East', 'West', 'South', 'Midwest'];
    return [
      t.finalFourTopLeft || 'East',
      t.finalFourBottomLeft || 'West',
      t.finalFourTopRight || 'South',
      t.finalFourBottomRight || 'Midwest',
    ];
  }, [tournamentQuery.data]);

  // Initialize regions state once after tournament loads
  useEffect(() => {
    if (initialized || !tournamentQuery.data) return;
    setRegions(createInitialRegions(regionList));
    setActiveTab(regionList[0]);
    setInitialized(true);
  }, [tournamentQuery.data, initialized, regionList]);

  const schools = useMemo(() => schoolsQuery.data || [], [schoolsQuery.data]);
  const schoolOptions = useMemo(
    () => schools.map((s) => ({ id: s.id, label: s.name })),
    [schools]
  );

  // Collect all used school IDs for exclusion
  const usedSchoolIds = useMemo(() => {
    const ids = new Set<string>();
    for (const regionName of regionList) {
      const regionState = regions[regionName];
      if (!regionState) continue;
      for (let seed = 1; seed <= 16; seed++) {
        for (const slot of regionState[seed]) {
          if (slot.schoolId) ids.add(slot.schoolId);
        }
      }
    }
    return ids;
  }, [regions, regionList]);

  // Validation stats
  const stats = useMemo(() => {
    let total = 0;
    let playIns = 0;
    const perRegion: Record<string, number> = {};
    const duplicates: string[] = [];
    const schoolCounts: Record<string, number> = {};

    for (const regionName of regionList) {
      let regionCount = 0;
      const regionState = regions[regionName];
      if (!regionState) continue;
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
      perRegion[regionName] = regionCount;
    }

    for (const [schoolId, count] of Object.entries(schoolCounts)) {
      if (count > 1) {
        const school = schools.find((s) => s.id === schoolId);
        duplicates.push(school?.name || schoolId);
      }
    }

    return { total, playIns, perRegion, duplicates };
  }, [regions, schools, regionList]);

  // Slot validation state per region
  // schoolId takes priority over flash — a valid selection never shows red
  const slotValidation = useMemo(() => {
    const result: Record<string, Record<string, 'none' | 'valid' | 'error'>> = {};
    for (const regionName of regionList) {
      const regionState = regions[regionName];
      if (!regionState) continue;
      const regionResult: Record<string, 'none' | 'valid' | 'error'> = {};
      for (let seed = 1; seed <= 16; seed++) {
        const slots = regionState[seed];
        slots.forEach((slot, slotIndex) => {
          const key = `${seed}-${slotIndex}`;
          const flashKey = `${regionName}-${key}`;
          if (slot.schoolId) {
            regionResult[key] = 'valid';
          } else if (flashingSlots[flashKey]) {
            regionResult[key] = 'error';
          } else {
            regionResult[key] = 'none';
          }
        });
      }
      result[regionName] = regionResult;
    }
    return result;
  }, [regions, flashingSlots, regionList]);

  const updateSlot = useCallback(
    (region: string, seed: number, slotIndex: number, update: Partial<TeamSlot>) => {
      setRegions((prev) => {
        const regionState = prev[region];
        if (!regionState) return prev;
        const updated = { ...regionState };
        const slots = [...updated[seed]];
        slots[slotIndex] = { ...slots[slotIndex], ...update };
        updated[seed] = slots;
        return { ...prev, [region]: updated };
      });
    },
    []
  );

  const addPlayIn = useCallback((region: string, seed: number) => {
    setRegions((prev) => {
      const regionState = prev[region];
      if (!regionState) return prev;
      const updated = { ...regionState };
      const slots = [...updated[seed]];
      if (slots.length < 2) {
        slots.push({ schoolId: '', searchText: '' });
      }
      updated[seed] = slots;
      return { ...prev, [region]: updated };
    });
  }, []);

  const removePlayIn = useCallback((region: string, seed: number) => {
    setRegions((prev) => {
      const regionState = prev[region];
      if (!regionState) return prev;
      const updated = { ...regionState };
      const slots = [...updated[seed]];
      if (slots.length > 1) {
        slots.pop();
      }
      updated[seed] = slots;
      return { ...prev, [region]: updated };
    });
  }, []);

  // Use a ref to access current regions for blur checking without stale closure
  const regionsRef = useRef(regions);
  regionsRef.current = regions;

  const handleSlotBlur = useCallback(
    (region: string, seed: number, slotIndex: number) => {
      const regionState = regionsRef.current[region];
      if (!regionState) return;
      const slot = regionState[seed][slotIndex];
      if (!slot.searchText || slot.schoolId) return;

      // Clear the orphaned search text
      updateSlot(region, seed, slotIndex, { searchText: '' });

      // Flash red briefly
      const flashKey = `${region}-${seed}-${slotIndex}`;
      setFlashingSlots((prev) => ({ ...prev, [flashKey]: true }));
      setTimeout(() => {
        setFlashingSlots((prev) => ({ ...prev, [flashKey]: false }));
      }, 1000);
    },
    [updateSlot]
  );

  const replaceTeamsMutation = useMutation({
    mutationFn: async () => {
      const teams: { schoolId: string; seed: number; region: string }[] = [];
      for (const regionName of regionList) {
        const regionState = regions[regionName];
        if (!regionState) continue;
        for (let seed = 1; seed <= 16; seed++) {
          for (const slot of regionState[seed]) {
            if (slot.schoolId) {
              teams.push({ schoolId: slot.schoolId, seed, region: regionName });
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

      <ValidationPanel stats={stats} regionNames={regionList} />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          {regionList.map((regionName) => (
            <TabsTrigger key={regionName} value={regionName}>
              {regionName} ({stats.perRegion[regionName] || 0})
            </TabsTrigger>
          ))}
        </TabsList>

        {regionList.map((regionName) => (
          <TabsContent key={regionName} value={regionName}>
            <Card>
              <RegionSeedForm
                region={regionName}
                regionState={regions[regionName] || createEmptyRegion()}
                schoolOptions={schoolOptions}
                schools={schools}
                usedSchoolIds={usedSchoolIds}
                updateSlot={updateSlot}
                addPlayIn={addPlayIn}
                removePlayIn={removePlayIn}
                onSlotBlur={handleSlotBlur}
                slotValidation={slotValidation[regionName]}
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
