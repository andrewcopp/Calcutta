import { schoolHandlers } from './school';
import { userHandlers } from './user';
import { bracketHandlers } from './bracket';
import { hallOfFameHandlers } from './hallOfFame';
import { tournamentHandlers } from './tournament';
import { calcuttaHandlers } from './calcutta';
import { adminHandlers } from './admin';
import { labHandlers } from './lab';

export const handlers = [
  ...schoolHandlers,
  ...userHandlers,
  ...bracketHandlers,
  ...hallOfFameHandlers,
  ...tournamentHandlers,
  ...calcuttaHandlers,
  ...adminHandlers,
  ...labHandlers,
];
