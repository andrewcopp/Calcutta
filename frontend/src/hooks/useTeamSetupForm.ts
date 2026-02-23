import { useState, useMemo, useCallback } from 'react';
import type { Tournament, TournamentTeam } from '../schemas/tournament';
import type { School } from '../schemas/school';
import { useTeamSetupMutations } from './useTeamSetupMutations';

export interface TeamSlot {
  schoolId: string;
  searchText: string;
}

export type RegionState = Record<number, TeamSlot[]>;

export function createEmptyRegion(): RegionState {
  const state: RegionState = {};
  for (let seed = 1; seed <= 16; seed++) {
    state[seed] = [{ schoolId: '', searchText: '' }];
  }
  return state;
}

export function createInitialRegions(regionNames: string[]): Record<string, RegionState> {
  const result: Record<string, RegionState> = {};
  for (const name of regionNames) {
    result[name] = createEmptyRegion();
  }
  return result;
}

export function getRegionList(tournament: Tournament): string[] {
  return [
    tournament.finalFourTopLeft || 'East',
    tournament.finalFourBottomLeft || 'West',
    tournament.finalFourTopRight || 'South',
    tournament.finalFourBottomRight || 'Midwest',
  ];
}

export function createRegionsFromTeams(
  regionNames: string[],
  teams: TournamentTeam[],
  schools: School[],
): Record<string, RegionState> {
  const regions = createInitialRegions(regionNames);
  const schoolMap = new Map(schools.map((s) => [s.id, s.name]));

  for (const team of teams) {
    const regionState = regions[team.region];
    if (!regionState) continue;

    const slot: TeamSlot = {
      schoolId: team.schoolId,
      searchText: schoolMap.get(team.schoolId) || '',
    };

    const slots = regionState[team.seed];
    if (!slots) continue;

    // First slot empty -- fill it
    if (!slots[0].schoolId) {
      slots[0] = slot;
    } else {
      // Play-in: add a second slot
      slots.push(slot);
    }
  }

  return regions;
}

export interface ValidationStats {
  total: number;
  playIns: number;
  perRegion: Record<string, number>;
  duplicates: string[];
}

export type SlotValidationState = Record<string, Record<string, 'none' | 'valid' | 'error'>>;

export function deriveUsedSchoolIds(regionList: string[], regions: Record<string, RegionState>): Set<string> {
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
}

export function deriveValidationStats(
  regionList: string[],
  regions: Record<string, RegionState>,
  schools: School[],
): ValidationStats {
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
}

export function deriveSlotValidation(
  regionList: string[],
  regions: Record<string, RegionState>,
  flashingSlots: Record<string, boolean>,
): SlotValidationState {
  const result: SlotValidationState = {};
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
}

export function applyUpdateSlot(
  prev: Record<string, RegionState>,
  region: string,
  seed: number,
  slotIndex: number,
  update: Partial<TeamSlot>,
): Record<string, RegionState> {
  const regionState = prev[region];
  if (!regionState) return prev;
  const updated = { ...regionState };
  const slots = [...updated[seed]];
  slots[slotIndex] = { ...slots[slotIndex], ...update };
  updated[seed] = slots;
  return { ...prev, [region]: updated };
}

export function applyAddPlayIn(
  prev: Record<string, RegionState>,
  region: string,
  seed: number,
): Record<string, RegionState> {
  const regionState = prev[region];
  if (!regionState) return prev;
  const updated = { ...regionState };
  const slots = [...updated[seed]];
  if (slots.length < 2) {
    slots.push({ schoolId: '', searchText: '' });
  }
  updated[seed] = slots;
  return { ...prev, [region]: updated };
}

export function applyRemovePlayIn(
  prev: Record<string, RegionState>,
  region: string,
  seed: number,
  slotIndex: number,
): Record<string, RegionState> {
  const regionState = prev[region];
  if (!regionState) return prev;
  const updated = { ...regionState };
  const slots = [...updated[seed]];
  if (slots.length > 1) {
    const keepIndex = slotIndex === 0 ? 1 : 0;
    updated[seed] = [slots[keepIndex]];
  }
  return { ...prev, [region]: updated };
}

