import type { ExternalReg } from '@/types/registry';
import transform from 'lodash/transform';
import isEqual from 'lodash/isEqual';
import isObject from 'lodash/isObject';
import type { StoredKey } from '@/types/cluster';

export const getExtStatus = (status: string, enabled: boolean) => {
  if (!enabled) {
    return "status-disabled";
  }
  if (status === "") {
      return "status-active";
  }
  return `status-${status}`;
};

export const getTypeStr = (e: ExternalReg) : 'external-system' | 'internal-registry' => {
if (e.External) {
  return "external-system";
}

return "internal-registry";
};

export const jsonDiff = (object: any, base: any): string[] => {
	function changes(object: any, base: { [x: string]: any; }) {
		return transform(object, function(result: { [x: string]: any; }, value: any, key: string | number) {
			if (!isEqual(value, base[key])) {
				result[key] = (isObject(value) && isObject(base[key])) ? changes(value, base[key]) : value;
			}
		});
	}

	return Object.keys(changes(object, base));
};

export function filterKeysByRegistry(keys:  Record<string, StoredKey>, registry: string): string[] {
  return Object.entries(keys).filter(([, data]) => (data.allowedRegistries || []).includes(registry)).map(([key]) => key);
}