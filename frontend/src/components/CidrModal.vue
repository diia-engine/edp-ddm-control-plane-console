<script lang="ts">
import $ from 'jquery';
import {defineComponent} from 'vue';

export default defineComponent({
    props: {
      cidrPopupShow: Boolean,
      currentCIDR: Array,
      currentCIDRValue: Object,
      cidrChanged: Boolean,
    },
    emits: ['update:cidrPopupShow', 'update:currentCIDR', 'update:currentCIDRValue', 'update:cidrChanged'],
    data() {
        return {
            disabled: false,
            editCIDR: '',
            cidrFormatError: false,
            dataCurrentCIDR: this.currentCIDR,
            dataCurrentCIDRValue: this.currentCIDRValue,
        };
    },
    methods: {
        createCIDR(e: any) {
            this.disabled = true;
            let cidrVal = String(this.editCIDR).toLowerCase();
            if (cidrVal !== "0.0.0.0/0" && !cidrVal.
            match(/^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:\/[0-2]\d|\/3[0-2])?$/)) {
              this.cidrFormatError = true;
              return;
            }


            this.dataCurrentCIDR?.push(this.editCIDR);
            this.$emit('update:currentCIDR', this.dataCurrentCIDR);

            if (this.dataCurrentCIDRValue) {
              this.dataCurrentCIDRValue.value = JSON.stringify(this.dataCurrentCIDR);
              this.$emit('update:currentCIDRValue', this.dataCurrentCIDRValue);
            }

            this.hideCIDRForm();
            this.$emit('update:cidrChanged', true);
        },
        hideCIDRForm() {
            this.$emit('update:cidrPopupShow', false);
            $("body").css("overflow", "scroll");
        },
    },
    watch: {
        cidrPopupShow() {
            this.disabled = false;
            this.editCIDR = '';
        }
    }
});
</script>

<template>
    <div class="popup-backdrop visible" v-cloak v-if="cidrPopupShow"></div>
    <div class="popup-window admin-window visible" v-cloak v-if="cidrPopupShow">
        <div class="popup-header">
            <p>Додати CIDR</p>
            <a href="#" @click.stop.prevent="hideCIDRForm" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <form @submit.prevent="createCIDR" id="cidr-form" method="post" action="">
            <div class="popup-body">
                <p class="popup-error" v-cloak v-if="cidrFormatError">Перевірте формат IP-адреси</p>
                <div class="rc-form-group">
                    <label for="cidr-value">IP-адреси та маски</label>
                    <input id="cidr-value" type="text" v-model="editCIDR" />
                    <p>Допустимі символи "0-9", "/", ".". Приклад: 172.16.0.0/12.</p>
                </div>
            </div>
            <div class="popup-footer active">
                <a href="#" id="cidr-cancel" class="hide-popup" @click="hideCIDRForm">відмінити</a>
                <button value="submit" name="cidr-apply" type="submit"
                    :disabled="disabled && !cidrFormatError">Підтвердити</button>
            </div>
        </form>
    </div>
</template>
