<template>
  <page-header-wrapper
    :tab-list="tabList"
    :tab-active-key="tabActiveKey"
    :tab-change="(key) => {
      this.tabActiveKey = key
    }"
    @back="() => $router.go(-1)"
  >
    <component :is="tabActiveKey"></component>
  </page-header-wrapper>
</template>

<script>
import { profile, loginLog } from './components'
const tabList = [
  { key: 'profile', tab: '用户信息' },
  { key: 'loginLog', tab: '登陆日志' }
]
export default {
  data () {
    return {
      tabList,
      tabActiveKey: 'profile'
      // END
    }
  },
  components: {
    profile,
    loginLog
  }
}
</script>