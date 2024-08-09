<script setup lang="ts">
import { toRefs, ref, computed } from 'vue';
import * as Yup from 'yup';
import { useField, useForm } from 'vee-validate';
import TextField from '@/components/common/TextField.vue';
import Banner from '@/components/common/Banner.vue';
import Typography from '@/components/common/Typography.vue';
import RadioGroup from '@/components/common/RadioGroup.vue';
import ToggleSwitch from '@/components/common/ToggleSwitch.vue';
import type { ComputeResources } from '@/types/registry';

type ParametersVirtualMachinesAWSProps = {
  computeResources: ComputeResources
  isPlatformAdmin: boolean;
  isEditAction: boolean
};

interface FormValues {
  instanceCount: number;
  awsInstanceType: string;
  awsInstanceVolumeType: string;
  instanceVolumeSize: number;
  awsSpotInstanceMaxPrice: number;
  maxPriceAWS: string;
}

enum AWSTypePrices {
  INSTANCE = 'instancePrice',
  OWN = 'ownPrice',
}

const radioButtons = [
  { label: 'On-Demand Instance price', value: AWSTypePrices.INSTANCE },
  { label: 'Вказати власну ($/година)', value: AWSTypePrices.OWN },
];

const props = defineProps<ParametersVirtualMachinesAWSProps>();
const { computeResources, isPlatformAdmin, isEditAction } = toRefs(props);

const awsSpotInstance = ref(computeResources.value?.awsSpotInstance);
const showMaxPriceAWS = ref(computeResources.value?.awsSpotInstance);
const disabled = !isPlatformAdmin.value && isEditAction.value;

const validationSchema = Yup.object<FormValues>({
  instanceCount: Yup.number().required().max(2000).min(1).integer(),
  awsInstanceType: Yup.string().required().min(1).matches(/^[a-zA-Z0-9.]*$/),
  awsInstanceVolumeType: Yup.string().required(),
  instanceVolumeSize: Yup.number().required().min(computeResources.value?.instanceVolumeSize || 1).max(200),
  maxPriceAWS: Yup.string(),
  awsSpotInstanceMaxPrice: Yup.number().when('maxPriceAWS', {
    is: (value: AWSTypePrices) =>
      value === AWSTypePrices.OWN && awsSpotInstance.value,
    then: (schema) => schema.required().min(1),
  }),
});

const { errors, validate } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    instanceCount: computeResources.value?.instanceCount || 2,
    awsInstanceType: computeResources.value?.awsInstanceType || 'r5.2xlarge',
    awsInstanceVolumeType: computeResources.value?.awsInstanceVolumeType || 'gp3',
    instanceVolumeSize: computeResources.value?.instanceVolumeSize || 80,
    ...(computeResources.value?.awsSpotInstanceMaxPrice && {awsSpotInstanceMaxPrice: computeResources.value?.awsSpotInstanceMaxPrice}),
    maxPriceAWS: computeResources.value?.awsSpotInstanceMaxPrice ? AWSTypePrices.OWN : AWSTypePrices.INSTANCE,
  },
});

const { value: instanceCount } = useField<number>('instanceCount');
const { value: awsInstanceType } = useField<string>('awsInstanceType');
const { value: awsInstanceVolumeType } = useField<string>(
  'awsInstanceVolumeType'
);
const { value: instanceVolumeSize } = useField<number>('instanceVolumeSize');
const { value: awsSpotInstanceMaxPrice } = useField<number>(
  'awsSpotInstanceMaxPrice'
);
const { value: maxPriceAWS } = useField<string>('maxPriceAWS');

function validator() {
  return new Promise((resolve) => {
    validate().then(async (res) => {
      if (res.valid) {
        resolve(true);
      }
    });
  });
}

function handleAwsSpotInstance() {
  showMaxPriceAWS.value = awsSpotInstance.value;
}

