<template>
  <div>
    <a-row>
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
    </a-row>
    <!-- 新增 Modal -->
    <a-modal :width="800" v-model="visible" title="修改" @ok="onOk">
      <a-form-model
        layout="horizontal"
        :model="form"
        v-bind="{
          labelCol: { span: 5 },
          wrapperCol: { span: 14 }
        }"
      >
        <a-form-model-item label="国内地域级IP黑名单：">
          <a-input
            v-model="form.regions"
            placeholder="输入内地省份名，多个用','号分隔，如: 北京,南京"
            @change="onChange"
          />
        </a-form-model-item>
        <a-form-model-item label="海外地域级IP黑名单">
          <a-input v-model="form.countries" placeholder="输入国家名，多个用','号分隔，如：美国,日本" />
        </a-form-model-item>
        <!-- <a-form-model-item :wrapper-col="{ span: 14, offset: 1 }">
          <a-button type="primary" :disabled="disabled" @click="onSubmit">
            更新
          </a-button>
        </a-form-model-item> -->
      </a-form-model>
    </a-modal>
  </div>
</template>

<script>
import { GetSiteRegionBlacklist, UpdateSiteRegionBlacklist } from '@/api/site/regionBlacklist'

export default {
  data() {
    return {
      disabled: true,
      visible: false,
      form: {
        regions: '',
        countries: '',
        match_mode: 0
      }
    }
  },
  methods: {
    onEdit() {
      this.visible = true
    },
    onOk() {
      this.doUpdateSiteRegionBlacklist()
    },

    onChange() {},

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
            this.visible = false
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
    this.doGetSiteRegionBlacklist()
  }
}
</script>

<style scoped>
.list-lable {
  color: rgba(0, 0, 0, 0.65);
  font-size: 14px;
  /* line-height: 30px; */
  /* font-weight: 500; */
}
.item {
  /* height: 10px; */
  line-height: 35px;
}

.right {
  text-align: right;
}
</style>
