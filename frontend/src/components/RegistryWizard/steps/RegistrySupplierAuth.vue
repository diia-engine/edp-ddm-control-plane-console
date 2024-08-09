<script lang="ts" setup>
import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
import ToggleSwitch from '@/components/common/ToggleSwitch.vue';
import { ref } from 'vue';
import { useForm, useField } from 'vee-validate';
import * as Yup from 'yup';
import { OfficerAuthType, type CitizenAuthFlow, PORTALS } from '@/types/registry';

interface FormValues {
  authType: OfficerAuthType,
  url: string,
  widgetHeight: number,
  clientId: string,
  secret: string,
  individualAccessEnabled: boolean,
}
interface RegistryRecipientAuthProps {
  keycloakSettings: {
    authFlows: {
      officerAuthFlow: {
        widgetHeight: number
      }
    }
    citizenAuthFlow: CitizenAuthFlow
    customHost: string
    identityProviders: {
      idGovUa: {
        clientId: string
        url: string
        secretKey: string
      }
    }
    realms: {
      officerPortal: {
        browserFlow: string
        selfRegistration: boolean
      }
    }
  },
  signWidgetSettings: {
    url: string
  },
  officerPortalSettings: {
    individualAccessEnabled: boolean,
  },
  isEnabledPortal: boolean;
}

const props = defineProps<RegistryRecipientAuthProps>();
const isSecretExists = props.keycloakSettings?.identityProviders?.idGovUa?.secretKey.length > 0;
const selfRegistrationEnabled = ref(props.keycloakSettings?.realms?.officerPortal?.selfRegistration || false);
const isEnabledPortal = ref(props.isEnabledPortal);
const portal = ref(props.isEnabledPortal ? '' : PORTALS.officer);
const defaultValues = {
  authType: OfficerAuthType.widget,
  url: "https://eu.iit.com.ua/sign-widget/v20200922/",
  widgetHeight: 720,
  clientId: "",
  secret: "",
  individualAccessEnabled: false,
};

const validationSchema = Yup.object<FormValues>({
  authType: Yup.string()
    .required()
    .oneOf([OfficerAuthType.registryIdGovUa, OfficerAuthType.widget]),
  url: Yup.string().required().url(),
  widgetHeight: Yup.number()
  .when('authType', {
    is: (value: OfficerAuthType) => value === OfficerAuthType.widget,
    then: (schema) => schema.required().min(1, 'required').integer().typeError('wrongFormat'),
  }),
  clientId: Yup.string()
    .when('authType', {
      is: (value: OfficerAuthType) => value === OfficerAuthType.registryIdGovUa,
      then: (schema) => schema.required(),
    }),
  secret: Yup.string()
    .when('authType', {
      is: (value: OfficerAuthType) => value === OfficerAuthType.registryIdGovUa && !isSecretExists,
      then: (schema) => schema.required(),
    }),
});

const { errors, validate, setFieldValue } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    authType: props.keycloakSettings?.realms?.officerPortal?.browserFlow as OfficerAuthType.widget || defaultValues.authType,
    url: (
      props.keycloakSettings?.realms?.officerPortal?.browserFlow === OfficerAuthType.widget
      ? props.signWidgetSettings?.url
      : props.keycloakSettings?.identityProviders.idGovUa.url
      ) || defaultValues.url,
    widgetHeight: props.keycloakSettings?.authFlows?.officerAuthFlow?.widgetHeight ?? defaultValues.widgetHeight,
    clientId: props.keycloakSettings?.identityProviders.idGovUa.clientId || defaultValues.clientId,
    secret: defaultValues.secret,
    individualAccessEnabled: props.officerPortalSettings?.individualAccessEnabled || defaultValues.individualAccessEnabled
  }
});

const { value: authType } = useField('authType');
const { value: url } = useField('url');
const { value: widgetHeight } = useField('widgetHeight');
const { value: clientId } = useField('clientId');
const { value: secret } = useField('secret');
const { value: individualAccessEnabled } = useField('individualAccessEnabled');

function validator() {
  return new Promise((resolve) => {
    validate().then((res) => {
      if (res.valid) {
        resolve(true);
      }
    });
  });
}

function handleEnabledPortalChange(enabled: boolean) {
  portal.value = enabled ? '' : PORTALS.officer;
}

defineExpose({
  validator,
});

