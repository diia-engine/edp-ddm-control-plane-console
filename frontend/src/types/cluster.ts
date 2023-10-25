export type FileKey = {
  deviceType: KeyType,
  fileKeyFile: File | null,
  fileKeyIssuer: string,
  fileKeyPassword: string,
  fileKeyName: string,
  allowedRegistries: string[],
};

export type HardwareKey = {
  deviceType: KeyType,
  hardKeyName: string,
  hardKeyType: string,
  hardKeyIssuer: string,
  hardKeyIssuerHost: string,
  hardKeyIssuerPort: string,
  hardKeySerialNumber: string,
  hardKeyPort: string,
  hardKeyHost: string,
  hardKeyMask: string,
  hardKeyPassword: string,
  allowedRegistries: string[],
};

export enum KeyType {
  file = 'file',
  hardware = 'hardware'
}

export type OutKey = FileKey & HardwareKey & { deviceType: KeyType};
export type StoredKey = {
  'device-type': KeyType,
  file?: string,
  issuer: string,
  password: string,
  type?: string,
  device?: string,
  'osplm.ini'?: string,
  allowedRegistries: string[],
}

export type TableViewKey = {
  name: string,
  deviceType: string,
  issuer: string,
  allowedRegistries: string[],
}

export type PreparedHardKey = {
  deviceType: KeyType,
  hardKeyName: string,
  hardKeyIssuer: string,
  hardKeyPassword: string,
  hardKeyType: string,
  hardKeyDevice: string,
  allowedRegistries: string[],
}

export type PreparedFileKey = {
  deviceType: KeyType,
  fileKeyFile: string,
  fileKeyIssuer: string,
  fileKeyPassword: string,
  fileKeyName: string,
  allowedRegistries: string[],
}
