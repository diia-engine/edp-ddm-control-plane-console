<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
import { getErrorMessage } from '@/utils';
import { computed, watch, type Ref } from 'vue';
import { useForm } from 'vee-validate';
import * as Yup from 'yup';
import type { CitizenAuthFlow, CitizenPortalSettings } from '@/types/registry';
import { CitizenAuthType } from '@/types/registry';

interface HTMLEvent<T extends EventTarget = HTMLElement> extends Event {
  target: T
}



interface OutFormValues extends CitizenAuthFlow {
  citizenPortal: CitizenPortalSettings
}

interface FormValues {
  edrCheckEnabled: boolean,
  authType: CitizenAuthType,
  widgetUrl: string,
  idGovUaUrl: string,
  widgetHeight: number,
  clientId: string,
  secret: string,
  copyFromAuthWidget: boolean,
  signWidgetHeight: number,
  signWidgetUrl: string,
}
interface RegistryRecipientAuthProps {
  keycloakSettings: CitizenAuthFlow,
  citizenPortalSettings: CitizenPortalSettings,
}
const props = defineProps<RegistryRecipientAuthProps>();
const isSecretExists = props.keycloakSettings?.registryIdGovUa?.clientSecret?.length > 0;

const validationSchema = Yup.object<FormValues>({
  authType: Yup.string()
    .required()
    .oneOf([CitizenAuthType.platformIdGovUa, CitizenAuthType.widget]),
  widgetUrl: Yup.string()
    .when('authType', {
      is: (value: CitizenAuthType) => value === CitizenAuthType.widget,
      then: (schema) => schema.required().url(),
    }),
  idGovUaUrl: Yup.string()
    .when('authType', {
      is: (value: CitizenAuthType) => value === CitizenAuthType.registryIdGovUa,
      then: (schema) => schema.required().url(),
    }),
  widgetHeight: Yup.number()
  .when('authType', {
    is: (value: CitizenAuthType) => value === CitizenAuthType.widget,
    then: (schema) => schema.required().min(1).integer().positive().typeError('wrongFormat'),
  }),
  clientId: Yup.string()
    .when('authType', {
      is: (value: CitizenAuthType) => value === CitizenAuthType.registryIdGovUa,
      then: (schema) => schema.required(),
    }),
  secret: Yup.string()
    .when('authType', {
      is: (value: CitizenAuthType) => value === CitizenAuthType.registryIdGovUa && !isSecretExists,
      then: (schema) => schema.required(),
    }),
  signWidgetHeight: Yup.number()
    .when('copyFromAuthWidget', {
      is: (value: boolean) => !value,
      then: (schema) => schema.required().min(1).integer().positive().typeError('wrongFormat'),
    }),
  signWidgetUrl: Yup.string()
    .when('copyFromAuthWidget', {
      is: (value: boolean) => !value,
      then: (schema) => schema.required().url(),
    }),
});
const defaultWidget = {
    url: 'https://eu.iit.com.ua/sign-widget/v20200922/',
    height: 720,
};
const defaultValues: OutFormValues = {
  authType: CitizenAuthType.widget,
  edrCheck: true,
  widget: defaultWidget,
  registryIdGovUa: {
    url: '',
    clientId: '',
    clientSecret: '',
  },
  citizenPortal: {
    signWidget: {
      copyFromAuthWidget: false,
      height: defaultWidget.height,
      url: defaultWidget.url,
    },
  }
};

const { errors, useFieldModel, validate, values, setFieldValue } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    edrCheckEnabled: props.keycloakSettings?.edrCheck ?? defaultValues.edrCheck,
    authType: props.keycloakSettings?.authType || defaultValues.authType,
    widgetUrl: props.keycloakSettings?.widget.url ?? defaultValues.widget.url,
    widgetHeight: props.keycloakSettings?.widget.height ?? defaultValues.widget.height,
    idGovUaUrl: props.keycloakSettings?.registryIdGovUa?.url ?? defaultValues.registryIdGovUa.url,
    clientId: props.keycloakSettings?.registryIdGovUa?.clientId ?? defaultValues.registryIdGovUa.clientId,
    secret: defaultValues.registryIdGovUa.clientSecret,
    copyFromAuthWidget: props.citizenPortalSettings?.signWidget?.copyFromAuthWidget || defaultValues.citizenPortal.signWidget.copyFromAuthWidget,
    signWidgetHeight: props.citizenPortalSettings?.signWidget?.height ?? defaultValues.citizenPortal.signWidget.height,
    signWidgetUrl: props.citizenPortalSettings?.signWidget?.url ?? defaultValues.citizenPortal.signWidget.url,
  }
});

const [
  edrCheckEnabled,
  authType,
  url,
  widgetHeight,
  clientId,
  secret,
  idGovUaUrl,
  copyFromAuthWidget,
  signWidgetHeight,
  signWidgetUrl,
] = useFieldModel([
  'edrCheckEnabled',
  'authType',
  'widgetUrl',
  'widgetHeight',
  'clientId',
  'secret',
  'idGovUaUrl',
  'copyFromAuthWidget',
  'signWidgetHeight',
  'signWidgetUrl',
]);

function validator() {
  return new Promise((resolve) => {
    validate().then((res) => {
      if (res.valid) {
        resolve(true);
      }
    });
  });
}

defineExpose({
  validator,
});

