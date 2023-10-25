<script setup lang="ts">
import * as yup from 'yup';
import { useForm } from 'vee-validate';
import { onUpdated, toRefs } from 'vue';
import axios, { AxiosError } from 'axios';

import RateLimitField from '@/views/registry/components/RateLimitField.vue';
import type { PublicApiLimits } from '@/types/registry';
import Modal from '@/components/common/Modal.vue';
import TextField from '@/components/common/TextField.vue';
import Typography from '@/components/common/Typography.vue';
import RadioGroup from '@/components/common/RadioGroup.vue';
import { getErrorMessage } from '@/utils';
import i18n from "@/localization";

enum UrlType {
  SEARCH_CONDITION = 'searchCondition',
  FILE = 'file',
}

interface Data {
  name: string;
  urlType?: UrlType;
  url: string;
  limits: PublicApiLimits;
}

interface RegistryEditPublicApiModalProps {
  publicApiPopupShow: boolean;
  publicApiValues: Data | null;
  publicApiList: Data[];
  registry: string;
}

const searchConditionRadioItem = [
  { label: i18n.global.t('components.registryEditPublicApiModal.fields.url.searchConditionOption'), value: UrlType.SEARCH_CONDITION },
];
const fileRadioItem = [
  { label: i18n.global.t('components.registryEditPublicApiModal.fields.url.fileOption'), value: UrlType.FILE },
];

const scUrlRegex = /^\/[A-Za-z0-9-]*$/i;
const fileUrlRegex = /\/files\/[A-Za-z0-9-]+\/\(\.\*\)\/[A-Za-z0-9-]+\/\(\.\*\)/i;
const allowedFileUrlChars = '[A-Za-z0-9-]*';
const props = defineProps<RegistryEditPublicApiModalProps>();
const { publicApiPopupShow, publicApiValues, publicApiList, registry } = toRefs(props);

const numberSchema = yup.number()
.transform((value) => value === null ? NaN : value)
.typeError('required')
.positive()
.max(Number.MAX_SAFE_INTEGER, 'moreThanMaxValue')
.integer();
const validationSchema = yup.object({
  name: yup.string().required().min(3).max(32).matches(/^[a-z0-9]([a-z0-9-]){1,30}[a-z0-9]$/).test({
    message: 'isUnique',
    test: function (value) {
      if (publicApiValues.value?.name) {
        return true;
      }
      return publicApiList.value?.findIndex(({ name }) => name === value) === -1;
    },
  }),
  url: yup.string().required()
  .test({
    message: 'matches',
    test: function (value: string) {
      if (this.parent.urlType !== UrlType.SEARCH_CONDITION) {
        return true;
      }
      return !!value.match(scUrlRegex);
    },
  })
  .test({
    message: 'required',
    test: function (value: string) {
      if (this.parent.urlType !== UrlType.FILE) {
        return true;
      }
      return !!value.match(fileUrlRegex);
    },
  })
  .test({
    message: 'isUnique',
    test: function (value) {
      if (publicApiValues.value?.url === value) {
        return true;
      }
      return publicApiList.value?.findIndex(({ url }) => url === value) === -1;
    },
  }),
  urlTable: yup.string().ensure().test({
    message: 'required',
    test: function (value: string) {
      if (this.parent.urlType !== UrlType.FILE) {
        return true;
      }

      return !!value;
    },
  }),
  urlColumn: yup.string().ensure().test({
    message: 'required',
    test: function (value: string) {
      if (this.parent.urlType !== UrlType.FILE) {
        return true;
      }

      return !!value;
    },
  }),
  limits: yup.object({
    second: numberSchema,
    minute: numberSchema,
    hour: numberSchema,
    day: numberSchema,
    month: numberSchema,
    year: numberSchema,
  }).required().test({
    message: 'rateLimitError',
    test: function (value) {
      return !!Object.keys(value).find((key) => {
        return value[key as keyof typeof value] !== undefined;
      });
    },
  }),
});

