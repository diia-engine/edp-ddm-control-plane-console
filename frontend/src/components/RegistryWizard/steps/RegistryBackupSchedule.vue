<script lang="ts">
export default {
  data() {
    return {
      pageRoot: this.$parent?.$parent as any,
    };
  },
};
</script>

<template>
  <h2>Резервне копіювання</h2>
  <p>Можливість вказати розклад створення резервних копій реєстру та термін їх зберігання.</p>
  <div class="toggle-switch backup-switch">
      <input @change="pageRoot.wizardBackupScheduleChange" v-model="pageRoot.$data.wizard.tabs.backupSchedule.enabled" class="switch-input"
              type="checkbox" id="backup-schedule-switch-input"  name="backup-schedule-enabled" />
      <label for="backup-schedule-switch-input">Toggle</label>
      <span>Налаштувати резервне копіювання</span>
  </div>

  <div v-show="pageRoot.$data.wizard.tabs.backupSchedule.enabled">
      <div class="rc-form-group"
            :class="{'error': (pageRoot.$data.wizard.tabs.backupSchedule.data.cronSchedule == '' || pageRoot.$data.wizard.tabs.backupSchedule.wrongCronFormat) && pageRoot.$data.wizard.tabs.backupSchedule.beginValidation}">
          <label for="cron-schedule">Розклад</label>
          <input @change="pageRoot.wizardCronExpressionChange" placeholder="5 4 * * *" type="text" id="cron-schedule"
                  name="cron-schedule" v-model="pageRoot.$data.wizard.tabs.backupSchedule.data.cronSchedule" />
          <p>Використовується Cron-формат.</p>
          <span v-if="pageRoot.$data.wizard.tabs.backupSchedule.wrongCronFormat && pageRoot.$data.wizard.tabs.backupSchedule.beginValidation">Невірний формат</span>
          <span v-if="pageRoot.$data.wizard.tabs.backupSchedule.data.cronSchedule == '' && pageRoot.$data.wizard.tabs.backupSchedule.beginValidation">Обов’язкове поле</span>
      </div>

      <div v-show="pageRoot.$data.wizard.tabs.backupSchedule.nextLaunches" class="rc-form-group">
          <label>Наступні запуски резервного копіювання (за київським часом)</label>
          <ul class="cron-next-dates">
              <li v-for="date in pageRoot.$data.wizard.tabs.backupSchedule.nextDates" v-bind:key="date">{{ date }}</li>
          </ul>
      </div>

      <div class="rc-form-group"
            :class="{'error': (pageRoot.$data.wizard.tabs.backupSchedule.data.days == '' || pageRoot.$data.wizard.tabs.backupSchedule.wrongDaysFormat) && pageRoot.$data.wizard.tabs.backupSchedule.beginValidation}">
          <label for="cron-schedule-days">Час зберігання (днів)</label>
          <input placeholder="3" type="text" id="cron-schedule-days" name="cron-schedule-days"
                  v-model="pageRoot.$data.wizard.tabs.backupSchedule.data.days" />
          <p>Значення може бути тільки додатним числом та не меншим за 1 день. Рекомендуємо встановити час
              збереження більшим за період між створенням копій.</p>
          <span v-if="pageRoot.$data.wizard.tabs.backupSchedule.wrongDaysFormat && pageRoot.$data.wizard.tabs.backupSchedule.beginValidation">Невірний формат</span>
          <span v-if="pageRoot.$data.wizard.tabs.backupSchedule.data.days == '' && pageRoot.$data.wizard.tabs.backupSchedule.beginValidation">Обов’язкове поле</span>
      </div>
  </div>
</template>
