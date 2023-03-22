import { createRouter, createWebHistory } from 'vue-router';
import DashboardView from '../views/DashboardView.vue';
import UpdateRegistry from '../views/registry/UpdateView.vue';
import RegistryView from '../views/registry/RegistryView.vue';
import EditCluster from '../views/cluster/EditView.vue';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/admin/dashboard',
      name: 'dashboard',
      component: DashboardView
    },
    {
      path: '/admin/registry/update/:registryName',
      name: 'updateRegistry',
      component: UpdateRegistry
    },
    {
      path: '/admin/registry/view/:registryName',
      name: 'registry',
      component: RegistryView
    },
    {
      path: '/admin/cluster/edit',
      name: 'editCluster',
      component: EditCluster
    },
  ]
});

export default router;
