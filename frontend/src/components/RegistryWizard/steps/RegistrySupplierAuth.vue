<script lang="ts" setup>
import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
import ToggleSwitch from '@/components/common/ToggleSwitch.vue';
import { computed, ref } from 'vue';
import { useForm, useField } from 'vee-validate';
import * as Yup from 'yup';
import { OfficerAuthType, type CitizenAuthFlow, PORTALS } from '@/types/registry';
import type { StoredKey } from '@/types/cluster';
import SelectVue from '@/components/common/Select.vue';
import { getErrorMessage } from '@/utils';
import { filterKeysByRegistry } from '@/utils/registry';
import Modal from '@/components/common/Modal.vue';

interface FormValues {
  authType: OfficerAuthType,
  url: string,
  widgetHeight: number,
  clientId: string,
  secret: string,
  individualAccessEnabled: boolean,
  singleIdentityEnabled: boolean,
  keyName: string,
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
        keyName: string
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
    singleIdentityEnabled: boolean,
  },
  isEnabledPortal: boolean,
  digitalSignatureKeys: Record<string, StoredKey> | null
  region: string;
  registryName: string;
}

const props = defineProps<RegistryRecipientAuthProps>();
const isSecretExists = props.keycloakSettings?.identityProviders?.idGovUa?.secretKey.length > 0;
const selfRegistrationEnabled = ref(props.keycloakSettings?.realms?.officerPortal?.selfRegistration || false);
const isEnabledPortal = ref(props.isEnabledPortal);
const portal = ref(props.isEnabledPortal ? '' : PORTALS.officer);
const showSingleIdentityPopUp = ref(false);
const defaultValues = {
  authType: OfficerAuthType.widget,
  url: "https://eu.iit.com.ua/sign-widget/v20240301/",
  widgetHeight: 720,
  clientId: "",
  secret: "",
  individualAccessEnabled: false,
  keyName: '',
  singleIdentityEnabled: false,
};

const storedKeyItems = computed<string[]>(() => props.digitalSignatureKeys ?
  filterKeysByRegistry(props.digitalSignatureKeys, props.registryName)
: []);

const validationSchema = Yup.object<FormValues>({
  authType: Yup.string()
    .required()
    .test('dsKeysNotFound', 'dsKeysNotFound',  (value) => {
      if (value === OfficerAuthType.registryIdGovUa && storedKeyItems.value.length === 0) {
        return false;
      }
      return true;
    }),
  url: Yup.string().required().url(),
  widgetHeight: Yup.number()
  .when('authType', {
    is: (value: OfficerAuthType) => value === OfficerAuthType.widget,
    then: (schema) => schema.required().min(1, 'checkFormat').integer().typeError('wrongFormat').positive(),
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
  keyName: Yup.string()
    .when('authType', {
      is: (value: OfficerAuthType) => value === OfficerAuthType.registryIdGovUa,
      then: (schema) => schema.required(),
    }),
});

const getCurrentKeyName = () => {
  const propsKeyName = props.keycloakSettings?.identityProviders?.idGovUa.keyName;
  // is stored key exists in available keys
  if (propsKeyName && props.digitalSignatureKeys?.[propsKeyName]) {
    return propsKeyName;
  }
  return defaultValues.keyName;
};

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
    keyName: getCurrentKeyName(),
    secret: defaultValues.secret,
    individualAccessEnabled: props.officerPortalSettings?.individualAccessEnabled || defaultValues.individualAccessEnabled,
    singleIdentityEnabled: props.officerPortalSettings?.singleIdentityEnabled || defaultValues.singleIdentityEnabled
  }
});

