<script lang="ts">
import { defineComponent, type PropType, ref } from 'vue';
import * as Yup from 'yup';
import { useForm, useField } from 'vee-validate';
import TextField from '@/components/common/TextField.vue';
import Typography from '@/components/common/Typography.vue';
import IconButton from '@/components/common/IconButton.vue';
import ToggleSwitch from '@/components/common/ToggleSwitch.vue';
import Banner from '@/components/common/Banner.vue';
import { jsonDiff } from '@/utils/registry';
import type { RegistryResource } from '@/types/registry';
import { REGISTRY_COMPONENTS } from '@/types/registry';
import { cloneDeep } from 'lodash';

interface WizardTemplateVariables {
  registryValues: any;
  defaultRegistryValues: any;
}

export default defineComponent({
  components: { TextField, Typography, IconButton, ToggleSwitch, Banner },
  setup() {
    const registryResources = ref({
      encoded: '',
      category: '' as any,
      categories: [] as Array<RegistryResource>,
      listOfCategoryNames: Object.keys(REGISTRY_COMPONENTS),
    });
    const crunchyPostgres = ref({
      maxConnections: '',
      storageSize: '',
    });
    const diffRegistryResourcesAndDefaultResources = ref<string[]>([]);
    const resourcesWithoutHpa = [
      REGISTRY_COMPONENTS.geoServer,
      REGISTRY_COMPONENTS.redis,
      REGISTRY_COMPONENTS.sentinel,
    ];
    const defaultEmptyResource = {
      config: {
        istio: {
          sidecar: {
            enabled: true,
            resources: {
              requests: {
                cpu: '',
                memory: '',
              },
              limits: {
                cpu: '',
                memory: '',
              },
            },
          },
        },
        container: {
          resources: {
            requests: {
              cpu: '',
              memory: '',
            },
            limits: {
              cpu: '',
              memory: '',
            },
          },
          envVars: [{ name: '', value: '' }],
        },
      },
    };

    const commonValidationSchema = Yup.object().shape({
      container: Yup.object().shape({
        envVars: Yup.array().of(
          Yup.object().shape({
            name: Yup.string().test((value, context) => {
              return !(!value && context.parent.value);
            }),
            value: Yup.string().test((value, context) => {
              return !(!value && context.parent.name);
            }),
          })
        ),
      }),
    });

    const validationSchemaWithHpa = Yup.object().shape({
      hpa: Yup.object().shape({
        enabled: Yup.bool(),
        minReplicas: Yup.number().when('enabled', {
          is: true,
          then: (schema) => schema.required().min(1).integer(),
        }),
        maxReplicas: Yup.number().when('enabled', {
          is: true,
          then: (schema) =>
            schema.required().min(1).integer().moreThan(Yup.ref('minReplicas')),
        }),
      }),
      replicas: Yup.number().when('hpa.enabled', {
        is: false,
        then: (schema) => schema.required().min(1).integer(),
      }),
    });

    const validationSchema = Yup.object({
      registryResourcesForm: Yup.array().of(
        Yup.object().shape({
          name: Yup.string(),
          config: Yup.object().when('name', {
            is: (name: string) => {
              return (
                name === REGISTRY_COMPONENTS.kafkaApi ||
                name === REGISTRY_COMPONENTS.restApi
              );
            },
            then: (schema) => {
              return schema
                .shape({
                  datasource: Yup.object().shape({
                    maxPoolSize: Yup.number().required().min(1).integer(),
                  }),
                })
                .concat(validationSchemaWithHpa)
                .concat(commonValidationSchema);
            },
            otherwise: (schema) => {
              return schema.when('name', {
                is: (name: REGISTRY_COMPONENTS) => {
                  return resourcesWithoutHpa.includes(name);
                },
                then: (schema) =>
                  schema
                    .shape({
                      replicas: Yup.number().required().min(1).integer(),
                    })
                    .concat(commonValidationSchema),
                otherwise: (schema) =>
                  schema
                    .concat(validationSchemaWithHpa)
                    .concat(commonValidationSchema),
              });
            },
          }),
        })
      ),
    });

    const { errors, validate } = useForm({
      validationSchema,
    });

    const { value: registryResourcesForm } = useField<RegistryResource[]>(
      'registryResourcesForm'
    );
    registryResourcesForm.value = [];

    function validator() {
      return new Promise((resolve) => {
        validate().then((res) => {
          if (res.valid) {
            resolve(true);
          }
        });
      });
    }

    return {
      errors,
      validator,
      registryResources,
      crunchyPostgres,
      diffRegistryResourcesAndDefaultResources,
      registryResourcesForm,
      resourcesWithoutHpa,
      defaultEmptyResource,
      REGISTRY_COMPONENTS,
    };
  },
  props: {
    templatePreloadedData: Object,
    formSubmitted: Boolean,
    isEditAction: Boolean,
    templateVariables: {
      required: true,
      type: Object as PropType<WizardTemplateVariables>,
    },
  },
  watch: {
    templatePreloadedData(data: any) {
      this.preloadRegistryResources(data);
    },
    formSubmitted() {
      this.encodeRegistryResources();
    },
  },
  methods: {
    decodeResourcesEnvVars(inEnvVars: Record<string, unknown>) {
      const envVars = [];
      for (const j in inEnvVars) {
        envVars.push({
          name: j,
          value: inEnvVars[j],
        });
      }
      return envVars;
    },
    addEnvVar(envVars: Array<Record<string, unknown>>, event: any) {
      event.preventDefault();
      envVars.push({ name: '', value: '' });
    },
    removeEnvVar(
      envVars: Array<Record<string, unknown>>,
      env: Record<string, unknown>
    ) {
      envVars.splice(envVars.indexOf(env), 1);
    },
    removeResource(cat: RegistryResource, event: any) {
      event.preventDefault();
      this.registryResources.listOfCategoryNames.unshift(cat.name);

      this.registryResources.categories.splice(
        this.registryResources.categories.indexOf(cat),
        1
      );
      this.registryResourcesForm.splice(
        this.registryResourcesForm.indexOf(cat),
        1
      );
    },
    addResource() {
      const category = this.registryResources.categories.find(
        (c) => c.name === this.registryResources.category
      );
      const indexListOfCategoryNames =
        this.registryResources.listOfCategoryNames.indexOf(
          this.registryResources.category
        );
      if (!category) {
        const emptyResource = {
          name: this.registryResources.category,
          config: {
            ...cloneDeep(this.defaultEmptyResource.config),
            ...(!this.resourcesWithoutHpa.includes(
              this.registryResources.category
            ) && {
              hpa: {
                enabled: false,
                maxReplicas: 3,
                minReplicas: 1,
              },
            }),
            ...((this.registryResources.category ===
              REGISTRY_COMPONENTS.kafkaApi ||
              this.registryResources.category ===
                REGISTRY_COMPONENTS.restApi) && {
              datasource: {
                maxPoolSize: 10,
              },
            }),
            replicas: 1,
          },
        };

        this.registryResources.categories.unshift(emptyResource);
        this.registryResourcesForm.push(emptyResource);
        this.registryResources.listOfCategoryNames.splice(
          indexListOfCategoryNames,
          1
        );
        this.registryResources.category = '';
        return;
      }
      this.registryResourcesForm.push(category);
      this.registryResources.listOfCategoryNames.splice(
        indexListOfCategoryNames,
        1
      );
      this.registryResources.category = '';
    },
    encodeRegistryResources() {
      const prepare = {} as Record<string, unknown>;
      this.registryResources.categories.forEach((el: any) => {
        const cloneEL = JSON.parse(JSON.stringify(el));
        const envVars = {} as Record<string, unknown>;
        cloneEL.config.container.envVars.forEach(function (el: any) {
          envVars[el.name] = el.value;
        });
        cloneEL.config.container.envVars = envVars;
        prepare[cloneEL.name] = { ...cloneEL.config };
      });

      this.cleanEmptyProperties(prepare);
      this.registryResources.encoded = JSON.stringify(prepare);
    },
    cleanEmptyProperties(obj: Record<string, unknown>) {
      if (this.isObject(obj)) {
        for (const key in obj) {
          if (this.isObject(obj[key])) {
            this.cleanEmptyProperties(obj[key] as Record<string, unknown>);
            if (Object.keys(obj[key] as Record<string, unknown>).length === 0) {
              delete obj[key];
            }
          } else if (obj[key] === '') {
            delete obj[key];
          }
        }
      }
    },
    mergeResource(data: Record<string, unknown>) {
      const emptyResource = {...cloneDeep(this.defaultEmptyResource.config)};
      this.mergeDeep(emptyResource, data);
      return emptyResource;
    },
    preloadRegistryResources(values: any) {
      const crunchyPostgres = values?.global?.crunchyPostgres;
      if (crunchyPostgres) {
        this.crunchyPostgres.maxConnections =
          crunchyPostgres.postgresql?.parameters?.max_connections;
        this.crunchyPostgres.storageSize = crunchyPostgres.storageSize;
      }
      const data = values?.global?.registry;

      if (!data) {
        return;
      }

      for (const i in data) {
        if (
          'container' in data[i] &&
          this.isObject(data[i].container) &&
          'envVars' in data[i].container
        ) {
          data[i].container.envVars = this.decodeResourcesEnvVars(
            data[i].container.envVars
          );
        }
        const mergedData = this.mergeResource(data[i]);

        this.registryResources.categories.push({
          name: i as any,
          config: mergedData,
        });
      }

      if (
        this.diffRegistryResourcesAndDefaultResources.length &&
        this.isEditAction
      ) {
        this.registryResources.categories.forEach((cat) => {
          if (this.diffRegistryResourcesAndDefaultResources.includes(cat.name)) {
            this.registryResourcesForm.push(cat);
          }
        });
        this.registryResources.listOfCategoryNames =
          this.registryResources.listOfCategoryNames.filter(
            (cat) =>
              !this.diffRegistryResourcesAndDefaultResources.includes(cat)
          );
      }
    },
    isObject(item: unknown) {
      return item && typeof item === 'object' && !Array.isArray(item);
    },
    mergeDeep(target: any, ...sources: any[]): any {
      if (!sources.length) return target;
      const source = sources.shift();
      if (this.isObject(target) && this.isObject(source)) {
        for (const key in source) {
          if (source[key] === null) {
            continue;
          }
          if (this.isObject(source[key])) {
            if (!target[key]) Object.assign(target, { [key]: {} });
            this.mergeDeep(target[key], source[key]);
          } else {
            Object.assign(target, { [key]: source[key] });
          }
        }
      }
      return this.mergeDeep(target, ...sources);
    },
    preloadDiffResult() {
      if (this.isEditAction) {
        const data = this.changeMaxPoolSizeToNumber(this.templateVariables.registryValues.global?.registry);
        const baseData = this.templateVariables.defaultRegistryValues?.global?.registry;

        this.cleanEmptyProperties(data);
        this.cleanEmptyProperties(baseData);

        this.diffRegistryResourcesAndDefaultResources = jsonDiff(
          data,
          baseData
        );
      }
    },
    showBanner(categoryName: string) {
      return (
        this.isEditAction &&
        (categoryName === REGISTRY_COMPONENTS.restApi ||
          categoryName === REGISTRY_COMPONENTS.kafkaApi ||
          categoryName === REGISTRY_COMPONENTS.soapApi)
      );
    },
    changeMaxPoolSizeToNumber(categories: Record<string, any>) {
      for (const i in categories) {
        if (categories[i]?.datasource?.maxPoolSize) {
          categories[i].datasource.maxPoolSize = parseInt(categories[i].datasource.maxPoolSize);
        }
      }
      return categories;
    }
  },
  mounted() {
    this.preloadDiffResult();
    this.preloadRegistryResources(this.templateVariables.registryValues);
  },
});
</script>
<template>
  <Typography variant="h3" class="h3">Ресурси реєстру</Typography>
  <Typography variant="bodyText">
    Ви можете додати окремі компоненти до реєстру.
  </Typography>
  <input type="hidden" name="resources" :value="registryResources.encoded" />

  <div class="registry-resources">
    <div class="rc-form-group crunchy-postgres">
      <Typography variant="h3" class="mb24">Crunchy Postgres</Typography>
      <TextField
        label="Max Connections"
        name="crunchy-postgres-max-connections"
        v-model="crunchyPostgres.maxConnections"
      />
      <TextField
        label="Storage Size"
        name="crunchy-postgres-storage-size"
        v-model="crunchyPostgres.storageSize"
      />
    </div>

    <hr class="divider" />

    <div
      class="cat-line"
      v-for="(cat, idx) in registryResourcesForm"
      v-bind:key="cat.name"
    >
      <Typography variant="h3" class="category-name">
        {{ cat.name }}
        <IconButton @onClick="removeResource(cat, $event)">
          <img src="@/assets/svg/trash.svg" />
        </IconButton>
      </Typography>
      <Banner
        v-if="showBanner(cat.name)"
        classes="mb32"
        description="Якщо потрібно одразу застосувати внесені зміни, необхідно викликати оновлення дата-моделі."
      />
      <template
        v-if="
          !resourcesWithoutHpa.includes(cat.name) &&
          typeof cat.config?.hpa?.enabled === 'boolean'
        "
      >
        <Typography variant="h5" class="upperText">
          HPA (Автоматичне горизонтальне масштабування)
        </Typography>
        <ToggleSwitch
          :name="`cat[${idx}].config.hpa.enabled`"
          label="Enable HPA (Автоматичне горизонтальне масштабування)"
          v-model="cat.config.hpa.enabled"
          classes="mt24"
        />
        <div class="rc-form-group mt24" v-if="!cat.config?.hpa?.enabled">
          <TextField
            required
            type="number"
            label="Replicas Amount"
            :name="`cat[${idx}].config.replicas`"
            v-model="cat.config.replicas"
            :error="errors[`registryResourcesForm[${idx}].config.replicas`]"
          />
        </div>
        <div
          class="rc-form-group mt24 rc-form-group-horz"
          v-if="cat.config?.hpa?.enabled"
        >
          <TextField
            required
            type="number"
            label="Min Replicas"
            :name="`cat[${idx}].config.hpa.minReplicas`"
            v-model="cat.config.hpa.minReplicas"
            :error="
              errors[`registryResourcesForm[${idx}].config.hpa.minReplicas`]
            "
            rootClass="mb0"
          />
          <div class="separator">–</div>
          <TextField
            required
            type="number"
            label="Max Replicas"
            :name="`cat[${idx}].config.hpa.maxReplicas`"
            v-model="cat.config.hpa.maxReplicas"
            :error="
              errors[`registryResourcesForm[${idx}].config.hpa.maxReplicas`]
            "
            rootClass="mb0"
          />
        </div>
      </template>
      <template v-else>
        <TextField
          v-if="cat.name !== REGISTRY_COMPONENTS.geoServer"
          required
          type="number"
          label="Replicas Amount"
          :name="`cat[${idx}].config.replicas`"
          v-model="cat.config.replicas"
          :error="errors[`registryResourcesForm[${idx}].config.replicas`]"
        />
      </template>
      <div class="rc-form-group mt32">
        <Typography variant="h5" class="upperText">Container limits</Typography>
        <Typography variant="small" class="mt16">
          Вказуйте значення та розмірність. Наприклад, “100m” для CPU
          (millicores) та “400Mi” для RAM (mebibytes).
        </Typography>
        <div class="rc-form-group mt24 rc-form-group-horz">
          <TextField
            label="CPU requests"
            :name="`cat[${idx}].config.container.resources.requests.cpu`"
            v-model="cat.config.container.resources.requests.cpu"
            rootClass="mb0"
          />
          <div class="separator">&</div>
          <TextField
            label="Memory requests"
            :name="`cat[${idx}].config.container.resources.requests.memory`"
            v-model="cat.config.container.resources.requests.memory"
            rootClass="mb0"
          />
        </div>
        <div class="rc-form-group rc-form-group-horz">
          <TextField
            label="CPU limits"
            :name="`cat[${idx}].config.container.resources.limits.cpu`"
            v-model="cat.config.container.resources.limits.cpu"
            rootClass="mb0"
          />
          <div class="separator">&</div>
          <TextField
            label="Memory limits"
            :name="`cat[${idx}].config.container.resources.limits.memory`"
            v-model="cat.config.container.resources.limits.memory"
            rootClass="mb0"
          />
        </div>
      </div>
      <div class="rc-form-group mt32" v-if="cat.name !== REGISTRY_COMPONENTS.redis">
        <Typography variant="h5" class="upperText">Istio sidecar</Typography>
        <ToggleSwitch
          :name="`cat[${idx}].config.istio.sidecar.enabled`"
          label="Enable Istio sidecar"
          v-model="cat.config.istio.sidecar.enabled"
          id="istio-sidecar-enabled"
          classes="mt24"
        />
        <template v-if="cat.config.istio.sidecar.enabled">
          <div class="rc-form-group rc-form-group-horz mt24">
            <TextField
              label="CPU requests"
              :name="`cat[${idx}].config.istio.sidecar.resources.requests.cpu`"
              v-model="cat.config.istio.sidecar.resources.requests.cpu"
              rootClass="mb0"
            />
            <div class="separator">&</div>
            <TextField
              label="Memory requests"
              :name="`cat[${idx}].config.istio.sidecar.resources.requests.memory`"
              v-model="cat.config.istio.sidecar.resources.requests.memory"
              rootClass="mb0"
            />
          </div>
          <div class="rc-form-group rc-form-group-horz">
            <TextField
              label="CPU limits"
              :name="`cat[${idx}].config.istio.sidecar.resources.limits.cpu`"
              v-model="cat.config.istio.sidecar.resources.limits.cpu"
              rootClass="mb0"
            />
            <div class="separator">&</div>
            <TextField
              label="Memory limits"
              :name="`cat[${idx}].config.istio.sidecar.resources.limits.memory`"
              v-model="cat.config.istio.sidecar.resources.limits.memory"
              rootClass="mb0"
            />
          </div>
        </template>
      </div>

      <div class="rc-form-group mt32">
        <Typography variant="h5" class="upperText">Змінні оточення</Typography>
        <div class="rc-form-group rc-form-group-horz mb0 mt24">
          <Typography variant="bodyText" class="env-name">Name</Typography>
          <Typography variant="bodyText" class="env-value">Value</Typography>
        </div>
        <div
          class="rc-form-group rc-form-group-horz mb8"
          v-for="(env, index) in cat.config.container.envVars"
          v-bind:key="index"
        >
          <TextField
            :name="`env[${index}].name`"
            v-model="env.name"
            rootClass="mb0"
            :error="
              errors[
                `registryResourcesForm[${idx}].config.container.envVars[${index}].name`
              ]
            "
          />
          <TextField
            :name="`env[${index}].value`"
            v-model="env.value"
            rootClass="mb0"
            :error="
              errors[
                `registryResourcesForm[${idx}].config.container.envVars[${index}].value`
              ]
            "
          />
          <IconButton
            @click="removeEnvVar(cat.config.container.envVars, env)"
            class="mt8"
          >
            <img src="@/assets/svg/close.svg" />
          </IconButton>
        </div>
        <div class="env-vars">
          <a
            @click="addEnvVar(cat.config.container.envVars, $event)"
            href="#"
            class="env-add-lnk"
          >
            <Typography variant="small">+ Додати змінну оточення</Typography>
          </a>
        </div>
      </div>

      <div
        v-if="
          (cat.name === REGISTRY_COMPONENTS.restApi || cat.name === REGISTRY_COMPONENTS.kafkaApi) &&
          typeof cat.config?.datasource?.maxPoolSize === 'number'
        "
        class="rc-form-group mb24"
      >
        <Typography variant="h5" class="upperText mb24">
          Database connection parameters
        </Typography>
        <TextField
          type="number"
          label="Maximum pool size"
          description="Допустиме значення параметру > 0"
          :name="`cat[${idx}].config.datasource.maxPoolSize`"
          v-model="cat.config.datasource.maxPoolSize"
          :error="
            errors[
              `registryResourcesForm[${idx}].config.datasource.maxPoolSize`
            ]
          "
        />
      </div>
      <hr class="divider" />
    </div>

    <div class="rc-form-group res-cat-select">
      <select v-model="registryResources.category">
        <option disabled selected value="">Оберіть компонент</option>
        <option
          v-for="name in registryResources.listOfCategoryNames"
          v-bind:key="name"
        >
          {{ name }}
        </option>
      </select>
      <button
        @click.prevent="addResource"
        :class="['btn', { 'btn-active': registryResources.category }]"
      >
        Додати
      </button>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.h3 {
  margin-bottom: 16px;
}

