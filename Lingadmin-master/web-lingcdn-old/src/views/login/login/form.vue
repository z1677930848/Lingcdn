<template>
  <n-form
    ref="formRef"
    label-placement="left"
    size="large"
    :model="mode === 'account' ? formInline : formMobile"
    :rules="mode === 'account' ? rules : mobileRules"
  >
    <template v-if="mode === 'account'">
      <n-form-item path="username">
        <n-input
          @keyup.enter="debounceHandleSubmit"
          v-model:value="formInline.username"
          placeholder="请输入用户名"
        >
          <template #prefix>
            <n-icon size="18" color="#808695">
              <PersonOutline />
            </n-icon>
          </template>
        </n-input>
      </n-form-item>
      <n-form-item path="pass">
        <n-input
          @keyup.enter="debounceHandleSubmit"
          v-model:value="formInline.pass"
          type="password"
          show-password-on="click"
          placeholder="请输入密码"
        >
          <template #prefix>
            <n-icon size="18" color="#808695">
              <LockClosedOutline />
            </n-icon>
          </template>
        </n-input>
      </n-form-item>

      <n-form-item path="code" v-show="codeBase64 !== ''">
        <n-input-group>
          <n-input
            :style="{ width: '100%' }"
            placeholder="请输入验证码"
            @keyup.enter="debounceHandleSubmit"
            v-model:value="formInline.code"
          >
            <template #prefix>
              <n-icon size="18" color="#808695" :component="SafetyCertificateOutlined" />
            </template>
          </n-input>

          <n-loading-bar-provider :to="loadingBarTargetRef" container-style="position: absolute;">
            <img
              ref="loadingBarTargetRef"
              style="width: 100px"
              :src="codeBase64"
              @click="refreshCode"
              loading="lazy"
              alt="点击获取验证码"
            />
            <loading-bar-trigger />
          </n-loading-bar-provider>
        </n-input-group>
      </n-form-item>
    </template>

    <template v-if="mode === 'mobile'">
      <n-form-item path="mobile">
        <n-input
          @keyup.enter="handleMobileSubmit"
          v-model:value="formMobile.mobile"
          placeholder="请输入手机号"
        >
          <template #prefix>
            <n-icon size="18" color="#808695">
              <MobileOutlined />
            </n-icon>
          </template>
        </n-input>
      </n-form-item>

      <n-form-item path="code">
        <n-input-group>
          <n-input
            @keyup.enter="handleMobileSubmit"
            v-model:value="formMobile.code"
            placeholder="请输入验证码"
          >
            <template #prefix>
              <n-icon size="18" color="#808695" :component="SafetyCertificateOutlined" />
            </template>
          </n-input>
          <n-button
            type="primary"
            ghost
            @click="sendMobileCode"
            :disabled="isCounting"
            :loading="sendLoading"
          >
            {{ sendLabel }}
          </n-button>
        </n-input-group>
      </n-form-item>
    </template>

    <n-space :vertical="true" :size="24">
      <div class="flex-y-center justify-between">
        <n-checkbox v-model:checked="autoLogin">自动登录</n-checkbox>
        <n-button :text="true" @click="handleResetPassword">忘记密码？</n-button>
      </div>
      <n-button type="primary" size="large" :block="true" :loading="loading" @click="handleLogin">
        登录
      </n-button>

      <FormOther moduleKey="register" tag="注册账号" @updateActiveModule="updateActiveModule" />
    </n-space>

    <DemoAccount @login="handleDemoAccountLogin" />
  </n-form>
</template>

