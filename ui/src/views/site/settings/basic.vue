<template>
  <div>
    <a-row v-if="!editable">
      <div class="item">
        <a-col :span="22"></a-col>
        <a-col :span="2">
          <a @click="onEdit">修改</a>
        </a-col>
      </div>
    </a-row>

    <a-form-model
      ref="form"
      :model="form"
      :rules="rules"
      :label-col="labelCol"
      :wrapper-col="wrapperCol"
      label-align="left"
    >
      <a-form-model-item label="域名">
        {{ domain }}
      </a-form-model-item>
      <a-form-model-item label="前置CDN">
        <a-switch v-if="editable" v-model="pre_cdn" />
        <!-- eslint-disable-next-line -->
        <template v-else
          ><span v-if="form.is_real_ip_from_header == 1">是</span>
          <span v-else>否</span>
        </template>
      </a-form-model-item>
      <a-form-model-item label="获取IP的Header" v-if="pre_cdn">
        <template v-if="editable">
          <a-radio-group v-model="real_ip_header_type" @change="onRealIPHeaderTypeChange">
            <a-radio :value="0" :style="radioStyle">取X-Forwarded-For中的第一个IP作为客户端源IP</a-radio>
            <a-radio :value="1" :style="radioStyle">自定义Header</a-radio>
          </a-radio-group>
          <a-input
            v-if="real_ip_header_type"
            type="text"
            v-model="form.real_ip_header"
            style="display: block; width: 20em"
          />
        </template>

        <template v-else>
          <span v-if="real_ip_header_type == 0">取X-Forwarded-For中的第一个IP作为客户端源IP</span>
          <span v-else-if="real_ip_header_type == 1">{{ form.real_ip_header }}</span>
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
import store from '@/store'
import { GetBasicConfigs, UpdateSiteBasicConfigs } from '@/api/site'
export default {
  data() {
    return {
      // visible: false,
      editable: false,
      domain: '',
      real_ip_header_type: 0,
      pre_cdn: false,
      form: {
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
        // height: '30px',
        lineHeight: '30px'
      }
    }
  },
  methods: {
    onEdit() {
      // this.visible = true
      this.editable = true
    },
    // onOk() {},

    onOK() {
      let payload = {}
      Object.assign(payload, this.form)
      payload.is_real_ip_from_header = 1

      if (!this.pre_cdn) {
        payload.is_real_ip_from_header = 0
        payload.real_ip_header = ''
      } else {
        if (this.real_ip_header_type == 0) {
          payload.real_ip_header = 'X-Forwarded-For'
        } else if (this.real_ip_header_type == 1 && payload.real_ip_header == '') {
          this.$message.error('Header不能用空')
          return
        }
      }

      // if (this.pre_cdn && this.real_ip_header_type == 0) {
      //   if (this.real_ip_header_type == 0) {
      //     payload.real_ip_header = 'X-Forwarded-For'
      //   }
      // }

      let id = this.$route.params.id
      UpdateSiteBasicConfigs(id, payload)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('更新成功！')
            this.editable = false
            this.getBasicSiteConfigs()
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(() => {
          this.$message.error('网络异常，请稍候再试')
        })
    },
    // onChange() {},
    onCancel() {
      // window.location.reload()
      this.editable = false
      this.$router.push({ name: 'basicSettings' })
      this.getBasicSiteConfigs()
    },

    onRealIPHeaderTypeChange() {
      if (this.real_ip_header_type == 1) {
        this.form.real_ip_header = ''
      }
    },

    getDomain() {
      // let id = this.$route.params.id
      // GetDomain(id).then(res => {
      //   if (res.code == 0) {
      //     this.domain = res.data
      //   }
      // })
    },

    getBasicSiteConfigs() {
      GetBasicConfigs(this.$route.params.id).then(res => {
        if (res.code == 0) {
          this.domain = res.data.host
          this.pre_cdn = Boolean(res.data.is_real_ip_from_header)
          this.form.real_ip_header = res.data.real_ip_header
          this.form.is_real_ip_from_header = res.data.is_real_ip_from_header

          if (!this.pre_cdn) {
            this.real_ip_header_type = 0
          } else {
            if (res.data.real_ip_header == 'X-Forwarded-For') {
              this.real_ip_header_type = 0
            } else {
              this.real_ip_header_type = 1
            }
          }
        }
      })
    }
  },
  mounted() {
    // this.domain = store.state.site.domain
    // if (this.domain == '') {
    //   this.getDomain()
    // }
    this.getBasicSiteConfigs()
  },
  activated() {
    // 在首次挂载、
    // 以及每次从缓存中被重新插入的时候调用
    this.editable = false
    this.getBasicSiteConfigs()
    // this.domain = store.state.site.domain
    // if (this.domain == '') {
    //   this.getDomain()
    // }
  }
}
</script>

<style scoped>
.ant-form-item {
  margin-bottom: 5px;
}
</style>
