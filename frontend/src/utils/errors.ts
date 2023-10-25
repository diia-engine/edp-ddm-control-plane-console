import i18n from "../localization";

export const getErrorMessage = (key: string): string => {
  switch (key) {
    case 'required':
      return i18n.global.t('errors.required');
    case 'moreThanMaxValue':
      return i18n.global.t('errors.moreThanMaxValue');
    case 'isUnique':
      return i18n.global.t('errors.isUnique');
    case 'rateLimitError':
      return i18n.global.t('errors.rateLimitError');
    case 'registryNameAlreadyExists':
      return i18n.global.t('errors.registryNameAlreadyExists');
    case 'invalidFileType':
      return i18n.global.t('errors.invalidFileType');
    case 'nonUniqKeyName':
      return i18n.global.t('errors.nonUniqKeyName');
    case 'dsKeysNotFound':
      return i18n.global.t('errors.dsKeysNotFound');
    default:
      return i18n.global.t('errors.checkFormat');
  }
};
