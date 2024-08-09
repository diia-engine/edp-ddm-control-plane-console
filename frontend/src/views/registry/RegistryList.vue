<script lang="ts">
import { defineComponent, inject } from 'vue';
import type { RegistryTemplateVariables } from '@/types/registry';
import $ from 'jquery';
import 'datatables.net-dt';
import { getImageUrl, getFormattedDate, getStatusTitle } from '@/utils';
import Modal from '@/components/common/Modal.vue';
import Typography from '@/components/common/Typography.vue';
import '@/assets/datatables.custom.css';

export default defineComponent({
  setup() {
    const variables = inject('TEMPLATE_VARIABLES') as RegistryTemplateVariables;
    const allowedToCreate = variables?.allowedToCreate;
    const registries = variables?.registries;
    const page = variables?.page;
    const gerritBranches = variables?.gerritBranches;

    return {
      allowedToCreate,
      registries,
      page,
      gerritBranches,
      getImageUrl,
      getFormattedDate,
      getStatusTitle,
    };
  },
  data() {
    return {
      showModalCreateRegistry: false,
      versionTemplate: '1.9.7',
    };
  },
  components: {
    Modal,
    Typography,
  },
  methods: {
    getStatus(registry: any) {
      const { Codebase } = registry;
      let status = Codebase.status.value;
      if (status === '') {
        status = 'active';
      }

      const statusAnnotation = Codebase.metadata.annotations['console-status'];

      if (
        statusAnnotation &&
        (statusAnnotation == 'inactive-branches' ||
          statusAnnotation == 'running-jobs')
      ) {
        status = 'inactive';
      }

      return status;
    },
    isAvailable(registry: any) {
      const { Codebase } = registry;
      return (
        !Codebase.metadata.deletionTimestamp &&
        this.getStatus(registry) != 'failed'
      );
    },
    getUrl(registry: any, action: string) {
      let url = `/admin/registry/${action}/${registry.Codebase.metadata.name}`;
      if (registry.Codebase.version) {
        return `${url}?version=${this.getMajorVersion(
          registry.Codebase.version
        )}`;
      }
      return url;
    },
    getMajorVersion(fullVersion: string) {
      if (fullVersion === '') {
        return '';
      }

      let parts = fullVersion.split('-');
      return parts[0];
    },
    canBeDeleted(registry: any): boolean {
      const { Codebase } = registry;
      if (Codebase.branches && Codebase.branches.length) {
        const res = Codebase.branches.find(
          (b: any) => b.status.value !== 'active'
        );
        if (res) {
          return false;
        }
      }

      return Codebase.status.available && Codebase.status.value == 'active';
    },
    createRegistry() {
      window.location.href = `/admin/registry/create?version=${this.versionTemplate}`;
    },
    handleCreateRegistry() {
      const isExistBranch196 = this.gerritBranches?.some(b => b.includes('1.9.6'));
      if (isExistBranch196) {
        this.showModalCreateRegistry = true;
        return;
      }
      window.location.href = '/admin/registry/create';
    },
    getVersionDescription() {
      if (this.versionTemplate === '1.9.7') {
        return 'Актуальна версія. Містить останні затверджені зміни і нові функціональні можливості.';
      }

      return 'Попередня стабільна версія. Рекомендуємо обирати лише в разі обґрунтованої необхідності.';
    },
  },
  mounted() {
    $(function () {
      let registryName: any;
      let registryInput = $('#registry-name');
      let popupFooter = $('.popup-footer');

      let hidePopup = function (e: any) {
        $('.popup-backdrop').hide();
        $(e.target).closest('.popup-window').hide();
        registryInput.val('');
      };

      let showPopup = function (deletePopupName: any) {
        $('.popup-backdrop').show();
        $(deletePopupName).show();
        popupFooter.removeClass('active');
        registryInput.val('');
      };

      $('.hide-popup').click(function (e) {
        hidePopup(e);
        return false;
      });

      registryInput.val('');
      registryInput.keyup(function (e) {
        popupFooter.removeClass('active');
        if (registryName.toString() === $(e.currentTarget).val()) {
          popupFooter.addClass('active');
        }
      });

      $('.delete-registry').click(function (e) {
        registryName = $(e.currentTarget).data('name');
        $('#delete-name').html(registryName);

        showPopup('#delete-popup');
      });

      $('.no-delete-registry').click(function (e) {
        registryName = $(e.currentTarget).data('name');
        $('#no-delete-name').html(registryName);

        showPopup('#no-delete-popup');
      });

      $('#delete-form').submit(function () {
        return registryName.toString() === registryInput.val();
      });
    });

    $(document).ready(function () {
      $('#registry-table').DataTable({
        ordering: true,
        paging: true,
        columnDefs: [
          { orderable: false, targets: 0 },
          { orderable: false, targets: 5 },
          { orderable: false, targets: 6 },
          {
            targets: 4,
          },
        ],
        order: [[4, 'desc']],
        language: {
          processing: 'Зачекайте...',
          lengthMenu: 'Показати _MENU_ записів',
          zeroRecords: 'Записи відсутні.',
          info: 'Записи з _START_ по _END_ із _TOTAL_ записів',
          infoEmpty: 'Записи з 0 по 0 із 0 записів',
          infoFiltered: '(відфільтровано з _MAX_ записів)',
          search: 'Пошук:',
          paginate: {
            first: 'Перша',
            previous: 'Попередня',
            next: 'Наступна',
            last: 'Остання',
          },
          aria: {
            sortAscending: ': активувати для сортування стовпців за зростанням',
            sortDescending: ': активувати для сортування стовпців за спаданням',
          },
        },
      });
    });
  },
});
</script>

