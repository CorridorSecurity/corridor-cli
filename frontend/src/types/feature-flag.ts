export enum FlagKey {
  EXTENSION_REQUIRES_CLI = "extension_requires_cli",
}

export interface FlagDefaults {
  [FlagKey.EXTENSION_REQUIRES_CLI]: boolean;
}

export const flagDefaults: FlagDefaults = {
  [FlagKey.EXTENSION_REQUIRES_CLI]: false,
};
