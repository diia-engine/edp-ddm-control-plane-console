import {  KeyType, type FileKey, type HardwareKey, type PreparedFileKey, type PreparedHardKey } from '@/types/cluster';
import ini from 'js-ini';

export interface AddKeyDTO {
  caName: string,
  caHost: string,
  caPort: string,
  keySn: string,
  keyHost: string,
  keyMask: string,
}

export interface CMP {
  Address: string,
  CommonName: string,
  Port: string,
  Use: string,
}

export interface KeyModule { 
  Address: string,
  AddressMask: string,
  OrderNumber: string,
  SN: string,
}

export default class OsplmIniEditor {
  static getSemicolonSeparatedString(currentValue: string, newValue: string): string {
    const isValueExists = currentValue.toString().split(';').some((value) => value === newValue.toString());
    if (isValueExists) {
      return currentValue.toString();
    }
    if (currentValue.toString() === '') {
      return newValue.toString();
    }
    return currentValue.toString().concat(';', newValue.toString());
  }
  private readonly CMP_PATH = '\\SOFTWARE\\Institute of Informational Technologies\\Certificate Authority-1.3\\End User\\CMP';
  private readonly KEY_PATH_PREFIX = '\\SOFTWARE\\Institute of Informational Technologies\\Key Medias\\NCM Gryada-301\\Modules\\';
  private iniObj: Record<string, any>;
  constructor(private readonly iniString: string) {
    this.iniObj = ini.parse(this.iniString, { autoTyping: false });
  }
  get keysLength(): number {
    return Object.entries(this.iniObj).filter(([key]) => key.startsWith(this.KEY_PATH_PREFIX)).length;
  }
  get lastOrderNumber() {
    return Object.entries(this.iniObj)
    .filter(([key]) => key.startsWith(this.KEY_PATH_PREFIX))
    .map(([,value]: [string, KeyModule]) => parseInt(value.OrderNumber)).sort().pop() || 0;
  }
  toString(): string {
    return ini.stringify(this.iniObj);
  }
  addKey(key: AddKeyDTO) {
    // add key to ini modules
    const keyModule: KeyModule = {
      Address: key.keyHost,
      AddressMask: key.keyMask,
      OrderNumber: (this.keysLength ? (this.lastOrderNumber + 1) : 0).toString(),
      SN: key.keySn.toString(),
    };
    this.iniObj[this.KEY_PATH_PREFIX.concat(key.keySn)] = keyModule;
    // add key to ini cmp
    const currentCmp: CMP = this.iniObj[this.CMP_PATH];
    const newCmp: CMP = {
      Address: OsplmIniEditor.getSemicolonSeparatedString(currentCmp.Address || '', key.caHost),
      // yes, we  should always use the same CommonName= in ini, so it's empty
      CommonName: '',
      Port: OsplmIniEditor.getSemicolonSeparatedString(currentCmp.Port || '', key.caPort),
      Use: '1',
    };
    this.iniObj[this.CMP_PATH] = newCmp;
    return this.iniObj;
  }
  removeKey(keyName: string, keys: Array<PreparedHardKey | PreparedFileKey | HardwareKey | FileKey>) {
    const hardKeys = keys.filter((key) => key.deviceType === KeyType.hardware) as Array<PreparedHardKey | HardwareKey>;
    const key = hardKeys.find((key) => key.hardKeyName === keyName)!;
    const keySn = this.getHardKeySnByName(key);
    const keysWithSerial = hardKeys.filter((k) => this.getHardKeySnByName(k) === keySn);
    // only last key with serial number should be removed from ini
    if (keysWithSerial.length > 1) {
      return;
    }
    // remove key from ini modules
    const keyPath = this.KEY_PATH_PREFIX.concat(keySn);
    delete this.iniObj[keyPath];
    // TODO: remove key from ini cmp
  }

  private getHardKeySnByName(key: PreparedHardKey | HardwareKey): string {
    return ('hardKeySerialNumber' in key) ? key.hardKeySerialNumber : key.hardKeyDevice.split(':')[0];
  }
}