export function applySlotBlur(
  prev: Record<string, RegionState>,
  region: string,
  seed: number,
  slotIndex: number,
): { regions: Record<string, RegionState>; shouldFlash: boolean; flashKey: string } {
  const regionState = prev[region];
  if (!regionState) return { regions: prev, shouldFlash: false, flashKey: '' };
  const slot = regionState[seed][slotIndex];
  if (!slot.searchText || slot.schoolId) return { regions: prev, shouldFlash: false, flashKey: '' };

  // Clear the orphaned search text
  const updated = { ...regionState };
  const slots = [...updated[seed]];
  slots[slotIndex] = { ...slots[slotIndex], searchText: '' };
  updated[seed] = slots;

  const flashKey = `${region}-${seed}-${slotIndex}`;
  return { regions: { ...prev, [region]: updated }, shouldFlash: true, flashKey };
}

export function collectTeamsForSubmission(
  regionList: string[],
  regions: Record<string, RegionState>,
): { schoolId: string; seed: number; region: string }[] {
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
  return teams;
}

export interface TeamSetupFormState {
  regionList: string[];
  regions: Record<string, RegionState>;
  activeTab: string;
  setActiveTab: (tab: string) => void;
  errors: string[];
  stats: ValidationStats;
  slotValidation: SlotValidationState;
  usedSchoolIds: Set<string>;
  updateSlot: (region: string, seed: number, slotIndex: number, update: Partial<TeamSlot>) => void;
  addPlayIn: (region: string, seed: number) => void;
  removePlayIn: (region: string, seed: number, slotIndex: number) => void;
  handleSlotBlur: (region: string, seed: number, slotIndex: number) => void;
  handleSubmit: () => void;
  isPending: boolean;
  createEmptyRegion: () => RegionState;
}

interface UseTeamSetupFormParams {
  tournament: Tournament;
  schools: School[];
  initialTeams?: TournamentTeam[];
}

export function useTeamSetupForm({ tournament, schools, initialTeams }: UseTeamSetupFormParams): TeamSetupFormState {
  const regionList = useMemo(() => getRegionList(tournament), [tournament]);

  const [regions, setRegions] = useState(() =>
    initialTeams && initialTeams.length > 0
      ? createRegionsFromTeams(regionList, initialTeams, schools)
      : createInitialRegions(regionList),
  );
  const [activeTab, setActiveTab] = useState(() => regionList[0]);
  const [flashingSlots, setFlashingSlots] = useState<Record<string, boolean>>({});

  const usedSchoolIds = useMemo(() => deriveUsedSchoolIds(regionList, regions), [regions, regionList]);
  const stats = useMemo(() => deriveValidationStats(regionList, regions, schools), [regions, schools, regionList]);
  const slotValidation = useMemo(
    () => deriveSlotValidation(regionList, regions, flashingSlots),
    [regions, flashingSlots, regionList],
  );

  const updateSlot = useCallback((region: string, seed: number, slotIndex: number, update: Partial<TeamSlot>) => {
    setRegions((prev) => applyUpdateSlot(prev, region, seed, slotIndex, update));
  }, []);

  const addPlayIn = useCallback((region: string, seed: number) => {
    setRegions((prev) => applyAddPlayIn(prev, region, seed));
  }, []);

  const removePlayIn = useCallback((region: string, seed: number, slotIndex: number) => {
    setRegions((prev) => applyRemovePlayIn(prev, region, seed, slotIndex));
  }, []);

  const handleSlotBlur = useCallback((region: string, seed: number, slotIndex: number) => {
    setRegions((prev) => {
      const result = applySlotBlur(prev, region, seed, slotIndex);
      if (result.shouldFlash) {
        setFlashingSlots((f) => ({ ...f, [result.flashKey]: true }));
        setTimeout(() => {
          setFlashingSlots((f) => ({ ...f, [result.flashKey]: false }));
        }, 1000);
      }
      return result.regions;
    });
  }, []);

  const { errors, handleSubmit, isPending } = useTeamSetupMutations({
    tournamentId: tournament.id,
    regionList,
    regions,
  });

  return {
    regionList,
    regions,
    activeTab,
    setActiveTab,
    errors,
    stats,
    slotValidation,
    usedSchoolIds,
    updateSlot,
    addPlayIn,
    removePlayIn,
    handleSlotBlur,
    handleSubmit,
    isPending,
    createEmptyRegion,
  };
}
