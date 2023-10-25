<script lang="ts">
import { defineComponent } from 'vue';
import KeyData from '@/components/RegistryWizard/steps/KeyData.vue';
import KeyVerification from '@/components/RegistryWizard/steps/KeyVerification.vue';
import KeysManagementBlockVue from './KeysManagementBlock.vue';
import type { OutKey } from '@/types/cluster';

export default defineComponent({
    inject: ["TEMPLATE_VARIABLES"],
    props: {
      activeTab: String,
      region: String
    },
    components: {
    KeyData,
    KeyVerification,
    KeysManagementBlockVue,
},
    data() {
      const dataArgs = this.TEMPLATE_VARIABLES as any;
        return {
            disabled: false,
            iniTemplate: document.getElementById('ini-template')!.innerText,
            digitalSignature: JSON
              .parse(
                dataArgs?.values || '{}'
              )?.DigitalSignature,
            usedKeys: dataArgs.usedKeys,
            registries: dataArgs.registries,
        };
    },
    methods: {
        submit(event: any) {
            event.preventDefault();
            this.disabled = true;

            let validatorTab = this.$refs.keyDataTab as any;
            if (this.activeTab === 'dataAboutKeyVerification') {
              validatorTab = this.$refs.keyVerificationTab as any;
            }
            if (this.activeTab === 'keysManagement') {
              validatorTab = this.$refs.keysManagementTab as any;
            }

            const clusterKeyDataForm = this.$refs.clusterKeyDataForm as HTMLFormElement;
            validatorTab.validator().then(() => {
              this.$nextTick(() => {
                if (this.activeTab === 'keysManagement') {
                  clusterKeyDataForm.onformdata = (e) => {
                    const formData = e.formData;
                    formData.append('tab', 'keysManagement');
                    const { keys, osplmIni } = validatorTab.getData() as { keys: OutKey[], osplmIni: string };
                    formData.append('osplmIni', osplmIni);
                    formData.append('keysJSON', JSON.stringify(keys.map(({fileKeyFile, ...key}) => {
                      if (typeof fileKeyFile === 'string') {
                        return { ...key, fileKeyFile };
                      }
                      return key;
                    })));
                    keys.forEach((signKey) => {
                      if (signKey.fileKeyFile && typeof signKey.fileKeyFile !== 'string') {
                        formData.append(signKey.fileKeyName, signKey.fileKeyFile);
                      }
                    });
                  };
                }
                HTMLFormElement.prototype.submit.call(clusterKeyDataForm);
              });
            }).catch(() => {
              this.disabled = false;
            });
        },
    }
});
</script>

<style scoped lang="scss">
  .rc-form-group button {
    margin-top: 20px;
  }
  .cluster-key-form {
    margin: 0 0 16px 0;
    padding: 0 8px 8px 0;
  }
  .w512 {
    width: $wizard-width;
  }
  .fullWidth {
    width: 100%;
  }
</style>

<template>
    <form @submit="submit" ref="clusterKeyDataForm" id="clusterKeyDataForm" class="cluster-key-form"
          enctype="multipart/form-data" method="post"
        action="/admin/cluster/key">

      <div v-if="activeTab === 'dataAboutKey'" class="w512">
        <KeyData
            registry-action="create"
            :page-description="$t('domains.cluster.clusterKey.text.addedUsersKeysDescription')"
            :region="region"
            ref="keyDataTab" />
      </div>
      <div v-if="activeTab === 'keysManagement'" class="fullWidth">
        <KeysManagementBlockVue
            ref="keysManagementTab"
            :ini-template="iniTemplate"
            :digitalSignature="digitalSignature"
            :used-keys="usedKeys"
            :registries="registries"
          />
      </div>

      <div v-if="activeTab === 'dataAboutKeyVerification'" class="w512">
        <KeyVerification
            registry-action="create"
            :page-description="$t('domains.cluster.clusterKey.text.addedUsersCertificatesDescription')"
            :region="region"
            ref="keyVerificationTab" />
      </div>

        <div class="rc-form-group">
            <button type="submit" name="submit" :disabled="disabled" onclick="window.localStorage.setItem('mr-scroll', 'true');">
              {{ $t('actions.confirm') }}
            </button>
        </div>
    </form>
</template>
