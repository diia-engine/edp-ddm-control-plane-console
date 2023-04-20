<script lang="ts">
import { defineComponent, type PropType } from 'vue';
import axios from 'axios';

interface RegistryTemplateVariables {
  registryValues: any;
  gerritProjects: any;
  model: {
    RegistryGitTemplate: string,
  },
  gerritBranches: string,
}

export default defineComponent({
  props: {
    templateVariables: {
      required: true,
      type: Object as PropType<RegistryTemplateVariables>,
    },
  },
  mounted() {
      this.projectBranches = JSON.parse(this.templateVariables.gerritBranches);
  },
  data() {
    return {
      validated: false,
      templateRequiredError: false,
      registryTemplate: '',
      projectBranches: {} as  { [project: string] : Array<string> },
      branchRequiredError: false,
      branches: [] as Array<string>,
      registryBranch: '',
    };
  },
  methods: {
    changeTemplateProject() {
      this.branches = this.projectBranches[this.registryTemplate];
    },
    validator(tab: any) {
      return new Promise < void  > ((resolve) => {
        tab.validated = false;
        tab.templateRequiredError = false;
        tab.branchRequiredError = false;
        if (tab.registryTemplate === "") {
          tab.templateRequiredError = true;
          return;
        }
        if (tab.registryBranch === "") {
          tab.branchRequiredError = true;
          return;
        }

        axios.get(`/admin/registry/preload-values`, { params: {
            "template": this.registryTemplate,
            "branch": this.registryBranch,
          }
        })
            .then((response) => {
              this.$emit('preloadTemplateData', response.data);
              tab.validated = true;
              resolve();
            })
            .catch((error) => {
              tab.validated = true;
              resolve();
            });
      });
    },
  },
});
</script>

<template>
  <h2>Налаштування шаблону реєстру</h2>
  <div class="rc-form-group"
       :class="{'error': templateRequiredError}">
    <label for="registry-git-template">Шаблон реєстру</label>
    <select name="registry-git-template" id="registry-git-template"
            v-model="registryTemplate" @change="changeTemplateProject">
      <option></option>
      <option v-for="project in templateVariables.gerritProjects" v-bind:key="project.spec.name">
        {{project.spec.name}}
      </option>
    </select>
    <span v-if="templateRequiredError">Обов’язкове поле</span>
  </div>
  <div class="rc-form-group"
       :class="{'error': branchRequiredError}">
    <label for="registry-git-branch">Гілка шаблону реєстру</label>
    <select name="registry-git-branch" id="registry-git-branch" v-model="registryBranch">
      <option v-for="branch in branches" v-bind:key="branch">{{ branch }}</option>
    </select>
    <span v-if="branchRequiredError">Обов’язкове поле</span>
  </div>
</template>