<template>
  <div>
    <!-- <a-form-model :label-col="labelCol" labelAlign="left">
      <a-form-model-item label="HTTPS" class="info-item">
        <template v-if="enableHttps">已启用</template>
        <template v-else>未启用</template>
      </a-form-model-item>
      <a-form-model-item label="证书名称">
        <template v-if="enableHttps">{{ certName }}</template>
        <template v-else>无</template>
        <a-button type="link" @click="onEditCert">编辑证书</a-button>
      </a-form-model-item>
    </a-form-model> -->
    <a-row>
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">HTTPS :</span>
        </a-col>
        <a-col :span="20">
          <span v-if="form.is_https == 1">已启用</span>
          <span v-else>未启用</span>
        </a-col>
        <a-col :span="2">
          <a @click="onEditCert">修改</a>
        </a-col>
      </div>
    </a-row>

    <a-row v-if="form.is_https == 1">
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">证书 :</span>
        </a-col>
        <a-col :span="12">
          <span>{{ form.cert_name }} </span>
        </a-col>
      </div>
    </a-row>
    <a-row v-if="form.is_https == 1">
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">强制HTTPS :</span>
        </a-col>
        <a-col :span="12">
          <span v-if="form.is_force_https == 1">是</span>
          <span v-else>否</span>
        </a-col>
      </div>
    </a-row>

    <a-modal
      title="修改HTTPS配置"
      :visible="visible"
      :confirm-loading="confirmLoading"
      @ok="handleOk"
      @cancel="handleCancel"
    >
      <a-form-model :model="form" :label-col="{ span: 5 }">
        <a-form-model-item label="开启HTTPS">
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
            <a-select-option :value="0">请选择证书</a-select-option>
            <a-select-option v-for="(item, index) in certificates" :value="item.id" :key="index">{{
              item.name
            }}</a-select-option>
          </a-select>
        </a-form-model-item>
        <a-form-model-item v-if="enableHttps" label="强制跳转">
          <a-switch v-model="isForceHttps" @change="onChange" />
        </a-form-model-item>
      </a-form-model>
    </a-modal>
  </div>
</template>

<script>
import store from '@/store'
// import { GetAll } from '@/api/certificate'
import { UpdateSiteHttpsConfigs } from '@/api/site'
import { GetHttpsConfigs } from '@/api/site/https'
import { GetCandidateCertificates } from '@/api/site/cert'
import { boolean } from 'yargs'
// import { GetCertName } from '@/api/certificate'
export default {
  data() {
    return {
      labelCol: { span: 2 },
      wrapperCol: { span: 20 },
      enableHttps: false,
      isForceHttps: false,

      form: {
        is_https: 0,
        cert_id: 0,
        is_force_https: 0,
        cert_name: ''
      },
      // edit
      visible: false,
      confirmLoading: false,
      certificates: []
    }
  },
  methods: {
    onChange() {},
    handleOk() {
      let payload = {}
      Object.assign(payload, this.form)
      if (this.enableHttps) {
        const isForceHttps = this.isForceHttps ? 1 : 0
        payload.is_https = 1
        payload.is_force_https = isForceHttps
      } else {
        payload.is_https = 0
        payload.cert_id = 0
        payload.is_force_https = 0
      }

      let id = this.$route.params.id
      UpdateSiteHttpsConfigs(id, payload)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('更新成功！')
            this.doGetConfigs()
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
      GetHttpsConfigs(id).then(res => {
        if (res.code == 0) {
          this.form = res.data
          // if (this.form.is_https) {
          //   this.enableHttps = true
          //   GetCertName(this.form.cert_id).then(r => {
          //     if (r.code == 0) {
          //       this.certName = r.data
          //     }
          //   })
          // } else {
          //   this.certName = '无'
          // }

          this.enableHttps = Boolean(this.form.is_https)
          this.isForceHttps = Boolean(this.form.is_force_https)
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