<template>
  <div class="registry" id="tooltip">
    <div class="registry-header">
      <h1>Реєстри</h1>
      <a
        href="#"
        class="registry-add"
        v-if="allowedToCreate"
        @click="handleCreateRegistry"
      >
        <img alt="add registry" src="@/assets/img/plus.png" />
        <span>Створити новий</span>
      </a>
    </div>
    <div class="registry-description">Перелік реєстрів та їх статусів.</div>
    <div class="registry-table-wrap">
      <table id="registry-table" class="registry-table row-border">
        <thead>
          <tr>
            <th>Статус</th>
            <th>Назва</th>
            <th>Версія</th>
            <th>Опис</th>
            <th>Час створення</th>
            <th></th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="($registry, $key) in registries" :key="$key">
            <td>
              <img
                v-if="$registry.Codebase.metadata.deletionTimestamp"
                title="Видалення"
                src="@/assets/img/action-delete.png"
                alt="delete registry"
              />

              <img
                v-else
                :title="getStatusTitle(getStatus($registry))"
                :src="getImageUrl(`status-${getStatus($registry)}`)"
                :alt="getStatusTitle(getStatus($registry))"
              />
            </td>
            <td>
              <a
                v-if="isAvailable($registry)"
                :href="getUrl($registry, 'view')"
                class="registry-name"
              >
                {{ $registry.Codebase.metadata.name }}
              </a>
              <template v-else>{{ $registry.Codebase.metadata.name }}</template>
            </td>
            <td>
              {{ $registry.Codebase.spec.defaultBranch }}
            </td>
            <td>
              {{ $registry.Codebase.spec.description }}
            </td>
            <td>
              {{
                getFormattedDate($registry.Codebase.metadata.creationTimestamp)
              }}
            </td>
            <td>
              <a
                v-if="$registry.CanUpdate && isAvailable($registry)"
                :href="getUrl($registry, 'edit')"
              >
                <img
                  title="Редагувати"
                  src="@/assets/img/action-edit.png"
                  alt="edit registry"
                />
              </a>
            </td>
            <td>
              <a
                v-if="
                  $registry.CanDelete &&
                  canBeDeleted($registry) &&
                  isAvailable($registry)
                "
                href="#"
                class="delete-registry"
                :data-name="$registry.Codebase.metadata.name"
              >
                <img
                  title="Видалити"
                  src="@/assets/img/action-delete.png"
                  alt="delete registry"
                />
              </a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <div class="popup-backdrop"></div>
  <div id="delete-popup" class="popup-window">
    <div class="popup-header">
      <p>Видалити "<span id="delete-name">business-pro</span>"?</p>
      <a href="#" class="popup-close hide-popup">
        <img alt="close popup window" src="@/assets/img/close.png" />
      </a>
    </div>
    <form id="delete-form" method="post" action="">
      <div class="popup-body">
        <p v-if="page === 'registry'">
          Щоб уникнути випадкової втрати даних, введіть назву реєстру перш, ніж
          ви зможете його видалити.
        </p>
        <p v-if="page === 'group'">
          Щоб уникнути випадкової втрати даних, введіть назву групи перш, ніж ви
          зможете її видалити.
        </p>

        <div class="rc-form-group">
          <input
            aria-label="registry name"
            type="text"
            id="registry-name"
            name="registry-name"
            required
            :placeholder="page === 'registry' ? 'Назва реєстру' : 'Назва групи'"
            autocomplete="off"
          />
        </div>
      </div>
      <div class="popup-footer">
        <a href="#" id="delete-cancel" class="hide-popup">відмінити</a>
        <button value="submit" name="codebase-delete" type="submit">
          Видалити
        </button>
      </div>
    </form>
  </div>
  <Modal
    title="Створити новий реєстр"
    submitBtnText="Підтвердити"
    :show="showModalCreateRegistry"
    @close="showModalCreateRegistry = false"
    @submit="createRegistry"
  >
    <div class="rc-form-group">
      <label>Оберіть версію</label>
      <select v-model="versionTemplate">
        <option>1.9.7</option>
        <option>1.9.6</option>
      </select>
    </div>
    <Typography variant="bodyText" class="mt24">
      {{ getVersionDescription() }}
    </Typography>
  </Modal>
</template>
<style lang="scss" scoped>
.registry-name {
  color: $blue-main;
  font-weight: 700;
}
.mt24 {
  margin-top: 24px;
}
</style>