const { handleSubmit, values, errors, setValues, setErrors, setFieldValue } = useForm({
  validationSchema,
});

const emit = defineEmits(['hideModalWindow']);

function hideModalWindow() {
  emit('hideModalWindow');
}

const getTablePath = (values: Data) => {
  return values.url.split('/')[2] || '';
};

const getColumnPath = (values: Data) => {
  return values.url.split('/')[4] || '';
};

onUpdated(()=> {
  if (publicApiValues?.value) {
    const url = publicApiValues.value.url || '';
    const isSearchCondition = !!url.match(scUrlRegex);
    setValues({
      ...publicApiValues.value,
      urlType: isSearchCondition ? UrlType.SEARCH_CONDITION : UrlType.FILE,
      urlTable: isSearchCondition ? '' : getTablePath(publicApiValues.value),
      urlColumn: isSearchCondition ? '' : getColumnPath(publicApiValues.value),
    });
  } else {
    setValues({ name: '', url: '', limits: {}, urlType: UrlType.SEARCH_CONDITION, urlTable: '', urlColumn: '' });
  }
  setErrors({});
});

const submit = handleSubmit(() => {
  const formData = new FormData();
  const limits = values.limits;
  const limitsFormatted = Object.keys(limits).reduce((result, key) => {
    if (isNaN(limits[key])) {
      return result;
    }
    return {
      ...result,
      [key]: Number(limits[key]),
    };
}, {});

  formData.append("reg-name", values.name);
  formData.append("reg-url", values.url);
  formData.append("reg-limits", JSON.stringify(limitsFormatted));

  if (publicApiValues.value?.name) {
    axios.post(`/admin/registry/public-api-edit/${registry.value}`, formData, {
      headers: {
          'Content-Type': 'multipart/form-data'
      }
    }).then(() => {
      window.location.assign(`/admin/registry/view/${registry.value}`);
    }).catch(({ response }: AxiosError<any>) => {
      setErrors(response?.data.errors);
    });
  } else {
    axios.post(`/admin/registry/public-api-add/${registry.value}`, formData, {
      headers: {
          'Content-Type': 'multipart/form-data'
      }
    }).then(() => {
      window.localStorage.setItem("mr-scroll", "true");
      window.location.assign(`/admin/registry/view/${registry.value}`);
    }).catch(({ response }: AxiosError<any>) => {
      setErrors(response?.data.errors);
    });
  }
});

const setTablePath = (tablePath: string) => {
  const newUrl = `/files/${tablePath}/(.*)/${values.urlColumn || ''}/(.*)`;
  setFieldValue('url', newUrl);
  setFieldValue('urlTable', tablePath);
};

const setColumnPath = (columnPath: string) => {
  const newUrl = `/files/${values.urlTable || ''}/(.*)/${columnPath}/(.*)`;
  setFieldValue('url', newUrl);
  setFieldValue('urlColumn', columnPath);
};

const setUrlType = (urlType: UrlType) => {
  setFieldValue('urlType', urlType);
  setFieldValue('url', '');
  setFieldValue('urlColumn', '');
  setFieldValue('urlTable', '');
};
</script>

