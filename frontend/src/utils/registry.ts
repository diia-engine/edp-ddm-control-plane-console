import type { ExternalReg } from '@/types/registry';

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