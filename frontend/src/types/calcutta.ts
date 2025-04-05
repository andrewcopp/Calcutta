export interface Calcutta {
  id: string;
  name: string;
  tournamentId: string;
  ownerId: string;
  created: string;
  updated: string;
}

export interface CalcuttaEntry {
  id: string;
  name: string;
  calcuttaId: string;
  userId?: string;
  created: string;
  updated: string;
}

export interface CalcuttaEntryTeam {
  id: string;
  entryId: string;
  teamId: string;
  bid: number;
  created: string;
  updated: string;
  team?: {
    id: string;
    schoolId: string;
    school?: {
      id: string;
      name: string;
    };
  };
} 