<script lang="ts" setup>
  import '../components/style.less';
  import { ref, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useUserStore } from '@/store/modules/user';
  import { useMessage, useLoadingBar } from 'naive-ui';
  import { ResultEnum } from '@/enums/httpEnum';
  import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5';
  import { PageEnum } from '@/enums/pageEnum';
  import { SafetyCertificateOutlined, MobileOutlined } from '@vicons/antd';
  import { GetCaptcha } from '@/api/base';
  import { aesEcb } from '@/utils/encrypt';
  import DemoAccount from './demo-account.vue';
  import FormOther from '../components/form-other.vue';
  import { useSendCode } from '@/hooks/common';
  import { SendSms } from '@/api/system/user';
  import { validate } from '@/utils/validateUtil';
  import { useDebounceFn } from '@vueuse/core';

  interface Props {
    mode: string;
  }

  const props = withDefaults(defineProps<Props>(), {
    mode: 'account',
  });

  interface FormState {
    username: string;
    pass: string;
    cid: string;
    code: string;
    password: string;
    redirect?: string;
  }

  interface FormMobileState {
    mobile: string;
    code: string;
  }

  const formRef = ref();
  const message = useMessage();
  const loading = ref(false);
  const autoLogin = ref(true);
  const codeBase64 = ref('');
  const loadingBar = useLoadingBar();
  const loadingBarTargetRef = ref<undefined | HTMLElement>(undefined);
  const userStore = useUserStore();
  const router = useRouter();
  const route = useRoute();
  const { sendLabel, isCounting, loading: sendLoading, activateSend } = useSendCode();
  const emit = defineEmits(['updateActiveModule']);
  const LOGIN_NAME = PageEnum.BASE_LOGIN_NAME;
  const debounceHandleSubmit = useDebounceFn((e) => {
    handleSubmit(e);
  }, 500);
  const formInline = ref<FormState>({
    username: '',
    pass: '',
    cid: '',
    code: '',
    password: '',
    redirect: '',
  });

  const formMobile = ref<FormMobileState>({
    mobile: '',
    code: '',
  });

  const rules = {
    username: { required: true, message: '请输入用户名', trigger: 'blur' },
    pass: { required: true, message: '请输入密码', trigger: 'blur' },
  };

  const mobileRules = {
    mobile: { required: true, message: '请输入手机号', trigger: 'blur' },
    code: { required: true, message: '请输入验证码', trigger: 'blur' },
  };

  const handleResetPassword = () => {
    message.warning('请联系管理员重置密码');
  };

  function updateActiveModule(key: string) {
    emit('updateActiveModule', key);
  }

  async function refreshCode() {
    const res = await GetCaptcha();
    codeBase64.value = 'data:image/png;base64,' + res.data?.b64s;
    formInline.value.cid = res.data?.cid || '';
  }

  function handleSubmit(e: Event) {
    e?.preventDefault();
    handleLogin();
  }

  function handleMobileSubmit(e: Event) {
    e?.preventDefault();
    handleLogin();
  }

  function handleDemoAccountLogin({ username, password }) {
    formInline.value.username = username;
    formInline.value.pass = password;
    handleLogin();
  }

  async function handleLogin() {
    loading.value = true;
    if (props.mode === 'account') {
      formInline.value.password = aesEcb(formInline.value.pass);
    }
    const data = props.mode === 'account' ? formInline.value : formMobile.value;
    const valid = await validate(formRef.value);
    if (!valid) {
      loading.value = false;
      return;
    }
    const res =
      props.mode === 'account'
        ? await userStore.login(data)
        : await userStore.mobileLogin(data);
    const { code } = res;
    if (code === ResultEnum.SUCCESS || code === 200) {
      const to = decodeURIComponent((route.query?.redirect || '') as string);
      message.success('登录成功');
      router.replace(to && to !== LOGIN_NAME ? to : PageEnum.BASE_HOME);
    } else {
      refreshCode();
    }
    loading.value = false;
  }

  async function sendMobileCode() {
    const mobile = formMobile.value.mobile;
    if (!mobile) {
      message.warning('请先输入手机号');
      return;
    }
    const valid = await validate(formRef.value, ['mobile']);
    if (!valid) return;
    activateSend();
    const params = {
      mobile,
      event: 'mobile_login',
    };
    await SendSms(params);
  }

  onMounted(async () => {
    const params = route.query?.redirect as string;
    if (params) formInline.value.redirect = params;
    await refreshCode();
  });
</script>
