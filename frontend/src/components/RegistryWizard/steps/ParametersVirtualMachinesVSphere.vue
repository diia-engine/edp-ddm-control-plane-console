<script setup lang="ts">
import { toRefs, computed } from 'vue';
import * as Yup from 'yup';
import { useField, useForm } from 'vee-validate';
import TextField from '@/components/common/TextField.vue';
import Banner from '@/components/common/Banner.vue';
import Typography from '@/components/common/Typography.vue';
import type { ComputeResources } from '@/types/registry';

type ParametersVirtualMachinesVSphereProps = {
  computeResources: ComputeResources
  isPlatformAdmin: boolean;
  isEditAction: boolean
};

interface FormValues {
  instanceCount: number;
  instanceVolumeSize: number;
  vSphereInstanceCPUCount: number;
  vSphereInstanceCoresPerCPUCount: number;
  vSphereInstanceRAMSize: number;
}

const props = defineProps<ParametersVirtualMachinesVSphereProps>();
const { computeResources, isPlatformAdmin, isEditAction } = toRefs(props);
const disabled = !isPlatformAdmin.value && isEditAction.value;

const validationSchema = Yup.object<FormValues>({
  instanceCount: Yup.number().required().max(2000).min(1),
  instanceVolumeSize: Yup.number().required().min(computeResources.value?.instanceVolumeSize || 1).max(200),
  vSphereInstanceCPUCount: Yup.number().required().min(1),
  vSphereInstanceCoresPerCPUCount: Yup.number().required().min(1),
  vSphereInstanceRAMSize: Yup.number().required().min(1),
});

const { errors, validate } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    instanceCount: computeResources.value?.instanceCount || 2,
    instanceVolumeSize: computeResources.value?.instanceVolumeSize || 80,
    vSphereInstanceCPUCount: computeResources.value?.vSphereInstanceCPUCount || 8,
    vSphereInstanceCoresPerCPUCount: computeResources.value?.vSphereInstanceCoresPerCPUCount || 1,
    vSphereInstanceRAMSize: computeResources.value?.vSphereInstanceRAMSize || 32768,
  },
});

const { value: instanceCount } = useField<string>('instanceCount');
const { value: instanceVolumeSize } = useField<string>('instanceVolumeSize');
const { value: vSphereInstanceCPUCount } = useField<string>(
  'vSphereInstanceCPUCount'
);
const { value: vSphereInstanceCoresPerCPUCount } = useField<string>(
  'vSphereInstanceCoresPerCPUCount'
);
const { value: vSphereInstanceRAMSize } = useField<string>(
  'vSphereInstanceRAMSize'
);

function validator() {
  return new Promise((resolve) => {
    validate().then(async (res) => {
      if (res.valid) {
        resolve(true);
      }
    });
  });
}

function prepareBannerDescription(): string {
  const bannerDescription = `Кластер OpenShift розгорнутий на інфраструктурі VSphere. Докладніше про допустимі значення параметрів віртуальних машин – в системних вимогах OpenShift.`;
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
    instanceCount: instanceCount.value,
    instanceVolumeSize: instanceVolumeSize.value,
    vSphereInstanceCPUCount: vSphereInstanceCPUCount.value,
    vSphereInstanceCoresPerCPUCount: vSphereInstanceCoresPerCPUCount.value,
    vSphereInstanceRAMSize: vSphereInstanceRAMSize.value,
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
      label="Розмір системного диску віртуальної машини (GB)"
      name="instanceVolumeSize"
      v-model="instanceVolumeSize"
      :error="errors.instanceVolumeSize"
      :description="prepareDescriptionInstanceVolumeSize()"
    >
    </TextField>
  </div>
  <div class="rc-form-group">
    <TextField
      required
      :disabled="disabled"
      label="Кількість vCPU віртуальної машини"
      name="vSphereInstanceCPUCount"
      v-model="vSphereInstanceCPUCount"
      :error="errors.vSphereInstanceCPUCount"
    >
    </TextField>
  </div>
  <div class="rc-form-group">
    <TextField
      required
      :disabled="disabled"
      label="Кількість ядер у кожного vCPU віртуальної машини"
      name="vSphereInstanceCoresPerCPUCount"
      v-model="vSphereInstanceCoresPerCPUCount"
      :error="errors.vSphereInstanceCoresPerCPUCount"
    >
    </TextField>
  </div>
  <div class="rc-form-group">
    <TextField
      required
      :disabled="disabled"
      label="Кількість RAM віртуальної машини (Мі)"
      name="vSphereInstanceRAMSize"
      v-model="vSphereInstanceRAMSize"
      :error="errors.vSphereInstanceRAMSize"
    >
    </TextField>
  </div>
</template>

<style scoped>
.h3 {
  margin-bottom: 16px;
}
.mb24 {
  margin-bottom: 24px;
}
</style>