.divider {
  margin-top: 32px;
  margin-bottom: 32px;
  border-color: $grey3;
}
.btn {
  width: auto;
  padding: 8px 16px;
  height: auto;
  margin-left: 16px;
  background-color: $grey-border-color;
  pointer-events: none;
  cursor: none;
}
.btn-active {
  background-color: $success-color;
  pointer-events: all;
}
.category-name {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.upperText {
  text-transform: uppercase;
}
.crunchy-postgres {
  margin-top: 16px;
}
.mt8 {
  margin-top: 8px;
}
.mb8 {
  margin-bottom: 8px;
}
.mb24 {
  margin-bottom: 24px;
}
.mt16 {
  margin-top: 16px;
}
.mt24 {
  margin-top: 24px;
}
.mt32 {
  margin-top: 32px;
}
.mb0 {
  margin-bottom: 0;
}
.mb32 {
  margin-bottom: 32px;
}
.rc-form-group-horz {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: stretch;
}
.separator {
  padding-top: 40px;
}
.env-add-lnk p {
  color: $blue-main;
  padding-left: 8px;
}
.env-add-lnk:hover {
  color: $blue-main;
}
.env-name {
  flex: 1;
  font-weight: 700;
}

.env-value {
  flex: 1;
  margin-left: -40px;
  font-weight: 700;
}
</style>