function handleChangeAuthType() {
  setFieldValue('secret', "");
  if (authType.value === OfficerAuthType.widget) {
    if (props.signWidgetSettings?.url !== "") {
      setFieldValue('url', props.signWidgetSettings?.url);
    }
    else {
      setFieldValue('url', defaultValues.url);
    }
    if (props.keycloakSettings.authFlows.officerAuthFlow.widgetHeight !== 0) {
      setFieldValue('widgetHeight', props.keycloakSettings.authFlows.officerAuthFlow.widgetHeight);
    }
  }
  else {
    setFieldValue('url', props.keycloakSettings.identityProviders.idGovUa.url);
    if (props.keycloakSettings.identityProviders.idGovUa.clientId !== "") {
      setFieldValue('clientId', props.keycloakSettings.identityProviders.idGovUa.clientId);
    }
  }
}

</script>

<template>
  <Typography variant="h3" class="heading">Кабінет надавача послуг</Typography>
  <input type="hidden" name="excludePortals[]" :value="portal"/>
  <ToggleSwitch
    name="enabledOfficerPortal"
    label="Розгорнути Кабінет надавача послуг"
    v-model="isEnabledPortal"
    @change="handleEnabledPortalChange"
  />
  <template v-if="isEnabledPortal">
    <div>
      <Typography variant="h5" upper-case class="subheading">Управління доступом</Typography>
      <Typography variant="bodyText" class="mb16">Налаштування доступу користувачам до Кабінету користувача з використанням КЕП фізичної особи.</Typography>
      <div class="toggle-switch backup-switch">
        <input v-model="individualAccessEnabled" class="switch-input"
              type="checkbox" id="rec-individual-access-enabled" name="rec-individual-access-enabled" />
        <label for="rec-individual-access-enabled">Toggle</label>
        <span>Дозволити доступ з КЕП фізичної особи</span>
      </div>
    </div>
    <Typography variant="bodyText" class="mb16">
      Є можливість використовувати власний віджет автентифікації або налаштувати інтеграцію з id.gov.ua.
    </Typography>

    <div class="rc-form-group">
        <label for="sup-auth-type">Вкажіть тип автентифікації</label>
        <select name="sup-auth-browser-flow" id="sup-auth-type"
                v-model="authType" @change="handleChangeAuthType">
            <option value="dso-officer-auth-flow">Віджет</option>
            <option value="id-gov-ua-officer-redirector">id.gov.ua</option>
        </select>
    </div>

    <TextField
      required
      label="Посилання"
      name="sup-auth-url"
      :error="errors.url"
      v-model="url"
      description="URL, повинен починатись з http:// або https://"
    />

    <div v-if="authType === OfficerAuthType.widget">
      <TextField
        required
        type="number"
        label="Висота віджета, px"
        name="sup-auth-widget-height"
        :error="errors.widgetHeight"
        v-model="widgetHeight"
      />
    </div>

    <div v-if="authType === OfficerAuthType.registryIdGovUa">
      <TextField
        required
        label="Ідентифікатор клієнта (client_id)"
        name="sup-auth-client-id"
        :error="errors.clientId"
        v-model="clientId"
      />
      <TextField
        required
        label="Клієнтський секрет (secret)"
        name="sup-auth-client-secret"
        v-model="secret"
        :error="errors.secret"
        type="password"
        placeholder="******"
      />
    </div>

    <div class="rc-self-registration">
      <Typography variant="h5" upper-case class="subheading">Самостійна реєстрація користувачів</Typography>
      <Typography variant="bodyText" class="mb16">Передбачає наявність у реєстрі попередньо змодельованого бізнес-процесу самореєстрації.</Typography>
      <div class="toggle-switch backup-switch">
        <input v-model="selfRegistrationEnabled" class="switch-input"
              type="checkbox" id="self-registration-switch-input" name="self-registration-enabled" />
        <label for="self-registration-switch-input">Toggle</label>
        <span>Дозволити самостійну реєстрацію</span>
      </div>
      <div class="wizard-warning" v-if="selfRegistrationEnabled">
        При вимкненні можливості, користувачі, які почали процес самореєстрації, не зможуть виконати свої задачі, якщо вони змодельовані.
      </div>
    </div>
  </template>
  
</template>

<style scoped>
  .rc-self-registration {
    margin-top: 32px;
  }
  .mb16 {
    margin-bottom: 16px;
  }
  
  .heading {
    margin-bottom: 24px;
  }

  .subheading {
    margin-top: 32px;
    margin-bottom: 24px;
  }
</style>