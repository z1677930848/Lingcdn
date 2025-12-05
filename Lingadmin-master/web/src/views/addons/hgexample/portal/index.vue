<template>
  <div>
    <n-card title="应用入口" class="proCard">
      <n-tabs type="line" animated>
        <n-tab-pane name="admin" tab="后台API">
          <Form path="/admin/hgexample/index/test?name=LingCDN" />
        </n-tab-pane>
        <n-tab-pane name="api" tab="前台API">
          <Form path="/api/hgexample/index/test?name=LingCDN" />
        </n-tab-pane>
        <n-tab-pane name="home" tab="前台页面">
          <Form path="/home/hgexample/index/test?name=LingCDN" />
        </n-tab-pane>
        <n-tab-pane name="static" tab="静态目�?>
          <Form path="/addons/hgexample/default" />
        </n-tab-pane>
        <n-tab-pane name="websocket" tab="Websocket">
          <Form path="/socket/hgexample/index/test?name=LingCDN" :showQrcode="false" />
          <WebsocketTest />
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import Form from './form.vue';
  import WebsocketTest from './websocketTest.vue';
</script>

<style scoped>
  ::v-deep(.card-tabs .n-tabs-nav--bar-type) {
    padding-left: 4px;
  }
</style>


