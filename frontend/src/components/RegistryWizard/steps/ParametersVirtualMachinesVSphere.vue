<script setup lang="ts">
import { toRefs, computed } from 'vue';
import * as Yup from 'yup';
import { useField, useForm } from 'vee-validate';
import TextField from '@/components/common/TextField.vue';
import Banner from '@/components/common/Banner.vue';
import Typography from '@/components/common/Typography.vue';
import type { ComputeResources } from '@/types/registry';
import i18n from '@/localization';

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
  const bannerDescription = i18n.global.t('components.parametersVirtualMachinesVSphere.text.openShiftClusterDeployed');
  if (isEditAction.value) {
    if (isPlatformAdmin.value) {
      return `${bannerDescription} \n\n ${i18n.global.t('components.parametersVirtualMachinesVSphere.text.immediatelyApplyChanges')}`;
    }
    return `${bannerDescription} \n\n ${i18n.global.t('components.parametersVirtualMachinesVSphere.text.contactPlatformAdministrator')}`;
  }
  return bannerDescription;
}

function prepareDescriptionInstanceVolumeSize(): string {
  if (isEditAction.value) {
    return i18n.global.t('components.parametersVirtualMachinesVSphere.text.rangeOfValidSizeValuesVolumeSize', { instanceVolumeSize: computeResources.value.instanceVolumeSize });
  }
  return i18n.global.t('components.parametersVirtualMachinesVSphere.text.rangeOfValidSizeValues');
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
  <Typography variant="h3" class="h3">{{ $t('components.parametersVirtualMachinesVSphere.title') }}</Typography>
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
      :label="$t('components.parametersVirtualMachinesVSphere.fields.instanceCount.label')"
      name="instanceCount"
      v-model="instanceCount"
      :error="errors.instanceCount"
      :description="$t('components.parametersVirtualMachinesVSphere.fields.instanceCount.description')"
    >
    </TextField>
  </div>
  <div class="rc-form-group">
    <TextField
      required
      :disabled="disabled"
      :label="$t('components.parametersVirtualMachinesVSphere.fields.instanceVolumeSize.label')"
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
      :label="$t('components.parametersVirtualMachinesVSphere.fields.vSphereInstanceCPUCount.label')"
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
      :label="$t('components.parametersVirtualMachinesVSphere.fields.vSphereInstanceCoresPerCPUCount.label')"
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
      :label="$t('components.parametersVirtualMachinesVSphere.fields.vSphereInstanceRAMSize.label')"
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
