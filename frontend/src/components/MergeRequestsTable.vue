<script setup lang="ts">
import { toRefs } from 'vue';

interface MRTableProps {
  mergeRequests: any;
  inPlatform?: boolean;
  mrAvailable?: string;
  createReleaseAvailable?: boolean;
}

const props = defineProps<MRTableProps>();
const { inPlatform, mergeRequests, mrAvailable, createReleaseAvailable } = toRefs(props);
</script>

<script lang="ts">
import $ from 'jquery';
import 'datatables.net-dt';
import {
  getDateTimestamp,
  getFormattedDate,
  getMergeRequestAction,
  getMergeRequestPlatformAction,
  getMergeRequestName,
  getMergeRequestStatus,
  mrIsInProgress,
  isRegistryUpdateMrOpen,
} from '@/utils';

export default {
  methods: {
    onViewClick(url: string) {
      this.$emit('onViewClick', url);
    },
    mrRefreshClick() {
      window.localStorage.setItem("mr-scroll", "true");
    },
    isRegistryMrAvailable(mergeRequest: any, isUpdate: boolean) {
      if (this.mrAvailable || mergeRequest.status.value !== 'NEW') {
        return isUpdate ? this.createReleaseAvailable : true;
      }
      return false;
    },
    isRegistryPipeInProgress(isUpdate: boolean) {
      return !(isUpdate ? this.mrAvailable && this.createReleaseAvailable : this.mrAvailable);
    }
  },
  mounted() {
    $("#mr-table").DataTable({
        ordering: true,
        paging: true,
        columnDefs: [
            { orderable: false, targets: 4 },
        ],
        order: [[0, 'desc']],
        language: {
            "processing": this.$t('components.mergeRequestsTable.table.wait'),
            "lengthMenu": this.$t('components.mergeRequestsTable.table.lengthMenu'),
            "zeroRecords": this.$t('components.mergeRequestsTable.table.zeroRecords'),
            "info": this.$t('components.mergeRequestsTable.table.info'),
            "infoEmpty": this.$t('components.mergeRequestsTable.table.infoEmpty'),
            "infoFiltered": this.$t('components.mergeRequestsTable.table.infoFiltered'),
            "search": this.$t('components.mergeRequestsTable.table.search'),
            "paginate": {
                "first": this.$t('components.mergeRequestsTable.table.first'),
                "previous": this.$t('components.mergeRequestsTable.table.previous'),
                "next": this.$t('components.mergeRequestsTable.table.next'),
                "last": this.$t('components.mergeRequestsTable.table.last')
            },
            "aria": {
                "sortAscending": this.$t('components.mergeRequestsTable.table.sortAscending'),
                "sortDescending": this.$t('components.mergeRequestsTable.table.sortDescending')
            }
        }
    });
  }
};
</script>

<style scoped>
.mr-refresh {
  margin-left: 5px;
}
</style>

<template>
  <table class="rg-info-table rg-info-table-config" id="mr-table">
      <thead>
          <tr>
              <th>{{ $t('components.mergeRequestsTable.table.date') }}</th>
              <th>{{ $t('components.mergeRequestsTable.table.request') }}</th>
              <th>{{ $t('components.mergeRequestsTable.table.operation') }}</th>
              <th>{{ $t('components.mergeRequestsTable.table.status') }}</th>
              <th></th>
          </tr>
      </thead>
      <tbody>
          <tr v-for="($al, $index) in mergeRequests" :key="$index">
              <td :data-order="getDateTimestamp($al.metadata.creationTimestamp)">{{getFormattedDate($al.metadata.creationTimestamp)}}</td>
              <td>{{ inPlatform ? $al.metadata.name : getMergeRequestName($al) }}</td>
              <td>{{ inPlatform ? getMergeRequestPlatformAction($al) : getMergeRequestAction($al) }}</td>
              <td class="mr-status">
                {{ getMergeRequestStatus($al) }}
                <a :title="$t('actions.refresh')" class="mr-refresh" href="" @click="mrRefreshClick" v-if="mrIsInProgress($al) || ((isRegistryUpdateMrOpen($al) && !createReleaseAvailable))"><i class="fa-solid fa-arrow-rotate-right"></i></a>
              </td>
              <td class="mr-actions">
                  <i v-if="!inPlatform && $al.status.value === 'NEW' && isRegistryPipeInProgress(isRegistryUpdateMrOpen($al))" :title="$t('components.mergeRequestsTable.table.registerUpdated')" class="fa-solid fa-lock"></i>

                  <span v-if="$al.status.changeUrl && (inPlatform || isRegistryMrAvailable($al, isRegistryUpdateMrOpen($al)))">
                      <a :title="$t('components.mergeRequestsTable.actions.review')"
                          @click.stop.prevent="onViewClick(`/admin/change/${$al.status.changeId}`)"
                          :href="`/admin/change/${$al.status.changeId}`">
                          <i class="fa-solid fa-eye fa-lg"></i>
                      </a>

                      <a :href="$al.status.changeUrl" target="_blank">
                          <img style="vertical-align: sub;" :title="$t('components.mergeRequestsTable.actions.reviewInGerrit')" alt="vcs"
                              src="@/assets/img/action-link.png" />
                      </a>
                  </span>
              </td>
          </tr>
      </tbody>
  </table>
</template>
