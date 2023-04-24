<script setup lang="ts">
import axios, { AxiosError } from 'axios';
import TextField from '@/components/common/TextField.vue';
import * as yup from 'yup';
import { useForm } from 'vee-validate';
import { parseCronExpression } from 'cron-schedule';
import { ref, toRefs } from 'vue';

interface Data {
  NexusSchedule: string;
  NexusExpiresInDays: string;
  ControlPlaneSchedule: string;
  ControlPlaneExpiresInDays: string;
  UserManagementSchedule: string;
  UserManagementExpiresInDays: string;
  MonitoringSchedule: string;
  MonitoringExpiresInDays: string;
}

interface BackupBlockProps {
  backupSchedule: Data;
}

const props = defineProps<BackupBlockProps>();
const { backupSchedule } = toRefs(props);
const beginValidation = ref(false);

const parseCronExpressionRules = () => {
  return yup.string()
    .required()
    .test({
      message: 'parseCronExpression',
      test: function (value) {
        try {
          parseCronExpression(value || "");
          return true;
        }
        catch (e: any) {
          return false;
        }
      },
    });
};

const validationSchema = yup.object({
  NexusSchedule: parseCronExpressionRules(),
  NexusExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
  ControlPlaneSchedule: parseCronExpressionRules(),
  ControlPlaneExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
  UserManagementSchedule: parseCronExpressionRules(),
  UserManagementExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
  MonitoringSchedule: parseCronExpressionRules(),
  MonitoringExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
});

const { handleSubmit, useFieldModel, setErrors, validate, values, errors } = useForm({
  validationSchema, initialValues: backupSchedule, validateOnMount: false,
});

const [
  NexusSchedule,
  NexusExpiresInDays,
  ControlPlaneSchedule,
  ControlPlaneExpiresInDays,
  UserManagementSchedule,
  UserManagementExpiresInDays,
  MonitoringSchedule,
  MonitoringExpiresInDays,
] = useFieldModel([
  'NexusSchedule',
  'NexusExpiresInDays',
  'ControlPlaneSchedule',
  'ControlPlaneExpiresInDays',
  'UserManagementSchedule',
  'UserManagementExpiresInDays',
  'MonitoringSchedule',
  'MonitoringExpiresInDays',
]);

NexusExpiresInDays.value = NexusExpiresInDays.value === '0' ? '' : NexusExpiresInDays.value;
ControlPlaneExpiresInDays.value = ControlPlaneExpiresInDays.value === '0' ? '' : ControlPlaneExpiresInDays.value;
UserManagementExpiresInDays.value = UserManagementExpiresInDays.value === '0' ? '' : UserManagementExpiresInDays.value;
MonitoringExpiresInDays.value = MonitoringExpiresInDays.value === '0' ? '' : MonitoringExpiresInDays.value;

const submit = handleSubmit(() => {
  let formData = new FormData();

  formData.append("nexus-schedule", values.NexusSchedule);
  formData.append("nexus-expires-in-days", values.NexusExpiresInDays);
  formData.append("control-plane-schedule", values.ControlPlaneSchedule);
  formData.append("control-plane-expires-in-days", values.ControlPlaneExpiresInDays);
  formData.append("control-plane-expires-in-days", values.ControlPlaneExpiresInDays);
  formData.append("user-management-schedule", values.UserManagementSchedule);
  formData.append("user-management-expires-in-days", values.UserManagementExpiresInDays);
  formData.append("monitoring-schedule", values.MonitoringSchedule);
  formData.append("monitoring-expires-in-days", values.MonitoringExpiresInDays);

  
  axios.post('/admin/cluster/backup-schedule', formData, {
      headers: {
          'Content-Type': 'multipart/form-data'
      }
  }).then(() => {
    beginValidation.value = false;
    window.location.assign('/admin/cluster/management');
  }).catch(({ response }: AxiosError<any>) => {
    setErrors(response?.data.errors);
  });
});

function onSubmit() {
  beginValidation.value = true;
  submit();
}

</script>

<template>
  <h2>Розклад резервного копіювання</h2>
  <form @submit.prevent="onSubmit" id="backup-schedule-form" class="registry-create-form wizard-form">
    <h3>Nexus</h3>
    <div class="form-group">
      <TextField
        label="Розклад"
        name="nexus-schedule"
        placeholder="0 10 * * *"
        description="Використовується Cron-формат."
        :value="NexusSchedule"
        :error="beginValidation ? errors.NexusSchedule : ''"
        @update="val => NexusSchedule = val"
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        label="Час зберігання (днів)"
        name="nexus-expires-in-days"
        placeholder="5"
        description="Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час збереження більшим за період між створенням копій."
        :value="NexusExpiresInDays"
        :error="beginValidation ? errors.NexusExpiresInDays : ''"
        @update="val => NexusExpiresInDays = val"
        @change="validate"
      />
    </div>

    <h3>Control Plane</h3>
    <div class="form-group">
      <TextField
        label="Розклад"
        name="control-plane-schedule"
        placeholder="0 10 * * *"
        description="Використовується Cron-формат."
        :value="ControlPlaneSchedule"
        :error="beginValidation ? errors.ControlPlaneSchedule : ''"
        @update="val => ControlPlaneSchedule = val"
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        label="Час зберігання (днів)"
        name="control-plane-expires-in-days"
        placeholder="5"
        description="Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час збереження більшим за період між створенням копій."
        :value="ControlPlaneExpiresInDays"
        :error="beginValidation ? errors.ControlPlaneExpiresInDays : ''"
        @update="val => ControlPlaneExpiresInDays = val"
        @change="validate"
      />
    </div>

    <h3>User Management</h3>
    <div class="form-group">
      <TextField
        label="Розклад"
        name="user-management-schedule"
        placeholder="0 10 * * *"
        description="Використовується Cron-формат."
        :value="UserManagementSchedule"
        :error="beginValidation ? errors.UserManagementSchedule : ''"
        @update="val => UserManagementSchedule = val"
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        label="Час зберігання (днів)"
        name="user-management-expires-in-days"
        placeholder="5"
        description="Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час збереження більшим за період між створенням копій."
        :value="UserManagementExpiresInDays"
        :error="beginValidation ? errors.UserManagementExpiresInDays : ''"
        @update="val => UserManagementExpiresInDays = val"
        @change="validate"
      />
    </div>

    <h3>Monitoring</h3>
    <div class="form-group">
      <TextField
        label="Розклад"
        name="monitoring-schedule"
        placeholder="0 10 * * *"
        description="Використовується Cron-формат."
        :value="MonitoringSchedule"
        :error="beginValidation ? errors.MonitoringSchedule : ''"
        @update="val => MonitoringSchedule = val"
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        label="Час зберігання (днів)"
        name="monitoring-expires-in-days"
        placeholder="5"
        description="Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час збереження більшим за період між створенням копій."
        :value="MonitoringExpiresInDays"
        :error="beginValidation ? errors.MonitoringExpiresInDays : ''"
        @update="val => MonitoringExpiresInDays = val"
        @change="validate"
      />
    </div>

    <div class="rc-form-group">
      <button type="submit" name="submit">Підтвердити</button>
    </div>
  </form>
</template>


<style lang="scss" scoped>

.form-group {
  margin-bottom: 24px;
}

</style>