function handleChangeAuthType(e: Event) {
  const event = e as HTMLEvent<HTMLSelectElement>;
  const isWidgetValuesExist = props.keycloakSettings?.widget && Object.keys(props.keycloakSettings?.widget).length > 0;
  if (event.target.value === CitizenAuthType.widget && !isWidgetValuesExist) {
    setFieldValue('widgetUrl', defaultValues.widget.url);
    setFieldValue('widgetHeight', defaultValues.widget.height);
  }
}

watch(copyFromAuthWidget as Ref<boolean>, (newValue, oldValue) => {
  if (newValue === true && oldValue === false) {
    setFieldValue('signWidgetUrl', url.value as string);
    setFieldValue('signWidgetHeight', widgetHeight.value as number);
  }
});
watch([url, widgetHeight], () => {
  if (copyFromAuthWidget.value === true) {
    setFieldValue('signWidgetUrl', url.value as string);
    setFieldValue('signWidgetHeight', widgetHeight.value as number);
  }
});
const preparedValues = computed(() => ({
  edrCheck: values.edrCheckEnabled,
  authType: values.authType,
  widget: {
    url: values.widgetUrl,
    height: values.widgetHeight
  },
  registryIdGovUa: {
    clientId: values.clientId,
    clientSecret: values.secret,
    url: values.idGovUaUrl,
  },
  citizenPortal: {
    signWidget: {
      copyFromAuthWidget: values.copyFromAuthWidget,
      url: values.signWidgetUrl,
      height: values.signWidgetHeight,
    }
  }
}));
</script>

<template>
  <input
    type="hidden"
    name="registry-citizen-auth"
    :value="JSON.stringify(preparedValues)"
  />
  <Typography variant="h3">Автентифікація отримувачів послуг</Typography>
  <Typography variant="h5" upper-case class="subheading">Перевірка даних в ЄДР</Typography>
  <Typography variant="bodyText" class="mb16">
    Перевірка даних з КЕП користувачів в ЄДР відбувається за умови налаштованої інтеграції поточного реєстру з ЄДР через
    ШБО Трембіта.
  </Typography>
  <div class="wizard-warning">Відключення перевірки в ЄДР доступне для версій реєстру 1.9.4 і вище.</div>
  <div class="toggle-switch">
    <input class="switch-input" type="checkbox" id="edr-check-input" name="edr-check-enabled"
           v-model="edrCheckEnabled"/>
    <label for="edr-check-input">Toggle</label>
    <span>Перевіряти наявність активного запису в ЄДР для бізнес-користувачів</span>
  </div>
  <Typography variant="h5" upper-case class="subheading">тип автентифікації</Typography>
  <Typography variant="bodyText" class="mb16">Є можливість використовувати власний віджет автентифікації або налаштувати інтеграцію з id.gov.ua.</Typography>
  <div class="rc-form-group" :class="{'error': !!errors.authType}">
      <label for="rec-auth-type">Вкажіть тип автентифікації</label>
      <select
        name="rec-auth-browser-flow" id="rec-auth-type"
        v-model="authType"
        @change="handleChangeAuthType"
      >
        <option selected :value="CitizenAuthType.widget">Віджет</option>
        <option :value="CitizenAuthType.platformIdGovUa">Платформенна інтеграція з id.gov.ua</option>
      </select>
      <span v-if="!!errors.authType">{{ getErrorMessage(errors.authType) }}</span>
  </div>
  <div v-if="authType === CitizenAuthType.widget">
    <TextField
      required
      label="Посилання"
      name="rec-auth-url"
      :error="errors.widgetUrl"
      v-model="url"
      description="URL, повинен починатись з http:// або https://"
    />
    <TextField
      required
      type="number"
      label="Висота віджета, px"
      name="rec-auth-widget-height"
      :error="errors.widgetHeight"
      v-model="widgetHeight"
    />
  </div>

  <div v-if="authType === CitizenAuthType.registryIdGovUa">
    <TextField
      required
      label="Посилання"
      name="rec-id-gov-ua-url"
      :error="errors.idGovUaUrl"
      v-model="idGovUaUrl"
      description="URL, повинен починатись з http:// або https://"
    />
    <TextField
      required
      label="Ідентифікатор клієнта (client_id)"
      name="rec-auth-client-id"
      v-model="clientId"
      :error="errors.clientId"
    />
    <TextField
      required
      label="Клієнтський секрет (secret)"
      name="rec-auth-client-secret"
      v-model="secret"
      :error="errors.secret"
      type="password"
    />
  </div>
  <Typography variant="h5" upper-case class="subheading">Віджет підпису документів</Typography>
  <div v-if="authType === CitizenAuthType.widget">
      <div class="toggle-switch">
      <input class="switch-input" type="checkbox" id="sign-widget-copy" name="sign-widget-copy"
            v-model="copyFromAuthWidget" />
      <label for="sign-widget-copy">Toggle</label>
      <span>Використовувати налаштування віджету автентифікації</span>
    </div>
  </div>
  <TextField
      v-if="!copyFromAuthWidget"
      root-class="mt16"
      required
      label="Посилання"
      name="rec-sign-widget-url"
      :error="errors.signWidgetUrl"
      v-model="signWidgetUrl"
      description="URL, повинен починатись з http:// або https://"
    />
    <TextField
      v-if="!copyFromAuthWidget"
      required
      type="number"
      label="Висота віджета, px"
      name="rec-sign-widget-height"
      :error="errors.signWidgetHeight"
      v-model="signWidgetHeight"
    />
</template>

<style scoped>
  .mb16 {
    margin-bottom: 16px;
  }
  .mt16 {
    margin-top: 16px;
  }
  .subheading {
    margin-top: 32px;
    margin-bottom: 32px;
  }
</style>
