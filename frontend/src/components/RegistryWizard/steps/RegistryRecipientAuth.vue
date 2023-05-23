<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
import { getErrorMessage } from '@/utils';
import { computed } from 'vue';
import { useForm } from 'vee-validate';
import * as Yup from 'yup';

interface HTMLEvent<T extends EventTarget = HTMLElement> extends Event {
  target: T
}

enum CitizenAuthType {
  widget = 'widget',
  registryIdGovUa = 'registry-id-gov-ua',
  platformIdGovUa = 'platform-id-gov-ua',
}

interface OutFormValues {
  edrCheck: boolean
  authType: CitizenAuthType
  widget: {
    url: string
    height: number
  }
}

interface FormValues {
  edrCheckEnabled: boolean,
  authType: CitizenAuthType,
  url: string,
  widgetHeight: number,
  clientId: string,
  secret: string,
}
interface RegistryRecipientAuthProps {
  data: OutFormValues,
}


const validationSchema = Yup.object<FormValues>({
  authType: Yup.string()
    .required(),
  url: Yup.string()
    .when('authType', {
      is: (value: CitizenAuthType) => value === CitizenAuthType.widget,
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
      is: (value: CitizenAuthType) => value === CitizenAuthType.registryIdGovUa,
      then: (schema) => schema.required(),
    })
});
const defaultValues: OutFormValues = {
  authType: CitizenAuthType.widget,
  edrCheck: true,
  widget: {
    url: 'https://eu.iit.com.ua/sign-widget/v20200922/',
    height: 720,
  }
};
const props = defineProps<RegistryRecipientAuthProps>();

const { errors, useFieldModel, validate, values, setFieldValue } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    edrCheckEnabled: props.data?.edrCheck ?? defaultValues.edrCheck,
    authType: props.data?.authType || defaultValues.authType,
    url: props.data?.widget.url ?? defaultValues.widget.url,
    widgetHeight: props.data?.widget.height ?? defaultValues.widget.height,
  }
});

const [
  edrCheckEnabled,
  authType,
  url,
  widgetHeight,
  clientId,
  secret
] = useFieldModel(['edrCheckEnabled', 'authType', 'url', 'widgetHeight', 'clientId', 'secret']);

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
  const isWidgetValuesExist = props.data?.widget && Object.keys(props.data?.widget).length > 0;
  if (event.target.value === CitizenAuthType.widget && !isWidgetValuesExist) {
    setFieldValue('url', defaultValues.widget.url);
    setFieldValue('widgetHeight', defaultValues.widget.height);
  }
}
const preparedValues = computed((): OutFormValues => ({
  edrCheck: values.edrCheckEnabled,
  authType: values.authType,
  widget: {
    url: values.url,
    height: values.widgetHeight
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
      :error="errors.url"
      v-model="url"
      description="URL, повинен починатись з http:// або https://"
    />
    <TextField
      required
      label="Висота віджета, px"
      name="rec-auth-widget-height"
      :error="errors.widgetHeight"
      v-model="widgetHeight"
    />
  </div>

  <div v-if="authType === CitizenAuthType.registryIdGovUa">
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
</template>

<style scoped>
  .mb16 {
    margin-bottom: 16px;
  }
  .subheading {
    margin-top: 32px;
    margin-bottom: 32px;
  }
</style>
