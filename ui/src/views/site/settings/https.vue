<template>
  <div>
    <a-form-model :label-col="labelCol" labelAlign="left">
      <a-form-model-item label="HTTPS" class="info-item">
        <template v-if="enableHttps">已启用</template>
        <template v-else>未启用</template>
      </a-form-model-item>
      <a-form-model-item label="证书名称">
        <template v-if="enableHttps">{{ certName }}</template>
        <template v-else>无</template>
        <a-button type="link" @click="onEditCert">编辑证书</a-button>
      </a-form-model-item>
    </a-form-model>
    <a-modal
      title="编辑证书"
      :visible="visible"
      :confirm-loading="confirmLoading"
      @ok="handleOk"
      @cancel="handleCancel"
    >
      <a-form-model :model="form" :label-col="{ span: 5 }">
        <a-form-model-item label="启用HTTPS">
          <a-switch v-model="enableHttps" @change="onChange" />
        </a-form-model-item>
        <a-form-model-item
          v-if="enableHttps"
          label="证书选择"
          prop="policy_id"
          :wrapper-col="{ span: 10 }"
          type="hidden"
        >
          <a-select
            show-search
            placeholder="请选择证书"
            option-filter-prop="children"
            v-model="form.cert_id"
            @dropdownVisibleChange="dropdownVisibleChange"
          >
            <a-select-option v-for="(item, index) in certificates" :value="item.id" :key="index">{{
              item.name
            }}</a-select-option>
          </a-select>
        </a-form-model-item>
      </a-form-model>
    </a-modal>
  </div>
</template>

<script>
import store from '@/store'
// import { GetAll } from '@/api/certificate'
import { GetConfigs, UpdateSiteHttps } from '@/api/site'
import { GetCandidateCertificates } from '@/api/site/cert'
import { GetCertName } from '@/api/certificate'
export default {
  data() {
    return {
      labelCol: { span: 2 },
      wrapperCol: { span: 20 },
      enableHttps: false,
      certName: '无',
      form: {
        is_https: 0,
        cert_id: null
      },
      // edit
      visible: false,
      confirmLoading: false,
      certificates: []
    }
  },
  methods: {
    onChange() {
      if (this.form.cert_id == 0) {
        this.form.cert_id = null
      }
    },
    handleOk() {
      let payload = {}
      Object.assign(payload, this.form)
      if (this.enableHttps) {
        payload.is_https = 1
      } else {
        payload.is_https = 0
        payload.cert_id = 0
      }
      let id = this.$route.params.id
      UpdateSiteHttps(id, payload)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('更新成功！')
            this.visible = false
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(() => {
          this.$message.error('网络异常，请稍候再试')
        })
    },

    handleCancel() {
      this.visible = false
    },

    onEditCert() {
      this.visible = true
    },

    dropdownVisibleChange() {
      this.doGetAllCertificates()
    },

    doGetAllCertificates() {
      let domain = store.state.site.domain || ''
      let params = {
        domain: domain
      }
      let siteID = this.$route.params.id
      GetCandidateCertificates(siteID, params).then(res => {
        this.certificates = res.data
      })
    },

    doGetConfigs() {
      let id = this.$route.params.id
      GetConfigs(id).then(res => {
        if (res.code == 0) {
          this.form.is_https = res.data.is_https
          this.form.cert_id = res.data.cert_id
          if (this.form.is_https) {
            this.enableHttps = true
            GetCertName(this.form.cert_id).then(r => {
              if (r.code == 0) {
                this.certName = r.data
              }
            })
          }
        }
      })
    }
  },
  mounted() {
    this.doGetConfigs()
  },
  activated() {
    this.doGetConfigs()
    this.doGetAllCertificates()
  }
}
</script>

<style lang="less" scoped></style>
