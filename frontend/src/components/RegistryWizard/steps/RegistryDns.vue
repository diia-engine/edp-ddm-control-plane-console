<script setup lang="ts">
import { toRefs, ref } from 'vue';
import * as Yup from 'yup';
import axios from 'axios';
import { useField, useForm } from 'vee-validate';
import TextField from '@/components/common/TextField.vue';
import Typography from '@/components/common/Typography.vue';
import ToggleSwitch from '@/components/common/ToggleSwitch.vue';
import type { OfficerPortalSettings, PortalSettings } from '@/types/registry';

type RegistryDnsProps =  {
  dnsManual: string, 
  keycloakHostname: string, 
  keycloakHostnames: string[], 
  keycloakCustomHost: string,
  portals: {
    citizen: PortalSettings
    officer: OfficerPortalSettings,
  }
}

interface FormValues {
  officerDnsEnabled: boolean;
  officerDns: string;
  officerSsl: any;
  citizenDnsEnabled: boolean;
  citizenDns: string;
  citizenSsl: any;
}

const REGEX_DNS_NAME =
  /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9/-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9/-]*[A-Za-z0-9])$/;
const props = defineProps<RegistryDnsProps>();
const { dnsManual, keycloakHostname, keycloakHostnames, keycloakCustomHost, portals } = toRefs(props);
const errorsFile = ref<Record<string, boolean>>({
  officerSsl: false,
  citizenSsl: false,
});


const validationSchema = Yup.object<FormValues>({
  officerDnsEnabled: Yup.bool(),
  officerDns: Yup.string().when('officerDnsEnabled', {
    is: true,
    then: (schema) => schema.required().max(63).matches(REGEX_DNS_NAME),
  }),
  officerSsl: Yup.mixed().when('officerDnsEnabled', {
    is: (enabled: boolean) => {
      return enabled && portals.value?.officer.customDns?.host !== officerDns.value;
    },
    then: (schema) => schema.required().test({
      message: 'invalidFileType',
      test: function () {
        return !(errorsFile.value?.officerSsl);
      },
    }),
    otherwise: (schema) => schema.notRequired(),
  }),
  citizenDnsEnabled: Yup.bool(),
  citizenDns: Yup.string().when('citizenDnsEnabled', {
    is: true,
    then: (schema) => schema.required().max(63).matches(REGEX_DNS_NAME),
  }),
  citizenSsl: Yup.mixed().when('citizenDnsEnabled', {
    is: (enabled: boolean) => {
      return enabled && portals.value?.citizen.customDns?.host !== citizenDns.value;
    },
    then: (schema) => schema.required().test({
      message: 'invalidFileType',
      test: function () {
        return !(errorsFile.value?.citizenSsl);
      },
    }),
    otherwise: (schema) => schema.notRequired(),
  }),
});

const { errors, validate } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    officerDnsEnabled: portals.value?.officer?.customDns?.enabled || false,
    officerDns: portals.value?.officer.customDns?.host || '',
    citizenDnsEnabled: portals.value?.citizen.customDns?.enabled || false,
    citizenDns: portals.value?.citizen.customDns?.host || '',
  }
});

const { value: officerDnsEnabled } = useField<boolean>('officerDnsEnabled');
const { value: officerDns } = useField<string>('officerDns');
const { value: officerSsl } = useField<string | null>('officerSsl');
const { value: citizenDnsEnabled } = useField<boolean>('citizenDnsEnabled');
const { value: citizenDns } = useField<string>('citizenDns');
const { value: citizenSsl } = useField<string | null>('citizenSsl');

function validator() {
  return new Promise((resolve) => {
    validate().then(async (res) => {
      if (res.valid) {
        resolve(true);
      }
    });
  });
}

async function handleFileField(e: Event) {
  const fileName = (e.target as HTMLInputElement).dataset.attr as string;
  try {
    const file = (e.target as HTMLInputElement).files![0];
    const formData = new FormData();
    formData.append('file', file);
    await axios.post('/admin/registry/check-pem', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    errorsFile.value[fileName] = false;
  } catch {
    errorsFile.value = {
      ...errorsFile.value,
      [fileName]: true,
    };
  }
}

function handleChangeOfficerDnsEnabled(enabled: boolean) {
  if (!enabled) {
    officerSsl.value = null;
  }
}

function handleChangeCitizenDnsEnabled(enabled: boolean) {
  if (!enabled) {
    citizenSsl.value = null;
  }
}

defineExpose({
  validator,
});
</script>

<template>
  <Typography variant="h3" class="mb24">Налаштування DNS</Typography>

  <div class="rc-form-group show-errors-span dns-inputs">
    <label for="keycloak-hostname" class="header-label">
      Сервіс управління користувачами та ролями (Keycloak)
    </label>

    <div class="text-input-label label-ssl">Доменне імʼя для Keycloak</div>
    <select id="keycloak-hostname" name="keycloak-custom-hostname">
      <option>{{ keycloakHostname }}</option>
      <option
        v-for="hn in keycloakHostnames"
        :selected="hn === keycloakCustomHost"
        v-bind:key="hn"
      >
        {{ hn }}
      </option>
    </select>
  </div>

  <div class="rc-form-group show-errors-span dns-inputs">
    <Typography upperCase variant="h5" class="mb24">
      Кабінет посадової особи
    </Typography>
    <ToggleSwitch
      name="officer-dns-enabled"
      label="Використати власні значення"
      v-model="officerDnsEnabled"
      @change="handleChangeOfficerDnsEnabled"
      classes="mb24"
    />
    <template v-if="officerDnsEnabled">
      <TextField
        required
        label="Доменне імʼя для кабінету посадової особи"
        name="officer-dns"
        v-model="officerDns"
        :error="errors?.officerDns"
      />
      <TextField
        required
        type="file"
        label="SSL-сертифікат для кабінету посадової особи (*.pem)"
        name="officer-ssl"
        v-model="officerSsl"
        :error="errors?.officerSsl"
        rootClass="input-file"
        @change="handleFileField"
        data-attr="officerSsl"
        accept=".pem"
      />
    </template>
  </div>

  <div class="rc-form-group show-errors-span dns-inputs">
    <Typography upperCase variant="h5" class="mb24">
      Кабінет отримувача послуг
    </Typography>
    <ToggleSwitch
      name="citizen-dns-enabled"
      label="Використати власні значення"
      v-model="citizenDnsEnabled"
      @change="handleChangeCitizenDnsEnabled"
      classes="mb24"
    />
    <template v-if="citizenDnsEnabled">
      <TextField
        required
        label="Доменне імʼя для кабінету отримувача послуг"
        name="citizen-dns"
        v-model="citizenDns"
        :error="errors?.citizenDns"
      />
      <TextField
        required
        type="file"
        label="SSL-сертифікат для кабінету отримувача послуг (*.pem)"
        name="citizen-ssl"
        v-model="citizenSsl"
        :error="errors?.citizenSsl"
        rootClass="input-file"
        @change="handleFileField"
        data-attr="citizenSsl"
        accept=".pem"
      />
    </template>
  </div>

  <div class="yellow-banner">
    <div class="banner-title">Увага!</div>
    <div class="banner-body">
      Необхідно виконати додаткову зовнішню конфігурацію за межами OpenShift
      кластеру та реєстру.
      <template v-if="dnsManual">
        Інструкція доступна
        <a target="_blank" :href="dnsManual">за посиланням</a>.
      </template>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.mb24 {
  margin-bottom: 24px;
}

.input-file :deep(input[type='file']) {
  height: 47px;
}
</style>
