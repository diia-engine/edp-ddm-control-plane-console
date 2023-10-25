import { createI18n } from 'vue-i18n';
import get from 'lodash/get';

import uk from '#/locale/uk.json';
import en from '#/locale/en.json';
import { isObject } from 'lodash';

const nsSeparator = '~';
const messages = {
  uk,
  en,
};

const messageResolver = (obj: any, path: string) => {
  const result = get(obj, path);
  if (!result) {
    const nsPath = path.split(nsSeparator);
    const key = nsPath.pop();
    const nsObj = get(obj, nsPath.join('.'));
    return isObject(nsObj) && key ? (nsObj as any)[key] : undefined;
  }
  return result;
};

const i18n = createI18n({
  locale: 'uk',
  fallbackLocale: 'uk',
  allowComposition: true,
  messages,
  keySeparator: '.',
  nsSeparator,
  messageResolver: messageResolver,
  globalize: true,
});

const getDateTimeFormat = () => {
  return i18n.global.locale === 'en' ? 'MM/DD/YYYY h:mm A' : 'DD.MM.YYYY HH:mm';
};

const getDateTimePreciseFormat = () => {
  return i18n.global.locale === 'en' ? 'MM/DD/YYYY h:mm:ss A' : 'DD.MM.YYYY HH:mm:ss';
};

const getNamespaceMessage = (namespace: string, key: string) => {
  return i18n.global.t(namespace + nsSeparator + key);
};

export { getDateTimeFormat, getNamespaceMessage, getDateTimePreciseFormat };

export default i18n;
