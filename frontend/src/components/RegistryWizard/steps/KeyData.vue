<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
</script>

<script lang="ts">
import { defineComponent } from 'vue';
import Mustache from "mustache";

export default defineComponent({
  props: {
    registryAction: String,
  },
  methods: {
    renderINITemplate() {
      const iniTemplate = document.getElementById("ini-template")?.innerHTML;
      this.hardwareData.iniConfig = Mustache.render(iniTemplate || '', {
        "CA_NAME": this.hardwareData.remoteCaName,
        "CA_HOST": this.hardwareData.remoteCaHost,
        "CA_PORT": this.hardwareData.remoteCaPort,
        "KEY_SN": this.hardwareData.remoteSerialNumber,
        "KEY_HOST": this.hardwareData.remoteKeyHost,
        "KEY_ADDRESS_MASK": this.hardwareData.remoteKeyMask,
      }).trim();
    },
    removeAllowedKey(item: any) {
      const searchIdx = this.allowedKeys.indexOf(item);
      if (searchIdx !== -1) {
        this.allowedKeys.splice(searchIdx, 1);
      }
    },
    addAllowedKey() {
      this.allowedKeys.push({ issuer: "", serial: "", removable: true });
    },
    hardwareDataChanged() {
      this.renderINITemplate();
      this.changed = true;
    },
    validator() {
      return new Promise<void>((resolve, reject) => {
        if (this.registryAction === "edit" && !this.changed) {
          resolve();
          return true;
        }
        this.renderINITemplate();
        this.validated = false;
        this.beginValidation = true;
        this.key6Required = false;
        let validationFailed = false;

        for (let i = 0; i < this.allowedKeys.length; i++) {
          if (this.allowedKeys[i].issuer === "" ||
              this.allowedKeys[i].serial === "") {
            validationFailed = true;
          }
        }
        if (this.deviceType === "hardware") {
          for (const key in this.hardwareData) {
            if (this.hardwareData[key as keyof typeof this.hardwareData] === "") {
              validationFailed = true;
            }
          }
        } else {
          for (const key in this.fileData) {
            if (this.fileData[key as keyof typeof this.fileData] === "") {
              validationFailed = true;
            }
          }

          const key6 = this.$refs.key6 as any;
          if (key6.files.length === 0) {
            this.key6Required = true;
            validationFailed = true;
          }
        }

        if (validationFailed) {
          reject();
          return false;
        }

        this.beginValidation = false;
        this.validated = true;
        resolve();
        return true;
      });
    },
  },
  data() {
    return {
      validated: false,
      changed: false,
      allowedKeys: [{ issuer: "", serial: "", removable: false }],
      beginValidation: false,
      key6Required: false,
      deviceType: "file",
      hardwareData: {
        remoteType: "криптомод. ІІТ Гряда-301",
        remoteKeyPWD: "",
        remoteCaName: "",
        remoteCaHost: "",
        remoteCaPort: "",
        remoteSerialNumber: "",
        remoteKeyPort: "",
        remoteKeyHost: "",
        remoteKeyMask: "",
        iniConfig: "",
      }, fileData: {
        signKeyIssuer: "",
        signKeyPWD: "",
      }
    };
  },
});
</script>

<style scoped>
    .key-data-page-description {
      margin-bottom: 32px;
    }
</style>

