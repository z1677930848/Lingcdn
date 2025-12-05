<script setup lang="ts">
  import { computed } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, NTag } from 'naive-ui';
  import { CountTo } from '@/components/CountTo';
  import { useUserStore } from '@/store/modules/user';
  import { formatToDateTime } from '@/utils/dateUtil';
  import {
    GlobeOutline,
    KeyOutline,
    LogInOutline,
    MedalOutline,
    PersonCircleOutline,
    ShieldCheckmarkOutline,
    StatsChartOutline,
    TimeOutline,
    WalletOutline,
  } from '@vicons/ionicons5';

  const router = useRouter();
  const message = useMessage();
  const userStore = useUserStore();
  const userInfo = computed(() => userStore.getUserInfo || userStore.info);

  const greeting = computed(
    () => userInfo.value?.realName || userInfo.value?.username || '用户'
  );

  const lastLoginAt = computed(() =>
    formatToDateTime(userInfo.value?.lastLoginAt ?? '') || '暂无记录'
  );

  const statsCards = computed(() => [
    {
      title: '账户余额',
      value: userInfo.value?.balance ?? 0,
      suffix: '元',
      icon: WalletOutline,
      color: '#18a058',
    },
    {
      title: '可用积分',
      value: userInfo.value?.integral ?? 0,
      suffix: '分',
      icon: MedalOutline,
      color: '#2080f0',
    },
    {
      title: '累计登录',
      value: userInfo.value?.loginCount ?? 0,
      suffix: '次',
      icon: LogInOutline,
      color: '#f0a020',
    },
    {
      title: '上次登录',
      text: lastLoginAt.value,
      icon: TimeOutline,
      color: '#8a63d2',
    },
  ]);

  const quickEntries = [
    {
      title: '域名管理',
      desc: '查看与接入站点',
      to: '/apply/attachment',
      icon: GlobeOutline,
      color: '#2080f0',
    },
    {
      title: '流量统计',
      desc: '命中率与带宽概览',
      to: '/monitor/serve-log',
      icon: StatsChartOutline,
      color: '#18a058',
    },
    {
      title: '证书中心',
      desc: '上传/续期证书',
      to: '/develop/addons',
      icon: ShieldCheckmarkOutline,
      color: '#f0a020',
    },
    {
      title: '个人设置',
      desc: '资料、安全与通知',
      to: '/home/account',
      icon: PersonCircleOutline,
      color: '#8a63d2',
    },
  ];

  const securityList = computed(() => [
    { label: '手机号', ok: !!userInfo.value?.mobile },
    { label: '邮箱', ok: !!userInfo.value?.email },
    { label: '邀请人', ok: !!userInfo.value?.inviteCode },
    { label: '二次验证', ok: false },
  ]);

  const goPage = (path: string) => {
    const matched = router.getRoutes().some((route) => route.path === path);
    if (matched) {
      router.push({ path });
    } else {
      message.warning('当前账号暂无对应页面权限');
    }
  };
</script>