<template>
  <Modal 
    :show="publicApiPopupShow"
    :title="publicApiValues?.name ? $t('components.registryEditPublicApiModal.titleEdit', { name: publicApiValues?.name }) : $t('components.registryEditPublicApiModal.titleSet')"
    :submitBtnText="publicApiValues?.name ? $t('actions.confirm') : $t('actions.give')"
    @close="hideModalWindow" @submit="submit"
  >
    <form id="backupPlace-form">
      <Typography variant="bodyText" class="content-text" v-if="!publicApiValues?.name">
        {{ $t('components.registryEditPublicApiModal.text.providePublicAccess') }}
      </Typography>
      <TextField 
        v-if="!publicApiValues?.name"
        :label="$t('components.registryEditPublicApiModal.fields.name.label')"
        name="name"
        :description="$t('components.registryEditPublicApiModal.fields.name.description')"
        v-model="values.name"
        :error="errors?.name"
        required
      />
      <label for="url" class="custom-label">
        {{ $t('components.registryEditPublicApiModal.fields.url.label') }}
      </label>
      <RadioGroup
        name="urlTypeSC"
        class="url-radio"
        :items="searchConditionRadioItem"
        :model-value="values.urlType"
        @update:model-value="setUrlType"
      />
      <TextField
        v-if="values.urlType === UrlType.SEARCH_CONDITION"
        rootClass="url-radio-content url-input"
        name="url"
        :description="$t('components.registryEditPublicApiModal.fields.url.description')"
        v-model="values.url"
        :error="errors?.url"
        required
      />
      <Typography variant="small" v-if="values.urlType === UrlType.SEARCH_CONDITION" class="url-radio-content">
        {{ $t('components.registryEditPublicApiModal.text.forExample') }} <b>/search-laboratories-by-city</b>
      </Typography>

      <RadioGroup
        name="urlTypeFile"
        class="url-radio"
        :items="fileRadioItem"
        :model-value="values.urlType"
        @update:model-value="setUrlType"
      />
      <div v-if="values.urlType === UrlType.FILE" class="file-input url-radio-content">
        <div class="file-text">/files/</div>
        <TextField
          name="urlTable"
          rootClass="file-part-input"
          placeholder="table-name"
          :model-value="values.urlTable"
          @update:model-value="(newValue: string) => setTablePath(newValue)"
          :allowed-characters="allowedFileUrlChars"
          :error="errors?.urlTable"
          required
          hideErrorMessage
        />
        <div class="file-text">/(.*)/</div>
        <TextField
          name="urlColumn"
          rootClass="file-part-input"
          placeholder="column-name"
          :model-value="values.urlColumn"
          @update:model-value="(newValue: string) => setColumnPath(newValue)"
          :allowed-characters="allowedFileUrlChars"
          :error="errors?.urlColumn"
          required
          hideErrorMessage
        />
        <div class="file-text">/(.*)</div>
      </div>
      <Typography
        v-if="values.urlType === UrlType.FILE  && errors?.url"
        class="file-url-error url-radio-content"
        variant="small"
      >
        {{ getErrorMessage(errors?.url || '') }}
      </Typography>
      <Typography variant="small" v-if="values.urlType === UrlType.FILE" class="url-radio-content">
        {{ $t('components.registryEditPublicApiModal.text.linkFormat') }} <b>/files/table-name/(.*)/column-name/(.*)</b><br/>
        {{ $t('components.registryEditPublicApiModal.text.linkUnique') }}
      </Typography>

      <RateLimitField
        :label="$t('components.registryEditPublicApiModal.fields.rateLimit.label')"
        name="limits"
        class="rate-limit-field"
        v-model="values.limits"
        :errors="{
          second: errors?.['limits.second'],
          minute: errors?.['limits.minute'],
          hour: errors?.['limits.hour'],
          day: errors?.['limits.day'],
          month: errors?.['limits.month'],
          year: errors?.['limits.year'],
          common: errors?.['limits'],
        }"
        required
      />
    </form>
  </Modal>
</template>

<style lang="scss" scoped>
.content-text {
  margin-bottom: 24px;
}
.field-header {
  font-weight: 700;
  margin-bottom: 8px;
  margin-top: 8px;
}

.custom-label {
  font-size: 16px;
  font-weight: bold;
  margin: 0 0 8px 0;
}

.rate-limit-field {
  margin-top: 32px!important;
}

.url-input {
  margin-bottom: 0;
}
.url-radio {
  margin-top: 16px;
  margin-bottom: 8px;
}

.url-radio-content {
  margin-left: 32px;
}
.file-input {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.file-text {
  margin-right: 4px;
  margin-left: 4px;
}

.file-part-input {
  width: 163px;
  margin: 0;
}

.file-url-error {
  margin-bottom: 8px;
  color: $error-color;
}
</style>

