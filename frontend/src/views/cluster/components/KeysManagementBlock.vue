<script setup lang="ts">
import ButtonVue from '@/components/common/Button.vue';
import Typography from '@/components/common/Typography.vue';
import Modal from '@/components/common/Modal.vue';
import TextField from '@/components/common/TextField.vue';
import SelectVue from '@/components/common/Select.vue';
import FileField from '@/components/common/FileField.vue';
import { ref, watch } from 'vue';
import * as Yup from 'yup';
import { useField, useForm } from 'vee-validate';
import { computed } from 'vue';
import { KeyType, type TableViewKey, type FileKey, type HardwareKey, type StoredKey, type PreparedHardKey, type PreparedFileKey } from '@/types/cluster';
import KeysTable from './KeysManagement/KeysTable.vue';
import { KEY_VARIANTS } from '@/constants/key';
import OsplmIniEditor from '@/utils/osplmIniEditor';
import Banner from '@/components/common/Banner.vue';
import { isArray } from 'lodash';

interface KeysManagementBlockProps {
  iniTemplate: string,
  digitalSignature?: {
    keys: Record<string, StoredKey>
    data: Record<string, string>
    env: Record<string, string>
  }
  usedKeys: Record<string, string[]>
  registries: string[]
}

type KeyForm = {
  deviceType: KeyType,
} & FileKey & HardwareKey;

type StateKey = PreparedHardKey | PreparedFileKey | HardwareKey | FileKey;

const props = defineProps<KeysManagementBlockProps>();

const keys = ref(Object.entries(props.digitalSignature?.keys || {})
  .map(([key, value]) => ({
    deviceType: value['device-type'],
    allowedRegistries: value.allowedRegistries,
    ...( value['device-type'] === KeyType.file ? {
      fileKeyName: key,
      fileKeyIssuer: value.issuer,
      fileKeyFile: value.file,
      fileKeyPassword: value.password,
    } : {
      hardKeyName: key,
      hardKeyIssuer: value.issuer,
      hardKeyPassword: value.password,
      hardKeyType: value.type,
      hardKeyDevice: value.device,
    } )
  } as StateKey))
  // filter allowed registries in each key to only those that are present in the current cluster
  .map((key) => {

    if (!key.allowedRegistries || (isArray(key.allowedRegistries) && key.allowedRegistries.length === 0)) {
      return key;
    }
    const isKeyAllowedRegistriesInRegistries = key.allowedRegistries.every((registry) => props.registries.includes(registry));
    if (isKeyAllowedRegistriesInRegistries) {
      return key;
    }
    return {
      ...key,
      allowedRegistries: key.allowedRegistries.filter((registry) => props.registries.includes(registry)),
    };
  }));
const addKeyModal = ref(false);
const cantRemoveKeyDialog = ref('');
const removeKeyPrompt = ref('');
const editIniModal = ref(false);
const currentKey = ref<StateKey | null>(null);

const uniqKeyNameValidator = (value: string) => !keys
          .value
          .find((k) => 'fileKeyName' in k ? k.fileKeyName === value : k.hardKeyName === value);
const keyNameRegExp = /^[a-z0-9-]+$/;
const keyFieldsSchema = Yup.object<KeyForm>({
  deviceType: Yup.string().required().oneOf([KeyType.file, KeyType.hardware]),
  fileKeyName: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.file,
    then: (schema) => schema
      .required()
      .min(3)
      .max(40)
      .test(
        'unique',
        'nonUniqKeyName',
        uniqKeyNameValidator,
      )
      .matches(keyNameRegExp),
  }),
  fileKeyIssuer: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.file,
    then: (schema) => schema.required(),
  }),
  fileKeyPassword: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.file,
    then: (schema) => schema.required(),
  }),
  fileKeyFile: Yup.mixed().nullable().when('deviceType', {
    is: (value: KeyType) => value === KeyType.file,
    then: (schema) => schema.required().nonNullable(),
  }),
  hardKeyName: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema
      .required()
      .min(3)
      .max(40)
      .test(
        'unique',
        'nonUniqKeyName',
        uniqKeyNameValidator,
      )
      .matches(keyNameRegExp),
  }),
  hardKeyType: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeyIssuer: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeyIssuerHost: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeyIssuerPort: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeySerialNumber: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeyPort: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeyHost: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeyMask: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
  hardKeyPassword: Yup.string().when('deviceType', {
    is: (value: KeyType) => value === KeyType.hardware,
    then: (schema) => schema.required(),
  }),
});