function prepareBannerDescription(): string {
  const bannerDescription = `Кластер OpenShift розгорнутий на інфраструктурі AWS. Докладніше про допустимі значення параметрів віртуальних машин – в системних вимогах OpenShift.`;
  if (isEditAction.value) {
    if (isPlatformAdmin.value) {
      return `${bannerDescription} \n\n Якщо потрібно одразу застосувати зміни для параметрів віртуальних машин (тип і розмір диску), то перед зміною необхідно попередньо вимкнути реєстр.`;
    }
    return `${bannerDescription} \n\n В разі необхідності редагування параметрів, зверніться до адміністратора Платформи.`;
  }
  return bannerDescription;
}

function prepareDescriptionInstanceVolumeSize(): string {
  if (isEditAction.value) {
    return `Допустимі значення: 50 - 200 GB, але не менше від поточного (${computeResources.value?.instanceVolumeSize} GB).`;
  }
  return 'Допустимі значення: 50 - 200 GB.';
}

const preparedComputeResources = computed(() =>
  JSON.stringify({
    instanceCount: Math.floor(instanceCount.value),
    awsInstanceType: awsInstanceType.value,
    awsSpotInstance: Boolean(awsSpotInstance.value),
    ...(awsSpotInstance.value &&
      maxPriceAWS.value === AWSTypePrices.OWN && {
        awsSpotInstanceMaxPrice: awsSpotInstanceMaxPrice.value,
      }),
    awsInstanceVolumeType: awsInstanceVolumeType.value,
    instanceVolumeSize: instanceVolumeSize.value,
  })
);

defineExpose({
  validator,
});
</script>

<template>
  <Typography variant="h3" class="h3">Параметри віртуальних машин</Typography>
  <Banner
    classes="mb24"
    :description="prepareBannerDescription()"
  />
  <input
    type="hidden"
    name="compute-resources"
    :value="preparedComputeResources"
  />
  <div class="rc-form-group">
    <TextField
      required
      :disabled="disabled"
      label="Кількість віртуальних машин"
      name="instanceCount"
      v-model="instanceCount"
      :error="errors.instanceCount"
      description="допустимі значення: 1 - 2000."
    >
    </TextField>
  </div>
  <div class="rc-form-group">
    <TextField
      required
      :disabled="disabled"
      label="Тип AWS EC2-інстансу"
      name="awsInstanceType"
      v-model="awsInstanceType"
      :error="errors.awsInstanceType"
      description="Наприклад: r5.2xlarge, m5.xlarge, c5.4xlarge."
    >
    </TextField>
  </div>
  <ToggleSwitch
    name="awsSpotInstance"
    :disabled="disabled"
    label="Використати AWS EC2 Spot-інстанс"
    v-model="awsSpotInstance"
    @change="handleAwsSpotInstance"
  />
  <template v-if="showMaxPriceAWS">
    <Typography variant="bodyText" class="mt16">
      Максимальна ціна AWS EC2-інстансу (за годину):
    </Typography>
    <RadioGroup
      name="maxPriceAWS"
      :disabled="disabled"
      :items="radioButtons"
      classes="mt16"
      v-model="maxPriceAWS"
    />
    <TextField
      v-if="maxPriceAWS === AWSTypePrices.OWN"
      :disabled="disabled"
      rootClass="textField"
      name="awsSpotInstanceMaxPrice"
      v-model="awsSpotInstanceMaxPrice"
      :error="errors.awsSpotInstanceMaxPrice"
    >
    </TextField>
  </template>
  <div class="rc-form-group mt24">
    <TextField
      required
      :disabled="disabled"
      label="Тип системного диску AWS EC2-інстансу"
      name="awsInstanceVolumeType"
      v-model="awsInstanceVolumeType"
      :error="errors.awsInstanceVolumeType"
    >
    </TextField>
  </div>
  <div class="rc-form-group">
    <TextField
      required
      :disabled="disabled"
      label="Розмір системного диску віртуальної машини (GB)"
      name="instanceVolumeSize"
      v-model="instanceVolumeSize"
      :error="errors.instanceVolumeSize"
      :description="prepareDescriptionInstanceVolumeSize()"
    >
    </TextField>
  </div>
</template>

<style scoped>
.h3 {
  margin-bottom: 16px;
}
.mt16 {
  margin-top: 16px;
}
.mb24 {
  margin-bottom: 24px;
}
.mt24 {
  margin-top: 24px;
}
.textField {
  margin-left: 32px;
}
</style>