<template>
  <div class="form-group">
    <Typography variant="h3">Дані про ключ</Typography>
  </div>
  <Typography variant="bodyText" class="key-data-page-description">Внесені ключі системного підпису та КЕП користувачів будуть застосовані для
    налаштувань поточного реєстру.</Typography>

  <input type="checkbox" style="display: none;" v-model="changed" name="key-data-changed" />
  <div class="rc-form-group">
    <label for="key-device-type">Тип носія</label>
    <select @change="changed = true;" v-model="deviceType" id="key-device-type"
            name="key-device-type">
      <option value="file">Файловий носій</option>
      <option value="hardware">Апаратний носій</option>
    </select>
  </div>
  <div class="key-type-hardware key-type-section" v-if="deviceType === 'hardware'" v-cloak>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteType === '' && beginValidation }">
      <label for="remote-type">Тип ключа</label>
      <input type="text" @change="hardwareDataChanged" name="remote-type" id="remote-type"
             v-model="hardwareData.remoteType" />
      <span v-if="hardwareData.remoteType === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteKeyPWD === '' && beginValidation }">
      <label for="remote-key-pwd">Пароль ключа</label>
      <input @change="hardwareDataChanged" type="password" id="remote-key-pwd" name="remote-key-pwd"
             v-model="hardwareData.remoteKeyPWD" />
      <span v-if="hardwareData.remoteKeyPWD === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteCaName === '' && beginValidation }">
      <label for="remote-ca-name">Ім'я АЦСК</label>
      <input @change="hardwareDataChanged" type="text" id="remote-ca-name" name="remote-ca-name"
             v-model="hardwareData.remoteCaName" />
      <span v-if="hardwareData.remoteCaName === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteCaHost === '' && beginValidation }">
      <label for="remote-ca-host">Хост АЦСК</label>
      <input @change="hardwareDataChanged" type="text" id="remote-ca-host" name="remote-ca-host"
             v-model="hardwareData.remoteCaHost" />
      <span v-if="hardwareData.remoteCaHost === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteCaPort === '' && beginValidation }">
      <label for="remote-ca-port">Порт АЦСК</label>
      <input @change="hardwareDataChanged" type="number" id="remote-ca-port" name="remote-ca-port"
             v-model="hardwareData.remoteCaPort" />
      <span v-if="hardwareData.remoteCaPort === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteSerialNumber === '' && beginValidation }">
      <label for="remote-serial-number">Серійний номер пристрою</label>
      <input @change="hardwareDataChanged" type="text" id="remote-serial-number" name="remote-serial-number"
             v-model="hardwareData.remoteSerialNumber" />
      <span
          v-if="hardwareData.remoteSerialNumber === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteKeyPort === '' && beginValidation }">
      <label for="remote-key-port">Порт ключа</label>
      <input @change="hardwareDataChanged" type="number" id="remote-key-port" name="remote-key-port"
             v-model="hardwareData.remoteKeyPort" />
      <span v-if="hardwareData.remoteKeyPort === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteKeyHost === '' && beginValidation }">
      <label for="remote-key-host">Хост ключа</label>
      <input type="text" @change="hardwareDataChanged" id="remote-key-host" name="remote-key-host"
             v-model="hardwareData.remoteKeyHost" />
      <span v-if="hardwareData.remoteKeyHost === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': hardwareData.remoteKeyMask === '' && beginValidation }">
      <label for="remote-key-mask">Маска ключа</label>
      <input @change="hardwareDataChanged" type="text" id="remote-key-mask" name="remote-key-mask"
             v-model="hardwareData.remoteKeyMask" />
      <span v-if="hardwareData.remoteKeyMask === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group">
      <label for="remote-ini-config">INI конфігурація</label>
      <textarea rows="5" id="remote-ini-config" v-model="hardwareData.iniConfig"
                name="remote-ini-config"></textarea>
    </div>
  </div>
  <div class="key-type-file key-type-section" v-if="deviceType === 'file'" v-cloak>
    <div class="rc-form-group" :class="{ 'error': key6Required }">
      <label for="key6">Файловий ключ (розширення .dat)</label>
      <input @change="changed = true; key6Required = false;" ref="key6" type="file"
             name="key6" id="key6" accept=".dat" />
      <span v-if="key6Required">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': fileData.signKeyIssuer === '' && beginValidation }">
      <label for="sign-key-issuer">АЦСК, що видав ключ</label>
      <input @change="changed = true;" type="text" id="sign-key-issuer" name="sign-key-issuer"
             v-model="fileData.signKeyIssuer" />
      <span v-if="fileData.signKeyIssuer === '' && beginValidation">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group"
         :class="{ 'error': fileData.signKeyPWD === '' && beginValidation }">
      <label for="sign-key-pwd">Пароль до файлового ключа</label>
      <input @change="changed = true;" type="password" id="sign-key-pwd" name="sign-key-pwd"
             v-model="fileData.signKeyPWD" />
      <span v-if="fileData.signKeyPWD === '' && beginValidation">Обов’язкове поле</span>
    </div>
  </div>

  <h3>Перелік дозволених ключів <button type="button" class="allowed-keys-add" id="allowed-keys-add"
                                        @click="addAllowedKey">+</button>
  </h3>
  <hr />
  <div class="rc-form-group allowed-keys-body" id="allowed-keys-body">
    <div class="allowed-keys-row" v-for="ak in allowedKeys" :key="ak.serial"
         :class="{ 'error': beginValidation && (ak.serial === '' || ak.issuer === '') }">
      <input @change="changed = true;" name="allowed-keys-issuer[]" v-model="ak.issuer"
             class="allowed-keys-input allowed-keys-issuer" aria-label="key issuer" placeholder="Емітент ключа"
             type="text" />
      <input @change="changed = true;" name="allowed-keys-serial[]"
             class="allowed-keys-input allowed-keys-serial" v-model="ak.serial" aria-label="key serial"
             placeholder="Серійний номер ключа" type="text" />
      <button v-if="ak.removable" class="allowed-keys-remove-btn" type="button"
              @click="removeAllowedKey(ak)">-</button>
    </div>
  </div>
</template>
