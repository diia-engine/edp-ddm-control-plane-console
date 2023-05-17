<script lang="ts">
interface WizardTemplateVariables {
  registryValues: any;
}
import {defineComponent, type PropType} from 'vue';
export default defineComponent({
  props: {
    templatePreloadedData: Object,
    formSubmitted: Boolean,
    templateVariables: {
      required: true,
      type: Object as PropType<WizardTemplateVariables>,
    },
  },
  data() {
    return {
      registryResources: {
        encoded: "",
        cat: "",
        cats: [
          "kong",
          "bpms",
          "digitalSignatureOps",
          "userTaskManagement",
          "userProcessManagement",
          "digitalDocumentService",
          "restApi",
          "kafkaApi",
          "soapApi",
        ],
        addedCats: [] as Array<any>,
      },
      crunchyPostgres: {
        maxConnections: '',
        storageSize: '',
      },
    };
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
    removeResourcesCatFromList(name: string) {
      const searchIdx = this.registryResources.cats.indexOf(name);
      if (searchIdx !== -1) {
        this.registryResources.cats.splice(searchIdx, 1);
      }
    },
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
      envVars.push({ name: "", value: "" });
    },
    removeEnvVar(envVars: Array<Record<string, unknown>>, env: Record<string, unknown>) {
      envVars.splice(envVars.indexOf(env), 1);
    },
    removeResourceCat(cat: { name: string }, event: any) {
      event.preventDefault();
      this.registryResources.cats.push(cat.name);
      this.registryResources.addedCats.splice(this.registryResources.addedCats.indexOf(cat), 1);
    },
    encodeRegistryResources() {
      const prepare = {} as Record<string, unknown>;
      this.registryResources.addedCats.forEach(function (el: any) {
        const cloneEL = JSON.parse(JSON.stringify(el));
        const envVars = {} as Record<string, unknown>;
        cloneEL.config.container.envVars.forEach(function (el: any) {
          envVars[el.name] = el.value;
        });
        cloneEL.config.container.envVars = envVars;
        prepare[cloneEL.name] = {
          istio: cloneEL.config.istio,
          container: cloneEL.config.container,
        };
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
          }
          else if (obj[key] === "") {
            delete obj[key];
          }
        }
      }
    },
    mergeResource(data: Record<string, unknown>) {
      const emptyResource = {
        istio: {
          sidecar: {
            enabled: false,
            resources: {
              requests: {
                cpu: "",
                memory: ""
              },
              limits: {
                cpu: "",
                memory: "",
              },
            },
          },
        },
        container: {
          resources: {
            requests: {
              cpu: "",
              memory: ""
            },
            limits: {
              cpu: "",
              memory: "",
            },
          },
          envVars: [{ name: "", value: "" }],
        },
      };
      this.mergeDeep(emptyResource, data);
      return emptyResource;
    },
    addResourceCat() {
      if (this.registryResources.cat === "") {
        return;
      }
      this.registryResources.addedCats.unshift({
        name: this.registryResources.cat,
        config: {
          istio: {
            sidecar: {
              enabled: false,
              resources: {
                requests: {
                  cpu: "",
                  memory: ""
                },
                limits: {
                  cpu: "",
                  memory: "",
                },
              },
            },
          },
          container: {
            resources: {
              requests: {
                cpu: "",
                memory: ""
              },
              limits: {
                cpu: "",
                memory: "",
              },
            },
            envVars: [{ name: "", value: "" }],
          },
        }
      });
      this.registryResources.cats.splice(this.registryResources.cats.indexOf(this.registryResources.cat), 1);
    },
    preloadRegistryResources(values: any) {
      const crunchyPostgres = values?.global?.crunchyPostgres;
      if (crunchyPostgres) {
        this.crunchyPostgres.maxConnections = crunchyPostgres.postgresql?.parameters?.max_connections;
        this.crunchyPostgres.storageSize = crunchyPostgres.storageSize;
      }

      const data = values?.global?.registry;
      if (!data) {
        return;
      }

      this.registryResources.cats = [
        "kong",
        "bpms",
        "digitalSignatureOps",
        "userTaskManagement",
        "userProcessManagement",
        "digitalDocumentService",
        "restApi",
        "kafkaApi",
        "soapApi",
      ];
      this.registryResources.addedCats = [];
      for (const i in data) {
        this.removeResourcesCatFromList(i);
        if ("container" in data[i] &&
            this.isObject(data[i].container) &&  "envVars" in data[i].container) {
          data[i].container.envVars = this.decodeResourcesEnvVars(data[i].container.envVars);
        }
        const mergedData = this.mergeResource(data[i]);
        this.registryResources.addedCats.push({
          name: i,
          config: mergedData,
        });
      }
    },
    isObject(item: unknown) {
      return (item && typeof item === "object" && !Array.isArray(item));
    },
    mergeDeep(target: any, ...sources: any[]): any {
      if (!sources.length)
        return target;
      const source = sources.shift();
      if (this.isObject(target) && this.isObject(source)) {
        for (const key in source) {
          if (source[key] === null) {
            continue;
          }
          if (this.isObject(source[key])) {
            if (!target[key])
              Object.assign(target, { [key]: {} });
            this.mergeDeep(target[key], source[key]);
          }
          else {
            Object.assign(target, { [key]: source[key] });
          }
        }
      }
      return this.mergeDeep(target, ...sources);
    },
  },
  mounted() {
    this.preloadRegistryResources(this.templateVariables.registryValues);
  },
});
</script>

