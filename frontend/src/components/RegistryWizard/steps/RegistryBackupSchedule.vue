<script lang="ts">
import { defineComponent, inject } from 'vue';
import { parseCronExpression } from 'cron-schedule';
import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
import RegistryBackupSavePlaceModal from '@/components/RegistryBackupSavePlaceModal.vue';
import RegistryBackupDeletePlaceModal from '@/components/RegistryBackupDeletePlaceModal.vue';

interface RegistryEditTemplateVariables {
  registryValues: any;
  model: any;
}

export default defineComponent({
  components: { RegistryBackupSavePlaceModal, RegistryBackupDeletePlaceModal, TextField, Typography },
  data() {
    return {
      validated: false,
      beginValidation: false,
      validator: this.backupScheduleValidation,
      errors: {
        cronSchedule: "",
        days: "",
        registryBackupCronExpression: "",
      },
      nextDates: [] as string[],
      registryBackupNextDates: [] as string[],
      backupPlacePopupShow: false,
      backupDeletePlacePopupShow: false,
      data: {
        enabled: false,
        cronSchedule: "",
        days: "",
        registryBackup: {
          obc: {
            cronExpression: "",
            backupBucket: "",
            endpoint: "",
            login: "",
            password: "",
          },
        },
      }
    };
  },
  mounted() {
    try {
      const templateVariables = inject('TEMPLATE_VARIABLES') as RegistryEditTemplateVariables;

      this.data.enabled = templateVariables.registryValues.global.registryBackup.enabled;
      this.data.cronSchedule = templateVariables.registryValues.global.registryBackup.schedule;
      this.data.days = templateVariables.registryValues.global.registryBackup.expiresInDays;
      this.data.registryBackup.obc.cronExpression = templateVariables.registryValues.global.registryBackup.obc.cronExpression;
      this.data.registryBackup.obc.backupBucket = templateVariables.registryValues.global.registryBackup.obc.backupBucket;
      this.data.registryBackup.obc.endpoint = templateVariables.registryValues.global.registryBackup.obc.endpoint;
      this.data.registryBackup.obc.login = templateVariables?.model.OBCLogin;
      this.data.registryBackup.obc.password = templateVariables?.model.OBCPassword;
    }
    catch (e: any) {
      console.log(e);
    }
    this.cronExpressionChange();
    this.backupCronExpressionChange();
  },
  methods: {
    showBackupPlaceModal() {
      this.backupPlacePopupShow = true;
    },
    hideBackupPlaceModal() {
      this.backupPlacePopupShow = false;
    },
    backupPlaceSubmit(data: { backupBucket: string; endpoint: string; login: string; password: string; }) {
      this.data.registryBackup.obc.backupBucket = data.backupBucket;
      this.data.registryBackup.obc.endpoint = data.endpoint;
      this.data.registryBackup.obc.login = data.login;
      this.data.registryBackup.obc.password = data.password;
      this.backupPlacePopupShow = false;
    },
    showBackupDeletePlaceModal() {
      this.backupDeletePlacePopupShow = true;
    },
    hideBackupDeletePlaceModal() {
      this.backupDeletePlacePopupShow = false;
    },
    backupDeletePlaceSubmit() {
      this.data.registryBackup.obc.backupBucket = '';
      this.data.registryBackup.obc.endpoint = '';
      this.data.registryBackup.obc.login = '';
      this.data.registryBackup.obc.password = '';
      
      this.backupDeletePlacePopupShow = false;
    },
    daysChange() {
      const days = parseInt(this.data.days);
      if (this.data.days === "") {
        this.errors.days = 'required';
      } else if (!/^[0-9]+$/.test(this.data.days) || isNaN(days) || days <= 0) {
        this.errors.days = 'invalidFormat';
      }
    },
    cronExpressionChange() {
      this.errors.cronSchedule = "";
      if (this.data.cronSchedule === "") {
        this.nextDates = [];
        this.errors.cronSchedule = 'required';
        return;
      }
      try {
        const cron = parseCronExpression(this.data.cronSchedule);
        this.nextDates = [];
        let dt = new Date();
        for (let i = 0; i < 3; i++) {
          const next = cron.getNextDate(dt);
          this.nextDates.push(`${next.toLocaleDateString("uk")} ${next.toLocaleTimeString("uk")}`);
          dt = next;
        }
      }
      catch (e: any) {
        this.nextDates = [];
        this.errors.cronSchedule = 'invalidFormat';
      }
    },
    backupCronExpressionChange() {
      this.errors.registryBackupCronExpression = "";
      if (this.data.registryBackup.obc.cronExpression === "") {
        this.registryBackupNextDates = [];
        this.errors.registryBackupCronExpression = 'required';
        return;
      }
      try {
        const cron = parseCronExpression(this.data.registryBackup.obc.cronExpression);
        this.registryBackupNextDates = [];
        let dt = new Date();
        for (let i = 0; i < 3; i++) {
          const next = cron.getNextDate(dt);
          this.registryBackupNextDates.push(`${next.toLocaleDateString("uk")} ${next.toLocaleTimeString("uk")}`);
          dt = next;
        }
      }
      catch (e: any) {
        this.registryBackupNextDates = [];
        this.errors.registryBackupCronExpression = 'invalidFormat';
      }
    },
    backupScheduleValidation() {
      return new Promise < void  > ((resolve) => {
        this.data.cronSchedule = this.data.cronSchedule.trim();
        this.data.registryBackup.obc.cronExpression = this.data.registryBackup.obc.cronExpression.trim();
        if (!this.data.enabled) {
          resolve();
          return;
        }
        this.beginValidation = true;
        this.validated = false;
        this.errors = {
          cronSchedule: "",
          days: "",
          registryBackupCronExpression: "",
        };
        if (this.data.cronSchedule !== "") {
          try {
            parseCronExpression(this.data.cronSchedule);
          }
          catch (e: any) {
            this.nextDates = [];
            this.errors.cronSchedule = 'invalidFormat';
          }
        } else {
          this.errors.cronSchedule = 'required';
        }
        if (this.data.registryBackup.obc.cronExpression !== "") {
          try {
            parseCronExpression(this.data.registryBackup.obc.cronExpression);
          }
          catch (e: any) {
            this.registryBackupNextDates = [];
            this.errors.registryBackupCronExpression = 'invalidFormat';
          }
        } else {
          this.errors.registryBackupCronExpression = 'required';
        }
        const days = parseInt(this.data.days);
        if (this.data.days === "") {
          this.errors.days = 'required';
        } else if (!/^[0-9]+$/.test(this.data.days) || isNaN(days) || days <= 0) {
          this.errors.days = 'invalidFormat';
        }
        if ( this.errors.days ||  this.errors.cronSchedule || this.data.cronSchedule === "" || this.data.days === "") {
          return;
        }
        if (this.errors.registryBackupCronExpression || this.data.registryBackup.obc.cronExpression === "") {
          return;
        }
        this.validated = true;
        this.beginValidation = false;
        resolve();
      });
    },
  },
});
</script>