<template>
  <div class="user-console">
    <n-grid cols="1 m:3" x-gap="16" y-gap="16" responsive="screen">
      <n-grid-item :span="2">
        <n-card class="welcome-card" :bordered="false">
          <div class="welcome-head">
            <div class="welcome-title">你好，{{ greeting }}</div>
            <n-tag type="primary" size="small" round>
              {{ userInfo?.deptName || '未分组部门' }}
            </n-tag>
          </div>
          <div class="welcome-meta">
            <div class="meta-item">
              <span class="meta-label">角色</span>
              <span class="meta-value">{{ userInfo?.roleName || '普通用户' }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">上次登录</span>
              <span class="meta-value">
                {{ lastLoginAt }}
                <n-tag v-if="userInfo?.lastLoginIp" size="small" class="ip-tag">
                  {{ userInfo?.lastLoginIp }}
                </n-tag>
              </span>
            </div>
            <div class="meta-item">
              <span class="meta-label">当前城市</span>
              <span class="meta-value">{{ userInfo?.cityLabel || '未填写' }}</span>
            </div>
          </div>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card class="shortcut-card" :bordered="false">
          <div class="shortcut-header">
            <div>
              <div class="shortcut-title">快速开始</div>
              <div class="shortcut-desc">常用入口与操作</div>
            </div>
            <n-tag type="success" size="small" round>
              {{ userInfo?.username }}
            </n-tag>
          </div>
          <div class="shortcut-actions">
            <n-button
              v-for="item in quickEntries"
              :key="item.title"
              :color="item.color"
              quaternary
              size="small"
              @click="goPage(item.to)"
            >
              <template #icon>
                <n-icon :component="item.icon" />
              </template>
              {{ item.title }}
            </n-button>
          </div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <n-grid cols="1 m:2 xl:4" x-gap="16" y-gap="16" responsive="screen" class="mt-4">
      <n-grid-item v-for="card in statsCards" :key="card.title">
        <n-card :bordered="false" size="small" class="stat-card">
          <div class="stat-icon" :style="{ color: card.color }">
            <n-icon :component="card.icon" size="26" />
          </div>
          <div class="stat-body">
            <div class="stat-title">{{ card.title }}</div>
            <div class="stat-value">
              <template v-if="card.text">
                {{ card.text }}
              </template>
              <template v-else>
                <CountTo
                  :startVal="0"
                  :endVal="card.value"
                  :duration="1.2"
                  class="value-number"
                />
                <span class="value-suffix">{{ card.suffix }}</span>
              </template>
            </div>
          </div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <n-grid cols="1 m:2" x-gap="16" y-gap="16" responsive="screen" class="mt-4">
      <n-grid-item>
        <n-card title="账号概览" :bordered="false" size="small" segmented>
          <n-descriptions label-placement="left" column="1" :label-style="{ width: '90px' }">
            <n-descriptions-item label="用户名">
              {{ userInfo?.username || '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="手机号">
              <n-tag v-if="userInfo?.mobile" type="success" size="small" round>
                {{ userInfo?.mobile }}
              </n-tag>
              <span v-else>未绑定</span>
            </n-descriptions-item>
            <n-descriptions-item label="邮箱">
              <n-tag v-if="userInfo?.email" type="success" size="small" round>
                {{ userInfo?.email }}
              </n-tag>
              <span v-else>未绑定</span>
            </n-descriptions-item>
            <n-descriptions-item label="注册地址">
              {{ userInfo?.address || '未填写' }}
            </n-descriptions-item>
            <n-descriptions-item label="注册时间">
              {{ formatToDateTime(userInfo?.createdAt || '') || '—' }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card title="安全状态" :bordered="false" size="small" segmented>
          <div class="security-list">
            <div v-for="item in securityList" :key="item.label" class="security-item">
              <div class="security-label">
                <n-icon :component="KeyOutline" size="18" class="security-icon" />
                {{ item.label }}
              </div>
              <NTag :type="item.ok ? 'success' : 'error'" size="small" round>
                {{ item.ok ? '已开启' : '未开启' }}
              </NTag>
            </div>
          </div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <n-grid cols="1 m:2" x-gap="16" y-gap="16" responsive="screen" class="mt-4">
      <n-grid-item>
        <n-card title="最近动态" :bordered="false" size="small" segmented>
          <n-timeline size="large">
            <n-timeline-item type="success" title="登录成功" :time="lastLoginAt">
              设备 IP：{{ userInfo?.lastLoginIp || '未知' }}
            </n-timeline-item>
            <n-timeline-item type="info" title="账号创建" :time="formatToDateTime(userInfo?.createdAt || '')">
              通过 {{ userInfo?.inviteCode ? '邀请码 ' + userInfo?.inviteCode : '自助注册' }}
            </n-timeline-item>
            <n-timeline-item type="warning" title="个人资料完善">
              如需修改头像、密码可前往「个人设置」
            </n-timeline-item>
          </n-timeline>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card title="常用入口" :bordered="false" size="small" segmented>
          <n-grid cols="1 s:2" x-gap="12" y-gap="12" responsive="screen">
            <n-grid-item v-for="item in quickEntries" :key="item.title">
              <div class="entry-card" @click="goPage(item.to)">
                <div class="entry-icon" :style="{ color: item.color }">
                  <n-icon :component="item.icon" size="22" />
                </div>
                <div class="entry-body">
                  <div class="entry-title">{{ item.title }}</div>
                  <div class="entry-desc">{{ item.desc }}</div>
                </div>
              </div>
            </n-grid-item>
          </n-grid>
        </n-card>
      </n-grid-item>
    </n-grid>
  </div>
</template>

<style scoped lang="less">
.user-console {
  .welcome-card {
    background: linear-gradient(135deg, #f3f8ff 0%, #f7f3ff 100%);
    box-shadow: 0 12px 40px rgba(64, 106, 255, 0.1);
  }

  .welcome-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .welcome-title {
    font-size: 18px;
    font-weight: 700;
    color: #111827;
  }

  .welcome-meta {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 10px;
  }

  .meta-item {
    display: flex;
    gap: 8px;
    align-items: center;
    color: #475569;
    font-size: 13px;
  }

  .meta-label {
    color: #6b7280;
    min-width: 68px;
  }

  .ip-tag {
    margin-left: 6px;
  }

  .shortcut-card {
    height: 100%;
  }

  .shortcut-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .shortcut-title {
    font-weight: 700;
    font-size: 16px;
  }

  .shortcut-desc {
    color: #6b7280;
    font-size: 13px;
  }

  .shortcut-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .stat-card {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .stat-icon {
    width: 44px;
    height: 44px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: #f8fafc;
    border-radius: 14px;
  }

  .stat-body {
    flex: 1;
  }

  .stat-title {
    color: #6b7280;
    font-size: 13px;
    margin-bottom: 4px;
  }

  .stat-value {
    font-weight: 700;
    font-size: 18px;
    display: flex;
    align-items: baseline;
    gap: 6px;
    color: #111827;
  }

  .value-number {
    line-height: 1;
  }

  .value-suffix {
    color: #6b7280;
    font-size: 13px;
  }

  .security-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 12px;
  }

  .security-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    border: 1px solid #eef2f7;
    border-radius: 12px;
    background: #fbfcff;
  }

  .security-label {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #1f2937;
    font-weight: 600;
  }

  .security-icon {
    color: #2080f0;
  }

  .entry-card {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 12px;
    border-radius: 12px;
    border: 1px solid #eef2f7;
    background: #fff;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      box-shadow: 0 12px 28px rgba(0, 0, 0, 0.06);
      transform: translateY(-2px);
    }
  }

  .entry-icon {
    width: 38px;
    height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: #f8fafc;
    border-radius: 10px;
  }

  .entry-body {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .entry-title {
    font-weight: 700;
    color: #111827;
  }

  .entry-desc {
    color: #6b7280;
    font-size: 12px;
  }
}
</style>
