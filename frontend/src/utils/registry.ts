import type { ExternalReg } from '@/types/registry';

export const getExtStatus = (status: string, enabled: boolean) => {
  if (status === "") {
      return "status-active";
  }
  if (!enabled) {
      return "status-disabled";
  }
  return `status-${status}`;
};

export const getTypeStr = (e: ExternalReg) : 'external-system' | 'internal-registry' => {
if (e.External) {
  return "external-system";
}

return "internal-registry";
};