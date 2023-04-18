<script lang="ts">
import $ from 'jquery';
import {defineComponent} from 'vue';

export default defineComponent({
    props: {
      cidrPopupShow: Boolean,
      title: String,
      subTitle: String,
    },
    data() {
        return {
            disabled: false,
            editCIDR: '',
            cidrFormatError: false,
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

            this.$emit('cidrAdded', this.editCIDR);
            this.hideCIDRForm();
        },
        hideCIDRForm() {
            this.$emit('update:cidrPopupShow', false);
            $("body").css("overflow", "scroll");
        },
    },
    watch: {
        cidrPopupShow() {
            this.cidrFormatError = false;
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
            <p>{{ title }}</p>
            <a href="#" @click.stop.prevent="hideCIDRForm" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <form @submit.prevent="createCIDR" id="cidr-form" method="post" action="">
            <div class="popup-body">
                <p class="popup-error" v-cloak v-if="cidrFormatError">Перевірте формат IP-адреси</p>
                <div class="rc-form-group">
                    <label for="cidr-value">{{ subTitle }}</label>
                    <input id="cidr-value" type="text" v-model="editCIDR" />
                    <p>Допустимі символи "0-9", "/", ".". Приклад: 172.16.0.0/12.</p>
                </div>
            </div>
            <div class="popup-footer active">
                <a href="#" id="cidr-cancel" class="hide-popup" @click="hideCIDRForm">відмінити</a>
                <button class="submit-green" value="submit" name="cidr-apply" type="submit"
                    :disabled="disabled && !cidrFormatError">Підтвердити</button>
            </div>
        </form>
    </div>
</template>