<template>
  <h2>Резервне копіювання</h2>
  <p>Можливість вказати розклад створення резервних копій реєстру та термін їх зберігання.</p>
  <div class="toggle-switch backup-switch">
      <input v-model="data.enabled" class="switch-input"
              type="checkbox" id="backup-schedule-switch-input"  name="backup-schedule-enabled" />
      <label for="backup-schedule-switch-input">Toggle</label>
      <span>Налаштувати резервне копіювання</span>
  </div>

  <div v-show="data.enabled">
      <div class="form-group">
        <TextField
          label="Розклад"
          name="cron-schedule"
          placeholder="5 4 * * *"
          description="Використовується Cron-формат."
          :value="data.cronSchedule"
          :error="beginValidation ? errors.cronSchedule : ''"
          @update="val => data.cronSchedule = val"
          @change="cronExpressionChange"
        />
      </div>
      <div v-show="nextDates.length" class="form-group">
          <label>Наступні запуски резервного копіювання (за київським часом)</label>
          <ul class="cron-next-dates">
              <li v-for="date in nextDates" v-bind:key="date">
                <Typography variant="bodyText">{{ date }}</Typography>
              </li>
          </ul>
      </div>

      <div class="form-group">
        <TextField
          label="Час зберігання (днів)"
          name="cron-schedule-days"
          placeholder="3"
          description="Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час збереження більшим за період між створенням копій."
          :value="data.days"
          :error="beginValidation ? errors.days : ''"
          @update="val => data.days = val"
          @change="daysChange"
        />
      </div>

      <h2>Резервне копіювання реплікацій об’єктів S3</h2>
      <div class="form-group">
        <TextField
          label="Розклад збереження резервних копій реплікацій об’єктів S3"
          name="registry-backup-obc-cron-expression"
          placeholder="5 4 * * *"
          description="Якщо Ви бажаєте встановити розклад, що відмінний від дефолтного, будь ласка, введіть значення розкладу у Cron-форматі, або вкажіть дефолтне значення за київським часом: 30 17 * * * *"
          :value="data.registryBackup.obc.cronExpression"
          :error="beginValidation ? errors.registryBackupCronExpression : ''"
          @update="val => data.registryBackup.obc.cronExpression = val"
          @change="backupCronExpressionChange"
        />
      </div>
      <div v-show="registryBackupNextDates.length" class="form-group">
        <label>Наступні запуски резервного копіювання (за київським часом)</label>
        <ul class="cron-next-dates">
            <li v-for="date in registryBackupNextDates" v-bind:key="date">
              <Typography variant="bodyText">{{ date }}</Typography>
            </li>
        </ul>
      </div>

      <div>
        <Typography variant="subheading">Місце зберігання резервних копій реплікацій об’єктів S3</Typography>
        <div class="rc-form-backup-obc" v-if="data.registryBackup.obc.endpoint">
          <div>
            <div class="bucket-field">
              <Typography variant="small">Ім’я бакета</Typography>
              <Typography variant="bodyText">{{ data.registryBackup.obc.backupBucket }}</Typography>
            </div>
            <div class="endpoint-field">
              <Typography variant="small">Endpoint</Typography>
              <Typography variant="bodyText">{{ data.registryBackup.obc.endpoint }}</Typography>
            </div>
            <input type="hidden" name="registry-backup-obc-backup-bucket" v-model="data.registryBackup.obc.backupBucket" class="rc-form-input-read-only" disbaled />
            <input type="hidden" name="registry-backup-obc-endpoint" v-model="data.registryBackup.obc.endpoint" class="rc-form-input-read-only" disbaled />
            <input type="hidden" name="registry-backup-obc-login" v-model="data.registryBackup.obc.login" class="rc-form-input-read-only" disbaled />
            <input type="hidden" name="registry-backup-obc-password" v-model="data.registryBackup.obc.password" class="rc-form-input-read-only" disbaled />
          </div>

          <div class="buttom-group">
            <a href="#" @click.stop.prevent="showBackupPlaceModal"
              class="icon-button">
              <img alt="edit button" src="@/assets/img/action-edit.png" />
            </a>
            <a href="#" @click.stop.prevent="showBackupDeletePlaceModal"
              class="icon-button">
              <img alt="delete button" src="@/assets/img/action-delete.png" />
            </a>
          </div>
        </div>
        <div class="rc-form-backup-obc-empty" v-if="!data.registryBackup.obc.endpoint">
          <Typography variant="bodyText">Використовуються значення за замовчуванням, задані при розгортанні реєстру.</Typography>
          <div>
            <a href="#" @click.stop.prevent="showBackupPlaceModal" class="icon-button set-data-button">
              <img alt="edit button" src="@/assets/img/action-edit.png" />
              <Typography variant="buttonText">Задати власні значення</Typography>
            </a>
          </div>
        </div>
      </div>
  </div>

  <RegistryBackupSavePlaceModal
    :backupPlacePopupShow="backupPlacePopupShow"
    :initialData="data.registryBackup.obc"
    @submit-data="backupPlaceSubmit"
    @hide-backup-place-modal="hideBackupPlaceModal"
  />

  <RegistryBackupDeletePlaceModal
    :backupDeletePlacePopupShow="backupDeletePlacePopupShow"
    @submit-data="backupDeletePlaceSubmit"
    @hide-backup-delete-place-modal="hideBackupDeletePlaceModal"
  />
</template>

<style lang="scss">

.endpoint-field {
  margin-top: 16px;
}

.form-group {
  margin-bottom: 24px;

  label {
    font-size: 16px;
    font-weight: bold;
    margin: 0 0 8px 0;
  }
}

.rc-form-backup-obc-empty {
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  padding: 8px 0px;
  gap: 8px;
  border-width: 1px 0px;
  border-style: solid;
  border-color: $grey-border-color;
}

.rc-form-backup-obc {
  width: 100%;
  box-sizing: border-box;
  display: flex;
  justify-content: space-between;
  padding: 8px 0px;
  margin-top: 8px;
  gap: 8px;
  border-width: 1px 0px;
  border-style: solid;
  border-color: $grey-border-color;

  .buttom-group {
    display: flex;
    margin-top: 10px;
  }

  .rc-form-backup-obc-data {
    font-weight: 300;
    font-size: 16px;
    line-height: 24px;
  }
}

.set-data-button {
  margin: 16px;
}

.icon-button {
  display: flex;
  align-items: baseline;
  margin-left: 20px;
}

.icon-button:hover {
  text-decoration: none;
}

.icon-button > img {
  width: 18px;
  height: 18px;
  margin-right: 13px;
}

</style>
