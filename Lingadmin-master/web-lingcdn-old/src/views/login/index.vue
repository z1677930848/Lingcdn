<template>
  <div class="view-account">
    <div class="view-account-header"></div>
    <div :style="containerCSS">
      <n-card :bordered="false">
        <header class="justify-between">
          <n-space justify="center">
            <div></div>
            <img src="~@/assets/images/logo.png" class="account-logo" alt="" />
            <n-gradient-text type="primary" :size="26">{{ projectName }}</n-gradient-text>
            <div></div>
          </n-space>
        </header>
        <main class="pt-24px">
          <div class="pt-18px">
            <transition name="fade-slide" appear>
              <component
                :is="activeModule.component"
                @updateActiveModule="handleUpdateActiveModule"
              />
            </transition>
          </div>
        </main>
      </n-card>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import LoginFrom from './login/index.vue';
  import RegisterFrom from './register/index.vue';
  import { useRouter } from 'vue-router';
  import { useUserStore } from '@/store/modules/user';

  const userStore = useUserStore();
  const projectName = computed(() => userStore.loginConfig?.projectName || 'Lingadmin');

  interface LoginModule {
    key: string;
    label: string;
    component: Component;
  }

  const router = useRouter();
  const activeModule = ref<LoginModule>({
    key: 'login',
    label: '账号登录',
    component: LoginFrom,
  });

  const modules: LoginModule[] = [
    { key: 'login', label: '账号登录', component: LoginFrom },
    // { key: 'register', label: '注册账号', component: RegisterFrom },
  ];

  const containerCSS = computed(() => {
    const val = document.body.clientWidth;
    return val <= 720
      ? {}
      : {
          flex: `1`,
          padding: `62px 12px`,
          'max-width': `484px`,
          'min-width': '320px',
          margin: '0 auto',
        };
  });

  function handleUpdateActiveModule(key: string) {
    const findItem = modules.find((item) => item.key === key);
    if (findItem) {
      activeModule.value = findItem;
    }
  }

  onMounted(() => {
    // 是否开放注册
    if (userStore.loginConfig?.loginRegisterSwitch === 1) {
      const findItem = modules.find((item) => item.key === 'register');
      if (!findItem) {
        modules.push({ key: 'register', label: '注册账号', component: RegisterFrom });
      }
    }

    const key = router.currentRoute.value.query?.scope as string;
    if (key) {
      handleUpdateActiveModule(key);
    }
  });
</script>

<style lang="less" scoped>
  .view-account {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: auto;

    &-top {
      padding: 32px 0;
      text-align: center;

      &-desc {
        font-size: 14px;
        color: #808695;
      }
    }

    &-other {
      width: 100%;
    }

    .default-color {
      color: #515a6e;

      .ant-checkbox-wrapper {
        color: #515a6e;
      }
    }

    &-header {
      width: 100%;
      height: 36vh;
      min-height: 260px;
      background-repeat: no-repeat;
      background-size: cover;
      background-color: #f7f9fb;
      background-image: url('../../assets/images/login.svg');
    }
  }

  .account-logo {
    height: 80px;
    width: auto;
  }
</style>