const defaultKeyData: KeyForm = {
  deviceType: KeyType.file,
  fileKeyIssuer: '',
  fileKeyFile: null,
  fileKeyPassword: '',
  fileKeyName: '',
  hardKeyIssuerHost: '',
  hardKeyIssuer: '',
  hardKeyIssuerPort: '',
  hardKeyHost: '',
  hardKeyMask: '',
  hardKeyPort: '',
  hardKeyPassword: '',
  hardKeyType: '',
  hardKeySerialNumber: '',
  hardKeyName: '',
  allowedRegistries: [],
};

const { errors, setFieldValue, values, resetForm, validate } = useForm<KeyForm>({
  initialValues: defaultKeyData,
  validationSchema: keyFieldsSchema,
});

const getKeyUsage = (keyName: string) => {
  return Object.entries(props.usedKeys)
    .filter((entry) => entry[1].includes(keyName))
    .map(([key]) => key);
};

const registriesWithEmptyKeys = Object.entries(props.usedKeys)
  .filter(([, keys]) => keys.length === 1 && keys[0] === '')
  .map(([r]) => r);

const isHardwareKey = (keyName: string) => keys.value.find((key) => 'hardKeyName' in key && key.hardKeyName === keyName);

function validator() {
  return Promise.resolve();
}
function getData() {
  return { keys: keys.value, osplmIni: hardKeyOsplmIni.value };
}

defineExpose({
  validator,
  getData,
});


const { value: deviceType } = useField<KeyType>('deviceType');
const { value: fileKeyName } = useField<KeyForm['fileKeyName']>('fileKeyName');
const { value: fileKeyIssuer } = useField<KeyForm['fileKeyIssuer']>('fileKeyIssuer');
const { value: fileKeyPassword } = useField<KeyForm['fileKeyPassword']>('fileKeyPassword');
const { value: hardKeyName } = useField<KeyForm['hardKeyName']>('hardKeyName');
const { value: hardKeyType } = useField<KeyForm['hardKeyType']>('hardKeyType');
const { value: hardKeyPassword } = useField<KeyForm['hardKeyPassword']>('hardKeyPassword');
const { value: hardKeyIssuer } = useField<KeyForm['hardKeyIssuer']>('hardKeyIssuer');
const { value: hardKeyIssuerHost } = useField<KeyForm['hardKeyIssuerHost']>('hardKeyIssuerHost');
const { value: hardKeyIssuerPort } = useField<KeyForm['hardKeyIssuerPort']>('hardKeyIssuerPort');
const { value: hardKeySerialNumber } = useField<KeyForm['hardKeySerialNumber']>('hardKeySerialNumber');
const { value: hardKeyHost } = useField<KeyForm['hardKeyHost']>('hardKeyHost');
const { value: hardKeyPort } = useField<KeyForm['hardKeyPort']>('hardKeyPort');
const { value: hardKeyMask } = useField<KeyForm['hardKeyMask']>('hardKeyMask');
const hardKeyOsplmIni = ref<string>(props.digitalSignature?.data?.['osplm.ini'] || '');
const tempIni = ref<string>('');
useField('fileKeyFile');

watch(deviceType, (newValue, oldValue) => {
  if (oldValue !== newValue && Object.keys(errors.value).length > 0) {
    resetForm({values: { deviceType: newValue } as KeyForm});
  }
});