const { value: authType } = useField('authType');
const { value: url } = useField('url');
const { value: widgetHeight } = useField('widgetHeight');
const { value: clientId } = useField('clientId');
const { value: secret } = useField('secret');
const { value: individualAccessEnabled } = useField('individualAccessEnabled');
const { value: singleIdentityEnabled } = useField('singleIdentityEnabled');
const { value: keyName } = useField('keyName');

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
  <Typography variant="h3" class="heading">{{ $t('components.registrySupplierAuth.title') }}</Typography>
  <input type="hidden" name="excludePortals[]" :value="portal"/>
  <ToggleSwitch
    name="enabledOfficerPortal"
    :label="$t('components.registrySupplierAuth.fields.enabledOfficerPortal.label')"
    v-model="isEnabledPortal"
    @change="handleEnabledPortalChange"
  />
  <template v-if="isEnabledPortal">
    <div v-if="region === 'ua'">
      <div>
        <Typography variant="h5" upper-case class="subheading">{{ $t('components.registrySupplierAuth.text.accessControl') }}</Typography>
        <Typography variant="bodyText" class="mb16">{{ $t('components.registrySupplierAuth.text.settingAccessUsers') }}</Typography>
        <div class="toggle-switch backup-switch">
          <input v-model="individualAccessEnabled" class="switch-input"
          type="checkbox" id="rec-individual-access-enabled" name="rec-individual-access-enabled" />
          <label for="rec-individual-access-enabled">Toggle</label>
          <span>{{ $t('components.registrySupplierAuth.text.allowAccessFromKEP') }}</span>
        </div>
        <Typography variant="bodyText" class="mb16">
          {{ $t('components.registrySupplierAuth.text.singleIdentityDescription') }} 
          <a href="#" @click.stop.prevent="showSingleIdentityPopUp = true">{{ $t('components.registrySupplierAuth.actions.moreDetails') }}</a>.
        </Typography>
        <div class="toggle-switch backup-switch">
          <input v-model="singleIdentityEnabled" class="switch-input"
                type="checkbox" id="rec-enable-single-identity" name="rec-enable-single-identity" />
          <label for="rec-enable-single-identity">Toggle</label>
          <span>{{ $t('components.registrySupplierAuth.text.singleIdentityEnabled') }}</span>
        </div>
      </div>
      <Typography variant="bodyText" class="mb16">
        {{ $t('components.registrySupplierAuth.text.possibleUseAuthenticationWidget') }}
      </Typography>

      <div class="rc-form-group" :class="{'error': !!errors.authType}">
          <label for="sup-auth-type">{{ $t('components.registrySupplierAuth.text.specifyAuthenticationType') }}</label>
          <select name="sup-auth-browser-flow" id="sup-auth-type"
                  v-model="authType" @change="handleChangeAuthType">
              <option :value="OfficerAuthType.widget">{{ $t('components.registrySupplierAuth.text.widget') }}</option>
              <option :value="OfficerAuthType.registryIdGovUa">{{ $t('components.registrySupplierAuth.text.registryIntegration') }}</option>
          </select>
        <span v-if="!!errors.authType">{{ getErrorMessage(errors.authType) }}</span>
      </div>

      <TextField required
        v-if="authType === OfficerAuthType.widget || (authType === OfficerAuthType.registryIdGovUa && storedKeyItems.length > 0)"
        :label="$t('components.registrySupplierAuth.fields.authURL.label')"
        name="sup-auth-url"
        :error="errors.url"
        v-model="url"
        :description="$t('components.registrySupplierAuth.fields.authURL.description')"
      />

      <div v-if="authType === OfficerAuthType.widget">
        <TextField
          required
          type="number"
          :label="$t('components.registrySupplierAuth.fields.widgetHeight.label')"
          name="sup-auth-widget-height"
          :error="errors.widgetHeight"
          v-model="widgetHeight"
        />
      </div>

    <div v-if="authType === OfficerAuthType.registryIdGovUa && storedKeyItems.length > 0">
      <TextField
        required
        :label="$t('components.registrySupplierAuth.fields.clientID.label')"
        name="sup-auth-client-id"
        :error="errors.clientId"
        v-model="clientId"
      />
      <TextField
        required
        :label="$t('components.registrySupplierAuth.fields.clientSecret.label')"
        name="sup-auth-client-secret"
        v-model="secret"
        :error="errors.secret"
        type="password"
        placeholder="******"
      />
      <SelectVue
        required
        :label="$t('components.registrySupplierAuth.fields.keyName.label')"
        :items="storedKeyItems"
        v-model="keyName"
        :error="errors.keyName"
        name="sup-auth-key-name"
        :description="$t('components.registrySupplierAuth.fields.keyName.description')"
       />
      </div>
    </div>

    <div class="rc-self-registration">
      <Typography variant="h5" upper-case class="subheading">{{ $t('components.registrySupplierAuth.text.selfRegistrationUsers') }}</Typography>
      <Typography variant="bodyText" class="mb16">{{ $t('components.registrySupplierAuth.text.presenceSelfRegistrationBusiness') }}</Typography>
      <div class="toggle-switch backup-switch">
        <input v-model="selfRegistrationEnabled" class="switch-input"
              type="checkbox" id="self-registration-switch-input" name="self-registration-enabled" />
        <label for="self-registration-switch-input">Toggle</label>
        <span>{{ $t('components.registrySupplierAuth.text.allowSelfRegistration') }}</span>
      </div>
      <div class="wizard-warning" v-if="selfRegistrationEnabled">
        {{ $t('components.registrySupplierAuth.text.notAbleCompleteTasksSelfRegistration') }}
      </div>
    </div>
    <Modal
      :title="$t('components.registrySupplierAuth.text.singleIdentityEnabled')"
      :submitBtnText="$t('actions.gotIt')"
      :hasCancelBtn=false
      :show="showSingleIdentityPopUp"
      @close="showSingleIdentityPopUp = false"
      @submit="showSingleIdentityPopUp = false"
    >
      <p>{{ $t('components.registrySupplierAuth.text.singleIdentityDescriptionPopup') }}</p>
    </Modal>
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
