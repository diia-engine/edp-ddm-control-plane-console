<script setup lang="ts">
  import { inject } from 'vue';
  interface DashboardTemplateVariables {
      gerritLink: string;
      jenkinsLink: string;
  }
  const variables = inject('TEMPLATE_VARIABLES') as DashboardTemplateVariables;
  const gerritLink = variables?.gerritLink;
  const jenkinsLink = variables?.jenkinsLink;
</script>

<template>
  <div class="dashboard-panel">
      <ul>
          <li>
              <img src="@/assets/img/gerrit-logo.png" alt="gerrit logo" />
              <div class="dashboard-item-content">
                  <a v-bind:href="gerritLink">Gerrit <img src="@/assets/img/action-link.png" alt="gerrit link"></a>
                  <p>
                      Gerrit - вільне програмне забезпечення з веб-інтерфейсом, інтегроване з розподіленою
                      системою контролю версій Git, призначене для спільного проведення інспекції вихідного коду.
                      Gerrit був розроблений співробітником корпорації Google Шоном Пірсом для розробки операційної
                      системи Android.
                  </p>
              </div>
          </li>
          <li>
              <img src="@/assets/img/jenkins-logo.png" alt="jenkins logo" />
              <div class="dashboard-item-content">
                  <a  v-bind:href="jenkinsLink">Jenkins <img src="@/assets/img/action-link.png" alt="gerrit link"></a>
                  <p>
                      Jenkins - програмна система з відкритим вихідним кодом на Java, призначена для забезпечення
                      процесу безперервної інтеграції програмного забезпечення. Відгалужуючись в 2008 році від
                      проекту Hudson, що належить компанії Oracle, основним його автором - Косуке Кавагуті.
                      Поширюється під ліцензією MIT.
                  </p>
              </div>
          </li>
      </ul>
  </div>
</template>
