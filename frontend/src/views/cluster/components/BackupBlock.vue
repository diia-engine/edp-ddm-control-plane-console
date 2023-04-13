<script setup lang="ts">
import { getErrorMessage } from '@/utils';
import axios from 'axios';
import { toRefs } from 'vue';

interface BackupProps {
    backupSchedule: any;
}

const props = defineProps<BackupProps>();
const { backupSchedule } = toRefs(props);

</script>
<script lang="ts">
export default {
    data() {
        return {
            disabled: false,
            errorsMap: {}
        };
    },
    methods: {
        submit() {
            this.disabled = true;
            let formData = new FormData();
            formData.append("nexus-schedule", this.backupSchedule.NexusSchedule);
            formData.append("nexus-expires-in-days", this.backupSchedule.NexusExpiresInDays);
            formData.append("control-plane-schedule", this.backupSchedule.ControlPlaneSchedule);
            formData.append("control-plane-expires-in-days", this.backupSchedule.ControlPlaneExpiresInDays);
            formData.append("control-plane-expires-in-days", this.backupSchedule.ControlPlaneExpiresInDays);
            formData.append("user-management-schedule", this.backupSchedule.UserManagementSchedule);
            formData.append("user-management-expires-in-days", this.backupSchedule.UserManagementExpiresInDays);
            formData.append("monitoring-schedule", this.backupSchedule.MonitoringSchedule);
            formData.append("monitoring-expires-in-days", this.backupSchedule.MonitoringExpiresInDays);

            axios.post('/admin/cluster/backup-schedule', formData, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                }
            }).then(() => {
                window.location.assign('/admin/cluster/management');
            }).catch(({ response }) => {
                this.disabled = false;
                this.errorsMap = response.data.errors;
            });
        },
    }
};
</script>

<template>
    <h2>Розклад резервного копіювання</h2>
    <form @submit.prevent="submit" id="backup-schedule-form" class="registry-create-form wizard-form">
        <h3>Nexus</h3>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.NexusSchedule }">
            <label for="nexus-schedule">Розклад</label>
            <input type="text" name="nexus-schedule" id="nexus-schedule" placeholder="0 10 * * *"
                v-model="backupSchedule.NexusSchedule" />
            <span v-for="$val in (errorsMap as any)?.NexusSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.NexusExpiresInDays }">
            <label for="nexus-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="nexus-expires-in-days" name="nexus-expires-in-days" placeholder="5"
                v-model="backupSchedule.NexusExpiresInDays" />
            <span v-for="$val in (errorsMap as any)?.NexusExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <h3>Control Plane</h3>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.ControlPlaneSchedule }">
            <label for="control-plane-schedule">Розклад</label>
            <input type="text" name="control-plane-schedule" id="control-plane-schedule" placeholder="0 10 * * *"
                v-model="backupSchedule.ControlPlaneSchedule" />
            <span v-for="$val in (errorsMap as any)?.ControlPlaneSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.ControlPlaneExpiresInDays }">
            <label for="control-plane-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="control-plane-expires-in-days" name="control-plane-expires-in-days" placeholder="5"
                v-model="backupSchedule.ControlPlaneExpiresInDays" />
            <span v-for="$val in (errorsMap as any)?.ControlPlaneExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <h3>User Management</h3>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.UserManagementSchedule }">
            <label for="user-management-schedule">Розклад</label>
            <input type="text" name="user-management-schedule" id="user-management-schedule" placeholder="0 10 * * *"
                v-model="backupSchedule.UserManagementSchedule" />
            <span v-for="$val in (errorsMap as any)?.UserManagementSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>

        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.UserManagementExpiresInDays }">
            <label for="user-management-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="user-management-expires-in-days" name="user-management-expires-in-days" placeholder="5"
                v-model="backupSchedule.UserManagementExpiresInDays" />
            <span v-for="$val in (errorsMap as any)?.UserManagementExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <h3>Monitoring</h3>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.MonitoringSchedule }">
            <label for="monitoring-schedule">Розклад</label>
            <input type="text" name="monitoring-schedule" id="monitoring-schedule" placeholder="0 10 * * *"
                v-model="backupSchedule.MonitoringSchedule" />
            <span v-for="$val in (errorsMap as any)?.MonitoringSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.MonitoringExpiresInDays }">
            <label for="monitoring-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="monitoring-expires-in-days" name="monitoring-expires-in-days" placeholder="5"
                v-model="backupSchedule.MonitoringExpiresInDays" />
            <span v-for="$val in (errorsMap as any)?.MonitoringExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <div class="rc-form-group">
            <button type="submit" name="submit" :disabled="disabled">Підтвердити</button>
        </div>
    </form>
</template>