function getKeyName(key: StateKey) {
  return 'fileKeyName' in key ? key.fileKeyName : key.hardKeyName;
}

function handleSaveKey() {
  validate().then((res) => {
    if (res.valid) {
      const hardwareKey = Object.fromEntries(
        Object.entries(values).filter(([key]) => key.startsWith('hardKey'))
      ) as HardwareKey;
      const fileKey = Object.fromEntries(
        Object.entries(values).filter(([key]) => key.startsWith('fileKey'))
      ) as FileKey;
      keys.value.push({
        ...(deviceType.value === KeyType.hardware ? hardwareKey: fileKey),
        deviceType: deviceType.value,
      });

      if (deviceType.value === KeyType.hardware) {
        const iniEditor =  new OsplmIniEditor(hardKeyOsplmIni.value || props.iniTemplate.trim());
        iniEditor.addKey({
          caHost: hardKeyIssuerHost.value,
          caPort: hardKeyIssuerPort.value,
          caName: '',
          keyHost: hardKeyHost.value,
          keySn: hardKeySerialNumber.value,
          keyMask: hardKeyMask.value,
        });
        hardKeyOsplmIni.value = iniEditor.toString();
      }
      addKeyModal.value = false;
      resetForm();
      }
  });
}
function handleFileChange(e: Event){
  const el = e.target as HTMLInputElement;
  setFieldValue('fileKeyFile', el?.files?.[0] as File);
}
function removeKey(keyName: string) {
  const keyUsage = getKeyUsage(keyName);
  if (keyUsage.length > 0) {
    cantRemoveKeyDialog.value = keyName;
    return;
  }

  if (isHardwareKey(keyName)) {
    const iniEditor =  new OsplmIniEditor(hardKeyOsplmIni.value);
    iniEditor.removeKey(keyName, keys.value);
    hardKeyOsplmIni.value = iniEditor.toString();
  }
  keys.value = keys.value.filter((key) => 'fileKeyName' in key ? key.fileKeyName !== keyName : key.hardKeyName !== keyName);
  removeKeyPrompt.value = '';
}
function handleRemoveKeyClick(keyName: string) {
  removeKeyPrompt.value = keyName;
}
function handleCloseRemoveKeyPrompt() {
  removeKeyPrompt.value = '';
}
function handleCloseModal() {
  addKeyModal.value = false;
  resetForm();
}
function handleAddKeyClick(keyType: KeyType) {
  setFieldValue('deviceType', keyType);
  addKeyModal.value = true;
}
function handleOpenIniEditor() {
  tempIni.value = hardKeyOsplmIni.value || props.iniTemplate.trim();
  editIniModal.value = true;
}
function handleSubmitIniModal() {
  hardKeyOsplmIni.value = tempIni.value;
  editIniModal.value = false;
}
function handleSaveAllowedRegistries() {
  keys.value = keys.value.map((key) => {
    if (getKeyName(key) === getKeyName(currentKey.value!)) {
      return {
        ...key,
        allowedRegistries: currentKey.value!.allowedRegistries,
      };
    }
    return key;
  });
  currentKey.value = null;
}
function handleEditKeyClick(keyName: string) {
  currentKey.value = {...keys.value.find((key) => getKeyName(key) === keyName)! };
}
function checkDisabledItems(registry: string) {
  const isKeyUsedInRegistry = props.usedKeys[registry] && isArray(props.usedKeys[registry]) && props.usedKeys[registry].includes(getKeyName(currentKey.value!));
  const isRegistryInAllowed = isArray(currentKey.value!.allowedRegistries) && currentKey.value!.allowedRegistries.includes(registry);
  return {
    disabled: isKeyUsedInRegistry && isRegistryInAllowed,
    title: registry,
    value: registry,
  };
}

