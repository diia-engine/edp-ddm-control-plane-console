import i18n from '@/localization';
import { KeyType } from '@/types/cluster';

export const KEY_VARIANTS = () => [
  { title: i18n.global.t('components.keysManagement.modals.addKey.fields.hardwareMedia'), value: KeyType.hardware },
  { title:  i18n.global.t('components.keysManagement.modals.addKey.fields.fileMedia'), value: KeyType.file },
];