<script setup lang="ts">
import { toRefs } from 'vue';

interface KeyFormProps {
    wizard: any;
    action?: any;
    model?: any;
}

const props = defineProps<KeyFormProps>();
const { action, wizard, model } = toRefs(props);

</script>
<script lang="ts">
export default {
    methods: {
        wizardTabChanged(key: string) {
            this.$emit('wizardTabChanged', key);
        },
        wizardKeyHardwareDataChanged() {
            this.$emit('wizardKeyHardwareDataChanged');
        },
        wizardAddAllowedKey() {
            this.$emit('wizardAddAllowedKey');
        },
        wizardRemoveAllowedKey(key: string) {
            this.$emit('wizardRemoveAllowedKey', key);
        }
    },
};
</script>

<template>
    <h2>Дані про ключ</h2>

    <input v-if="action === 'edit'" type="checkbox" style="display: none;" name="edit-smtp"
        v-model="wizard.tabs.key.changed" />

    <div class="rc-form-group">
        <label for="key-device-type">Тип носія</label>
        <select @change="wizardTabChanged('key')" v-model="wizard.tabs.key.deviceType" id="key-device-type"
            name="key-device-type">
            <option :selected="model?.KeyDeviceType === 'file'" value="file">Файловий носій</option>
            <option :selected="model?.KeyDeviceType === 'hardware'" value="hardware">Апаратний носій</option>
        </select>
    </div>
    <div class="key-type-hardware key-type-section" v-if="wizard.tabs.key.deviceType == 'hardware'" v-cloak>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteType == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-type">Тип ключа</label>
            <input type="text" @change="wizardKeyHardwareDataChanged" name="remote-type" id="remote-type"
                v-model="wizard.tabs.key.hardwareData.remoteType" />
            <span v-if="wizard.tabs.key.hardwareData.remoteType == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteKeyPWD == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-key-pwd">Пароль ключа</label>
            <input @change="wizardKeyHardwareDataChanged" type="password" id="remote-key-pwd" name="remote-key-pwd"
                v-model="wizard.tabs.key.hardwareData.remoteKeyPWD" />
            <span v-if="wizard.tabs.key.hardwareData.remoteKeyPWD == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteCaName == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-ca-name">Ім'я АЦСК</label>
            <input @change="wizardKeyHardwareDataChanged" type="text" id="remote-ca-name" name="remote-ca-name"
                v-model="wizard.tabs.key.hardwareData.remoteCaName" />
            <span v-if="wizard.tabs.key.hardwareData.remoteCaName == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteCaHost == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-ca-host">Хост АЦСК</label>
            <input @change="wizardKeyHardwareDataChanged" type="text" id="remote-ca-host" name="remote-ca-host"
                v-model="wizard.tabs.key.hardwareData.remoteCaHost" />
            <span v-if="wizard.tabs.key.hardwareData.remoteCaHost == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteCaPort == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-ca-port">Порт АЦСК</label>
            <input @change="wizardKeyHardwareDataChanged" type="number" id="remote-ca-port" name="remote-ca-port"
                v-model="wizard.tabs.key.hardwareData.remoteCaPort" />
            <span v-if="wizard.tabs.key.hardwareData.remoteCaPort == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteSerialNumber == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-serial-number">Серійний номер пристрою</label>
            <input @change="wizardKeyHardwareDataChanged" type="text" id="remote-serial-number" name="remote-serial-number"
                v-model="wizard.tabs.key.hardwareData.remoteSerialNumber" />
            <span
                v-if="wizard.tabs.key.hardwareData.remoteSerialNumber == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteKeyPort == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-key-port">Порт ключа</label>
            <input @change="wizardKeyHardwareDataChanged" type="number" id="remote-key-port" name="remote-key-port"
                v-model="wizard.tabs.key.hardwareData.remoteKeyPort" />
            <span v-if="wizard.tabs.key.hardwareData.remoteKeyPort == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteKeyHost == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-key-host">Хост ключа</label>
            <input type="text" @change="wizardKeyHardwareDataChanged" id="remote-key-host" name="remote-key-host"
                v-model="wizard.tabs.key.hardwareData.remoteKeyHost" />
            <span v-if="wizard.tabs.key.hardwareData.remoteKeyHost == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.hardwareData.remoteKeyMask == '' && wizard.tabs.key.beginValidation }">
            <label for="remote-key-mask">Маска ключа</label>
            <input @change="wizardKeyHardwareDataChanged" type="text" id="remote-key-mask" name="remote-key-mask"
                v-model="wizard.tabs.key.hardwareData.remoteKeyMask" />
            <span v-if="wizard.tabs.key.hardwareData.remoteKeyMask == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group">
            <label for="remote-ini-config">INI конфігурація</label>
            <textarea rows="5" id="remote-ini-config" v-model="wizard.tabs.key.hardwareData.iniConfig"
                name="remote-ini-config"></textarea>
        </div>
    </div>
    <div class="key-type-file key-type-section" v-if="wizard.tabs.key.deviceType == 'file'" v-cloak>
        <div class="rc-form-group" :class="{ 'error': wizard.tabs.key.key6Required }">
            <label for="key6">Файловий ключ (розширення .dat)</label>
            <input @change="wizardTabChanged('key'); wizard.tabs.key.key6Required = false;" ref="key6" type="file"
                name="key6" id="key6" accept=".dat" />
            <span v-if="wizard.tabs.key.key6Required">Обов’язкове поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.fileData.signKeyIssuer == '' && wizard.tabs.key.beginValidation }">
            <label for="sign-key-issuer">АЦСК, що видав ключ</label>
            <input @change="wizardTabChanged('key')" type="text" id="sign-key-issuer" name="sign-key-issuer"
                v-model="wizard.tabs.key.fileData.signKeyIssuer" />
            <span v-if="wizard.tabs.key.fileData.signKeyIssuer == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
        <div class="rc-form-group"
            :class="{ 'error': wizard.tabs.key.fileData.signKeyPWD == '' && wizard.tabs.key.beginValidation }">
            <label for="sign-key-pwd">Пароль до файлового ключа</label>
            <input @change="wizardTabChanged('key')" type="password" id="sign-key-pwd" name="sign-key-pwd"
                v-model="wizard.tabs.key.fileData.signKeyPWD" />
            <span v-if="wizard.tabs.key.fileData.signKeyPWD == '' && wizard.tabs.key.beginValidation">Обов’язкове
                поле</span>
        </div>
    </div>
    <h3>Дані для перевірки ключа</h3>
    <hr />
    <div class="rc-form-group" :class="{ 'error': wizard.tabs.key.caCertRequired }">
        <label for="ca-cert">Публічні сертифікати АЦСК (розширення .p7b)</label>
        <input @change="wizardTabChanged('key'); wizard.tabs.key.caCertRequired = false;" ref="keyCaCert" type="file"
            name="ca-cert" id="ca-cert" accept=".p7b" />
        <span v-if="wizard.tabs.key.caCertRequired">Обов’язкове поле</span>
    </div>
    <div class="rc-form-group" :class="{ 'error': wizard.tabs.key.caJSONRequired }">
        <label for="ca-json">Перелік АЦСК (розширення .json)</label>
        <input @change="wizardTabChanged('key'); wizard.tabs.key.caJSONRequired = false;" ref="keyCaJSON" type="file"
            name="ca-json" id="ca-json" accept=".json" />
        <span v-if="wizard.tabs.key.caJSONRequired">Обов’язкове поле</span>
    </div>
    <h3>Перелік дозволених ключів <button type="button" class="allowed-keys-add" id="allowed-keys-add"
            @click="wizardAddAllowedKey">+</button>
    </h3>
    <hr />
    <div class="rc-form-group allowed-keys-body" id="allowed-keys-body">
        <div class="allowed-keys-row" v-for="ak in wizard.tabs.key.allowedKeys"
            :class="{ 'error': wizard.tabs.key.beginValidation && (ak.serial == '' || ak.issuer == '') }" :key="ak">
            <input @change="wizardTabChanged('key')" name="allowed-keys-issuer[]" v-model="ak.issuer"
                class="allowed-keys-input allowed-keys-issuer" aria-label="key issuer" placeholder="Емітент ключа"
                type="text" />
            <input @change="wizardTabChanged('key')" name="allowed-keys-serial[]"
                class="allowed-keys-input allowed-keys-serial" v-model="ak.serial" aria-label="key serial"
                placeholder="Серійний номер ключа" type="text" />
            <button v-if="ak.removable" class="allowed-keys-remove-btn" type="button"
                @click="wizardRemoveAllowedKey(ak)">-</button>
        </div>
    </div>
</template>
