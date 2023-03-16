<template>
  <div>
    <a-row :gutter="16">
      <a-col :md="24" :lg="16">
        <a-form-model
          ref="form"
          labelAlign="left"
          :model="form"
          :rules="rules"
          :label-col="labelCol"
          :wrapper-col="wrapperCol"
          layout="vertical"
        >
          <a-form-model-item label="域名" :wrapper-col="{ span: 15 }">
            {{ domain }}
          </a-form-model-item>
          <a-form-model-item label="回源协议">
            <a-radio-group v-model="form.origin_protocol" :disabled="disableEdit" @change="onChange">
              <a-radio :value="1">HTTP</a-radio>
              <a-radio :value="2">HTTPS</a-radio>
              <a-radio :value="3">跟随</a-radio>
            </a-radio-group>
          </a-form-model-item>
          <a-form-model-item label="前置CDN">
            <a-switch v-model="pre_cdn" :disabled="disableEdit" @change="onChange" />
          </a-form-model-item>
          <a-form-model-item v-if="pre_cdn" label="获取IP的Header">
            <a-radio-group v-model="real_ip_header_type" :disabled="disableEdit" @change="onChange">
              <a-radio :value="0" :style="radioStyle">取X-Forwarded-For中的第一个IP作为客户端源IP</a-radio>
              <a-radio :value="1" :style="radioStyle">自定义Header</a-radio>
            </a-radio-group>
            <a-input
              v-if="real_ip_header_type"
              type="text"
              v-model="form.real_ip_header"
              :disabled="disableEdit"
              @change="onChange"
            />
          </a-form-model-item>
          <!-- form END -->
        </a-form-model>
        <!-- 取消,确认按钮 -->
        <div :style="buttonStyle">
          <a-button type="primary" @click="onEdit" v-if="disableEdit">编辑</a-button>
          <a-button type="primary" @click="onSubmit" v-if="!disableEdit">保存</a-button>
          <a-button :style="{ marginLeft: '10px' }" v-if="!disableEdit" @click="onCancel">取消</a-button>
        </div>
        <!-- END -->
      </a-col>
    </a-row>
  </div>
</template>

<script>
import store from '@/store'
import { GetConfigs, UpdateSiteConfigs } from '@/api/site'
export default {
  data() {
    return {
      domain: '',
      real_ip_header_type: 0,
      // disableSubmit: true,
      disableEdit: true,
      pre_cdn: false,
      form: {
        origin_protocol: 1,
        is_real_ip_from_header: 0,
        real_ip_header: ''
      },
      rules: {
        // protocol: [{ required: true, message: '' }]
      },
      labelCol: { span: 3 },
      wrapperCol: { span: 20 },
      radioStyle: {
        display: 'block',
        height: '30px',
        lineHeight: '30px'
      },
      buttonStyle: {
        // position: 'absolute',
        // right: 0,
        // bottom: 0,
        // width: '100%',
        // borderTop: '1px solid #e9e9e9',
        margin: '30px 0px'
        // background: '#fff',
        // textAlign: 'left',
        // zIndex: 1
      }
    }
  },
  methods: {
    onSubmit() {
      let payload = {}
      Object.assign(payload, this.form)
      payload.is_real_ip_from_header = 1

      if (!this.pre_cdn) {
        payload.is_real_ip_from_header = 0
        payload.real_ip_header = ''
      }

      if (this.pre_cdn && this.real_ip_header_type == 0) {
        if (this.real_ip_header_type == 0) {
          payload.real_ip_header = 'X-Forwarded-For'
        }
      }

      let id = this.$route.params.id
      UpdateSiteConfigs(id, payload)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('更新成功！')
            this.disableEdit = true
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(() => {
          this.$message.error('网络异常，请稍候再试')
        })

      // this.disableSubmit = true
    },
    onChange() {
      // if (this.pre_cdn) {
      //   this.real_ip_header_type = 0
      //   this.form.real_ip_header = ''
      // }
      // console.log(this.form.config.pre_cdn)
    },
    onCancel() {
      // window.location.reload()
      this.$router.push({ name: 'basicSettings' })
      this.getSiteConfigs()
      this.disableEdit = true
    },
    onEdit() {
      this.disableEdit = false
    },

    getSiteConfigs() {
      GetConfigs(this.$route.params.id).then(res => {
        if (res.code == 0) {
          this.form.origin_protocol = res.data.origin_protocol
          this.pre_cdn = Boolean(res.data.is_real_ip_from_header)
          this.form.real_ip_header = res.data.real_ip_header
          if (res.data.real_ip_header == 'X-Forwarded-For') {
            this.real_ip_header_type = 0
          } else {
            this.real_ip_header_type = 1
          }
        }
      })
    }
  },
  mounted() {
    this.domain = store.state.site.domain
    this.getSiteConfigs()
  },
  activated() {
    // 在首次挂载、
    // 以及每次从缓存中被重新插入的时候调用
    this.getSiteConfigs()
  }
}
</script>

<style lang="less" scoped>
// .info-item {
//   padding: 0px 0px 10px 0px;
// }

// .info-item span {
//   font-size: 14px;
//   color: rgba(0, 0, 0, 0.85);
// }
</style>