<style scoped>
  .crunchy-postgres {
    margin-top: 16px;
  }
</style>

<template>
  <h2>Ресурси реєстру</h2>
  <input type="hidden" name="resources" :value="registryResources.encoded" />

  <div class="registry-resources">
      <div class="rc-form-group res-cat-select">
          <select v-model="registryResources.cat">
              <option v-for="cat in registryResources.cats" v-bind:key="cat">{{ cat }}</option>
          </select>
          <button @click.prevent="addResourceCat">+</button>
      </div>

      <div class="cat-line" v-for="cat in registryResources.addedCats" v-bind:key="cat.name">
          <h4>{{ cat.name }}
              <button type="button" @click="removeResourceCat(cat, $event)" class="remove-cat">-</button></h4>
          <div class="rc-form-group">
              <h5>Istio sidecar</h5>
              <div class="sidecar-enabled">
                  <input v-model="cat.config.istio.sidecar.enabled" type="checkbox" id="istio-sidecar-enabled">
                  <label for="istio-sidecar-enabled">Enabled</label>
              </div>
              <label>Requests</label>
              <input v-model="cat.config.istio.sidecar.resources.requests.cpu" type="text" placeholder="CPU" />
              <input v-model="cat.config.istio.sidecar.resources.requests.memory" type="text" placeholder="Memory" />

              <label>Limits</label>
              <input v-model="cat.config.istio.sidecar.resources.limits.cpu" type="text" placeholder="CPU" />
              <input v-model="cat.config.istio.sidecar.resources.limits.memory" type="text" placeholder="Memory" />
          </div>

          <div class="rc-form-group">
              <h5>Container</h5>
              <label>Requests</label>
              <input v-model="cat.config.container.resources.requests.cpu" type="text" placeholder="CPU" />
              <input v-model="cat.config.container.resources.requests.memory" type="text" placeholder="Memory" />

              <label>Limits</label>
              <input v-model="cat.config.container.resources.limits.cpu" type="text" placeholder="CPU" />
              <input v-model="cat.config.container.resources.limits.memory" type="text" placeholder="Memory" />

              <label>Змінні оточення</label>
              <div class="env-vars">
                  <div class="env-var-line" v-for="env in cat.config.container.envVars" v-bind:key="env.value" >
                      <input class="env-name" type="text" placeholder="Name" v-model="env.name" />
                      <input class="env-value" type="text" placeholder="Value" v-model="env.value" />
                      <button @click="removeEnvVar(cat.config.container.envVars, env)" class="remove-env-var">-</button>
                  </div>
                  <a class="env-add-lnk" @click="addEnvVar(cat.config.container.envVars, $event)" href="#">Додати змінну оточення</a>
              </div>
          </div>
      </div>


      <div class="rc-form-group crunchy-postgres">
        <h5>Crunchy Postgres</h5>
        <label>Max Connections</label>
        <input v-model="crunchyPostgres.maxConnections" type="text" name="crunchy-postgres-max-connections" />

        <label>Storage Size</label>
        <input v-model="crunchyPostgres.storageSize" type="text" name="crunchy-postgres-storage-size" />
      </div>
  </div>
</template>
