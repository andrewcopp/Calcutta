export const PERMISSIONS = {
  ADMIN: 'admin',
  ADMIN_USERS_READ: 'admin.users.read',
  ADMIN_API_KEYS_WRITE: 'admin.api_keys.write',
  ADMIN_BUNDLES_EXPORT: 'admin.bundles.export',
  ADMIN_HOF_READ: 'admin.hof.read',
  TOURNAMENT_GAME_WRITE: 'tournament.game.write',
  POOL_CONFIG_WRITE: 'pool.config.write',
  ADMIN_USERS_WRITE: 'admin.users.write',
  LAB_READ: 'lab.read',
} as const;

export const ADMIN_PERMISSIONS = [
  PERMISSIONS.ADMIN_USERS_READ,
  PERMISSIONS.ADMIN_API_KEYS_WRITE,
  PERMISSIONS.ADMIN_BUNDLES_EXPORT,
  PERMISSIONS.ADMIN_HOF_READ,
  PERMISSIONS.TOURNAMENT_GAME_WRITE,
  PERMISSIONS.POOL_CONFIG_WRITE,
] as const;
