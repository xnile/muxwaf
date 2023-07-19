<template>
  <div>
    <a-row>
      <div class="item">
        <a-col :span="22"></a-col>
        <a-col :span="2">
          <a @click="onEdit">修改</a>
        </a-col>
      </div>
    </a-row>

    <!-- <a-row>
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">国内封禁地区 :</span>
        </a-col>
        <a-col :span="20">
          <span>{{ form.regions || '无' }}</span>
        </a-col>
        <a-col :span="2">
          <a @click="onEdit">修改</a>
        </a-col>
      </div>
    </a-row>
    <a-row>
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">封禁国家 :</span>
        </a-col>
        <a-col :span="12">
          <span>{{ form.countries || '无' }}</span>
        </a-col>
      </div>
    </a-row> -->
    <!-- 新增 Modal -->
    <!-- <a-modal :width="800" v-model="visible" title="修改" @ok="onOk"> -->
    <a-form-model ref="form" :model="form" :rules="rules" label-align="left" layout="horizontal">
      <a-form-model-item label="国内地域级IP黑名单：">
        <a-input
          v-if="editable"
          v-model="form.regions"
          placeholder="输入内地省份名，多个用','号分隔，如: 北京,南京"
          @change="onChange"
          style="width: 50%"
        />
        <template v-else>
          <span style="padding-left: 20px;">{{ form.regions || '无' }}</span>
        </template>
      </a-form-model-item>
      <a-form-model-item label="海外地域级IP黑名单" style="margin-top: 10px;">
        <a-input
          v-if="editable"
          v-model="form.countries"
          placeholder="输入国家名，多个用','号分隔，如：美国,日本"
          style="width: 50%"
        />
        <template v-else>
          <span style="padding-left: 20px;">{{ form.countries || '无' }}</span>
        </template>
      </a-form-model-item>
      <template v-if="editable">
        <a-row style="margin-top: 40px;">
          <a-col>
            <a-button style="margin-left: 10px;" type="primary" @click="onOK">
              保存
            </a-button>
            <a-button style="margin-left: 10px;" @click="onCancel">
              取消
            </a-button>
          </a-col>
        </a-row>
      </template>
    </a-form-model>
  </div>
</template>

<script>
import { GetSiteRegionBlacklist, UpdateSiteRegionBlacklist } from '@/api/site/regionBlacklist'

export default {
  data() {
    return {
      disabled: true,
      editable: false,
      form: {
        regions: '',
        countries: '',
        match_mode: 0
      },
      rules: {}
    }
  },
  methods: {
    onEdit() {
      this.editable = true
    },
    onOK() {
      this.doUpdateSiteRegionBlacklist()
    },

    onChange() {},

    onCancel() {
      this.editable = false
    },

    doUpdateSiteRegionBlacklist() {
      let siteID = this.$route.params.id
      let countries = this.form.countries ? this.form.countries.split(',') : []
      let regions = this.form.regions ? this.form.regions.split(',') : []

      let data = {
        countries,
        regions,
        match_mode: this.form.match_mode
      }
      UpdateSiteRegionBlacklist(siteID, data)
        .then(res => {
          if (res.code == 0) {
            this.editable = false
            this.$message.success('更新成功')
            this.doGetSiteRegionBlacklist()
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(() => {
          this.$message.error('网络异常，请稍候再试')
        })
    },

    doGetSiteRegionBlacklist() {
      let siteID = this.$route.params.id
      GetSiteRegionBlacklist(siteID)
        .then(res => {
          if (res.code == 0) {
            this.form.countries = res.data.countries.join(',')
            this.form.regions = res.data.regions.join(',')
            this.form.match_mode = res.data.match_mode
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(() => {
          this.$message.error('网络异常，请稍候再试')
        })
    }
  },
  computed: {},
  mounted() {
    this.doGetSiteRegionBlacklist()
  },
  activated() {
    // 在首次挂载、
    // 以及每次从缓存中被重新插入的时候调用
    this.editable = false
    this.doGetSiteRegionBlacklist()
  }
}
</script>

<style scoped>
.ant-form-item {
  margin-bottom: 0;
}
</style>
