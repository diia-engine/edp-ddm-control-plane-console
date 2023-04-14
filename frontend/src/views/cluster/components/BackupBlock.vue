<script setup lang="ts">
import { getErrorMessage } from '@/utils';
import axios from 'axios';

</script>
<script lang="ts">
export default {
    props: ['backupSchedule'],
    data() {
        return {
            disabled: false,
            errorsMap: {},
            backupScheduleData: {
                NexusSchedule: '',
                NexusExpiresInDays: '',
                ControlPlaneSchedule: '',
                ControlPlaneExpiresInDays: '',
                UserManagementSchedule: '',
                UserManagementExpiresInDays: '',
                MonitoringSchedule: '',
                MonitoringExpiresInDays: '',
            },
        };
    },
     mounted: function () {
        const {
            NexusSchedule,
            NexusExpiresInDays,
            ControlPlaneSchedule,
            ControlPlaneExpiresInDays,
            UserManagementSchedule,
            UserManagementExpiresInDays,
            MonitoringSchedule,
            MonitoringExpiresInDays,
        } = this.backupSchedule;

        this.backupScheduleData = {
            NexusSchedule,
            ControlPlaneSchedule,
            UserManagementSchedule,
            MonitoringSchedule,
            NexusExpiresInDays: NexusExpiresInDays === '0' ? '' : NexusExpiresInDays,
            ControlPlaneExpiresInDays: ControlPlaneExpiresInDays === '0' ? '' : ControlPlaneExpiresInDays,
            UserManagementExpiresInDays: UserManagementExpiresInDays === '0' ? '' : UserManagementExpiresInDays,
            MonitoringExpiresInDays: MonitoringExpiresInDays === '0' ? '' : MonitoringExpiresInDays,
        };
    },
    methods: {
        submit() {
            this.disabled = true;
            let formData = new FormData();
            formData.append("nexus-schedule", this.backupScheduleData.NexusSchedule);
            formData.append("nexus-expires-in-days", this.backupScheduleData.NexusExpiresInDays);
            formData.append("control-plane-schedule", this.backupScheduleData.ControlPlaneSchedule);
            formData.append("control-plane-expires-in-days", this.backupScheduleData.ControlPlaneExpiresInDays);
            formData.append("control-plane-expires-in-days", this.backupScheduleData.ControlPlaneExpiresInDays);
            formData.append("user-management-schedule", this.backupScheduleData.UserManagementSchedule);
            formData.append("user-management-expires-in-days", this.backupScheduleData.UserManagementExpiresInDays);
            formData.append("monitoring-schedule", this.backupScheduleData.MonitoringSchedule);
            formData.append("monitoring-expires-in-days", this.backupScheduleData.MonitoringExpiresInDays);

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
                v-model="backupScheduleData.NexusSchedule" />
            <p>Використовується Cron-формат.</p>
            <span v-for="$val in (errorsMap as any)?.NexusSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.NexusExpiresInDays }">
            <label for="nexus-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="nexus-expires-in-days" name="nexus-expires-in-days" placeholder="5"
                v-model="backupScheduleData.NexusExpiresInDays" />
            <p>Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час
                збереження більшим за період між створенням копій.</p>
            <span v-for="$val in (errorsMap as any)?.NexusExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <h3>Control Plane</h3>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.ControlPlaneSchedule }">
            <label for="control-plane-schedule">Розклад</label>
            <input type="text" name="control-plane-schedule" id="control-plane-schedule" placeholder="0 10 * * *"
                v-model="backupScheduleData.ControlPlaneSchedule" />
            <p>Використовується Cron-формат.</p>
            <span v-for="$val in (errorsMap as any)?.ControlPlaneSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.ControlPlaneExpiresInDays }">
            <label for="control-plane-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="control-plane-expires-in-days" name="control-plane-expires-in-days" placeholder="5"
                v-model="backupScheduleData.ControlPlaneExpiresInDays" />
            <p>Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час
                збереження більшим за період між створенням копій.</p>
            <span v-for="$val in (errorsMap as any)?.ControlPlaneExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <h3>User Management</h3>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.UserManagementSchedule }">
            <label for="user-management-schedule">Розклад</label>
            <input type="text" name="user-management-schedule" id="user-management-schedule" placeholder="0 10 * * *"
                v-model="backupScheduleData.UserManagementSchedule" />
            <p>Використовується Cron-формат.</p>
            <span v-for="$val in (errorsMap as any)?.UserManagementSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>

        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.UserManagementExpiresInDays }">
            <label for="user-management-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="user-management-expires-in-days" name="user-management-expires-in-days" placeholder="5"
                v-model="backupScheduleData.UserManagementExpiresInDays" />
            <p>Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час
                збереження більшим за період між створенням копій.</p>
            <span v-for="$val in (errorsMap as any)?.UserManagementExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <h3>Monitoring</h3>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.MonitoringSchedule }">
            <label for="monitoring-schedule">Розклад</label>
            <input type="text" name="monitoring-schedule" id="monitoring-schedule" placeholder="0 10 * * *"
                v-model="backupScheduleData.MonitoringSchedule" />
            <p>Використовується Cron-формат.</p>
            <span v-for="$val in (errorsMap as any)?.MonitoringSchedule" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>
        <div class="rc-form-group" :class="{ error: (errorsMap as any)?.MonitoringExpiresInDays }">
            <label for="monitoring-expires-in-days">Час зберігання в днях</label>
            <input type="text" id="monitoring-expires-in-days" name="monitoring-expires-in-days" placeholder="5"
                v-model="backupScheduleData.MonitoringExpiresInDays" />
            <p>Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час
                збереження більшим за період між створенням копій.</p>
            <span v-for="$val in (errorsMap as any)?.MonitoringExpiresInDays" :key="$val">
                {{ getErrorMessage($val) }}
            </span>
        </div>

        <div class="rc-form-group">
            <button type="submit" name="submit" :disabled="disabled">Підтвердити</button>
        </div>
    </form>
</template>