const preparedKeys = computed((): TableViewKey[]  => keys.value
  .map(el => ({
    deviceType: el.deviceType,
    issuer: 'fileKeyIssuer' in el ? el.fileKeyIssuer : el.hardKeyIssuer,
    name: 'fileKeyName' in el ? el.fileKeyName : el.hardKeyName,
    allowedRegistries: el.allowedRegistries || [],
  }))
  .sort((a, b) => a.name.localeCompare(b.name))
);
const fileKeys = computed(() => preparedKeys.value.filter((key) => (key as any).deviceType === KeyType.file));
const hardwareKeys = computed(() => preparedKeys.value.filter((key) => (key as any).deviceType === KeyType.hardware));
</script>
<template>
  <div class="mb24">
    <Typography variant="h2" class="mb24 w512">{{$t("components.keysManagement.title")}}</Typography>
    <Banner v-show="registriesWithEmptyKeys.length" class="mb24" :description="$t('components.keysManagement.deprecatedKeysSyntax', { registries: registriesWithEmptyKeys.join(', ') })" />
    <Typography variant="bodyText" class="mb24 w512">{{$t("components.keysManagement.subTitle")}}</Typography>
 
    <KeysTable
      :title="$t('components.keysManagement.table.fileKeysTitle')"
      :keys="fileKeys"
      @onRemoveKeyClick="handleRemoveKeyClick"
      @onEditKeyClick="handleEditKeyClick"
    />

    <ButtonVue fa-icon-class="fa-solid fa-plus" :class="!hardwareKeys.length && 'inline-block mr24'" variant="withIcon" @click="handleAddKeyClick(KeyType.file)">
      {{ $t("components.keysManagement.actions.addFileKey") }}
    </ButtonVue>
    <KeysTable
      :title="$t('components.keysManagement.table.hardwareKeysTitle')"
      :keys="hardwareKeys"
      @onRemoveKeyClick="handleRemoveKeyClick"
      @onEditKeyClick="handleEditKeyClick"
    />
    <ButtonVue class="mr24" fa-icon-class="fa-solid fa-plus" variant="withIcon"  @click="handleAddKeyClick(KeyType.hardware)">
      {{ $t("components.keysManagement.actions.addHardwareKey") }}
    </ButtonVue>
    <ButtonVue v-show="hardwareKeys.length" fa-icon-class="fa-solid fa-edit" variant="withIcon"  @click="handleOpenIniEditor">
      {{ $t("components.keysManagement.actions.editOsplmIni") }}
    </ButtonVue>

    <Modal :show="addKeyModal" @close="handleCloseModal" :title="$t('components.keysManagement.modals.addKey.title')" @submit="handleSaveKey">
      <TextField
        v-if="deviceType === KeyType.file"
        :error="errors.fileKeyName"
        v-model="fileKeyName"
        :label="$t('components.keysManagement.modals.addKey.fields.techKeyName.label')"
        :description="$t('components.keysManagement.modals.addKey.fields.techKeyName.description')"
      />
      <TextField
        v-if="deviceType === KeyType.hardware"
        :error="errors.hardKeyName"
        v-model="hardKeyName"
        :label="$t('components.keysManagement.modals.addKey.fields.techKeyName.label')"
        :description="$t('components.keysManagement.modals.addKey.fields.techKeyName.description')"
      />
      <SelectVue
        hidden="true"
        :items="KEY_VARIANTS()"
        v-model="deviceType"
        :label="$t('components.keysManagement.modals.addKey.fields.mediaType')"
        :error="errors.deviceType"
      />
      <template v-if="deviceType === KeyType.file">
        <FileField
          accept=".dat"
          @change="handleFileChange"
          @reset="handleFileChange"
          :label="$t('components.keysManagement.modals.addKey.fields.fileKey.label')"
          :sub-label="$t('components.keysManagement.modals.addKey.fields.fileKey.description')"
          :error="errors.fileKeyFile"
          id="file-key-add-input"
        />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.keyIssuer')" v-model="fileKeyIssuer" :error="errors.fileKeyIssuer" />
        <TextField type="password" :label="$t('components.keysManagement.modals.addKey.fields.fileKeyPassword')" v-model="fileKeyPassword" :error="errors.fileKeyPassword" />
      </template>
      <template v-if="deviceType === KeyType.hardware">
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeyType')" v-model="hardKeyType" :error="errors.hardKeyType" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeyPassword')" type="password" v-model="hardKeyPassword" :error="errors.hardKeyPassword" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.keyIssuer')" v-model="hardKeyIssuer" :error="errors.hardKeyIssuer" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeyIssuerHost')" v-model="hardKeyIssuerHost" :error="errors.hardKeyIssuerHost" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeyIssuerPort')" v-model="hardKeyIssuerPort" :error="errors.hardKeyIssuerPort" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeySerialNumber')" v-model="hardKeySerialNumber" :error="errors.hardKeySerialNumber" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeyHost')" v-model="hardKeyHost" :error="errors.hardKeyHost" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeyPort')" v-model="hardKeyPort" :error="errors.hardKeyPort" />
        <TextField :label="$t('components.keysManagement.modals.addKey.fields.hardKeyMask')" v-model="hardKeyMask" :error="errors.hardKeyMask" />
      </template>
    </Modal>
    <Modal :red-button="true" :submit-btn-text="$t('actions.remove')" :show="!!removeKeyPrompt" :title="$t('components.keysManagement.modals.removeKey.title')" @close="handleCloseRemoveKeyPrompt" @submit="removeKey(removeKeyPrompt)">
      <Typography variant="bodyText">{{ $t("components.keysManagement.modals.removeKey.text", {key: removeKeyPrompt}) }}</Typography>
    </Modal>
    <Modal :submit-btn-text="$t('actions.save')" :show="!!editIniModal" :title="$t('components.keysManagement.modals.editOsplmIni.title')" @close="editIniModal = false" @submit="handleSubmitIniModal">
      <TextField :label="$t('components.keysManagement.modals.editOsplmIni.fields.hardKeyOsplmIni')" multiline rows="20" v-model="tempIni" :error="tempIni.length ? undefined : 'required'" />
    </Modal>
    <Modal
      :has-cancel-btn="false"
      :submit-btn-text="$t('actions.gotIt')"
      :show="!!cantRemoveKeyDialog"
      :title="$t('components.keysManagement.modals.cantRemoveKey.title')"
      @close="cantRemoveKeyDialog = ''"
      @submit="cantRemoveKeyDialog = ''; handleCloseRemoveKeyPrompt()"
    >
      <Typography variant="bodyText">
       {{ $t('components.keysManagement.modals.cantRemoveKey.text.keyUsage', { key: cantRemoveKeyDialog }) }}
        <ul>
          <li v-for="registryName in getKeyUsage(cantRemoveKeyDialog)" v-bind:key="registryName">{{ registryName }}</li>
        </ul>
        {{ $t('components.keysManagement.modals.cantRemoveKey.text.firstChangeKey') }}
      </Typography>
    </Modal>
    <Modal :show="!!currentKey" @close="currentKey = null" :title="$t('components.keysManagement.modals.editKey.title')" @submit="handleSaveAllowedRegistries">
      <Banner v-show="getKeyUsage(getKeyName(currentKey!))?.length" class="mb24" :description="$t('components.keysManagement.modals.editKey.banner', {registries: getKeyUsage(getKeyName(currentKey!)).join(', ')})" />
      <SelectVue
        onchange="handleAllowedRegistryChange"
        multiple
        v-model="currentKey!.allowedRegistries"
        :items="props.registries"
        :label="$t('components.keysManagement.modals.editKey.fields.allowedRegistries')"
        :item-props="checkDisabledItems"
      />
    </Modal>
  </div>
</template>

<style scoped lang="scss">
.rg-info-table {
  table-layout: fixed;
  tbody tr {
    border-bottom: 1px solid #dee2e6;
  }
}
.w512 {
  width: $wizard-width;
}
.inline-block {
  display: inline-block;
}
.mr24 {
  margin-right: 24px;
}
</style>