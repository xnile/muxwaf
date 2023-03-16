<template>
  <page-header-wrapper>
    <div class="page-header-index-wide">
      <a-card :bordered="false" :bodyStyle="{ padding: '16px 0', height: '100%' }" :style="{ height: '100%' }">
        <div class="account-settings-info-main" :class="{ mobile: isMobile }">
          <div class="account-settings-info-left">
            <a-menu
              :mode="isMobile ? 'horizontal' : 'inline'"
              :style="{ border: '0', width: isMobile ? '560px' : 'auto' }"
              :selectedKeys="selectedKeys"
              type="inner"
              @openChange="onOpenChange"
            >
              <a-menu-item key="/system/site/:id/settings/basic">
                <router-link :to="{ name: 'basicSettings' }">
                  基本配置
                </router-link>
              </a-menu-item>
              <a-menu-item key="/system/site/:id/settings/origin">
                <router-link :to="{ name: 'originSettings' }">
                  源站配置
                </router-link>
              </a-menu-item>
              <a-menu-item key="/system/site/:id/settings/https">
                <router-link :to="{ name: 'httpsSettings' }">
                  HTTPS配置
                </router-link>
              </a-menu-item>
              <a-menu-item key="/system/site/:id/settings/regionBlacklist">
                <router-link :to="{ name: 'regionBlacklistSettings' }">
                  IP黑名单(地域级)
                </router-link>
              </a-menu-item>
            </a-menu>
          </div>
          <div class="account-settings-info-right">
            <div class="account-settings-info-title">
              <span>{{ $t($route.meta.title) }}</span>
              <span>{{ operation }}</span>
            </div>
            <route-view @changeOperation="onChangeOperation($event)"></route-view>
          </div>
        </div>
      </a-card>
    </div>
  </page-header-wrapper>
</template>

<script>
import { RouteView } from '@/layouts'
import { baseMixin } from '@/store/app-mixin'

export default {
  components: {
    RouteView
  },
  mixins: [baseMixin],
  data() {
    return {
      // horizontal  inline
      mode: 'inline',
      operation: '',

      openKeys: [],
      selectedKeys: [],

      // cropper
      preview: {},
      option: {
        img: '/avatar2.jpg',
        info: true,
        size: 1,
        outputType: 'jpeg',
        canScale: false,
        autoCrop: true,
        // 只有自动截图开启 宽度高度才生效
        autoCropWidth: 180,
        autoCropHeight: 180,
        fixedBox: true,
        // 开启宽度和高度比例
        fixed: true,
        fixedNumber: [1, 1]
      },

      pageTitle: ''
    }
  },
  mounted() {
    this.updateMenu()
    console.log(this.$route.params.id)
  },
  methods: {
    onOpenChange(openKeys) {
      this.openKeys = openKeys
    },
    updateMenu() {
      const routes = this.$route.matched.concat()
      this.selectedKeys = [routes.pop().path]
    },
    onChangeOperation(e) {
      console.log(e)
      this.operation = e
    }
  },
  watch: {
    $route(val) {
      this.updateMenu()
    }
  }
}
</script>

<style lang="less" scoped>
.account-settings-info-main {
  width: 100%;
  display: flex;
  height: 100%;
  overflow: auto;

  &.mobile {
    display: block;

    .account-settings-info-left {
      border-right: unset;
      border-bottom: 1px solid #e8e8e8;
      width: 100%;
      height: 50px;
      overflow-x: auto;
      overflow-y: scroll;
    }
    .account-settings-info-right {
      padding: 20px 40px;
    }
  }

  .account-settings-info-left {
    border-right: 1px solid #e8e8e8;
    width: 224px;
  }

  .account-settings-info-right {
    flex: 1 1;
    padding: 8px 40px;

    .account-settings-info-title {
      color: rgba(0, 0, 0, 0.85);
      font-size: 20px;
      font-weight: 500;
      line-height: 28px;
      margin-bottom: 12px;
    }
    .account-settings-info-view {
      padding-top: 12px;
    }
  }
}
</style>
