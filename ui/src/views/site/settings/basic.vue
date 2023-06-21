<template>
  <div>
    <a-row>
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">域名 :</span>
        </a-col>
        <a-col :span="20">
          <span>{{ domain }}</span>
        </a-col>
        <a-col :span="2">
          <a @click="onEdit">修改</a>
        </a-col>
      </div>
    </a-row>
    <!-- <a-row>
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">回源协议 :</span>
        </a-col>
        <a-col :span="12">
          <span v-if="form.origin_protocol == 1">http</span>
          <span v-else-if="form.origin_protocol == 2">https</span>
          <span v-else-if="form.origin_protocol == 3">跟随</span>
        </a-col>
      </div>
    </a-row> -->
    <a-row>
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">前置CDN :</span>
        </a-col>
        <a-col :span="12">
          <span v-if="form.is_real_ip_from_header == 1">是</span>
          <span v-else>否</span>
        </a-col>
      </div>
    </a-row>
    <a-row v-if="form.is_real_ip_from_header == 1">
      <div class="item">
        <a-col :span="2">
          <span class="list-lable">获取IP头 :</span>
        </a-col>
        <a-col :span="12">
          <span v-if="real_ip_header_type == 0">取X-Forwarded-For中的第一个IP作为客户端源IP</span>
          <span v-else-if="real_ip_header_type == 1">{{ form.real_ip_header }}</span>
        </a-col>
      </div>
    </a-row>

    <!-- 新增 Modal -->
    <a-modal :width="800" v-model="visible" title="基本配置" @ok="onOk">
      <!-- <a-form-model-item label="域名">
        {{ domain }}
      </a-form-model-item> -->
      <!-- <a-form-model-item label="回源协议">
        <a-radio-group v-model="form.origin_protocol">
          <a-radio :value="1">HTTP</a-radio>
          <a-radio :value="2">HTTPS</a-radio>
          <a-radio :value="3">跟随</a-radio>
        </a-radio-group>
      </a-form-model-item> -->
      <a-form-model-item label="前置CDN">
        <a-switch v-model="pre_cdn" />
      </a-form-model-item>
      <a-form-model-item v-if="pre_cdn" label="获取IP的Header">
        <a-radio-group v-model="real_ip_header_type">
          <a-radio :value="0" :style="radioStyle">取X-Forwarded-For中的第一个IP作为客户端源IP</a-radio>
          <a-radio :value="1" :style="radioStyle">自定义Header</a-radio>
        </a-radio-group>
        <a-input
          v-if="real_ip_header_type"
          type="text"
          v-model="form.real_ip_header"
          style="display: block; width: 20em"
        />
      </a-form-model-item>
    </a-modal>
  </div>
</template>

<script>
import store from '@/store'
import { GetDomain, GetConfigs, UpdateSiteBasicConfigs } from '@/api/site'
export default {
  data() {
    return {
      visible: false,
      domain: '',
      real_ip_header_type: 0,
      pre_cdn: false,
      form: {
        // origin_protocol: 1,
        is_real_ip_from_header: 0,
        real_ip_header: ''
      },
      rules: {
        // protocol: [{ required: true, message: '' }]
      },
      // labelCol: { span: 3 },
      // wrapperCol: { span: 20 },
      radioStyle: {
        display: 'block',
        // height: '30px',
        lineHeight: '30px'
      }
    }
  },
  methods: {
    onEdit() {
      this.visible = true
    },
    // onOk() {},

    onOk() {
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
      UpdateSiteBasicConfigs(id, payload)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('更新成功！')
            this.visible = false
            this.getSiteConfigs()
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
      this.$router.push({ name: 'basicSettings' })
      this.getSiteConfigs()
    },

    getDomain() {
      let id = this.$route.params.id
      GetDomain(id).then(res => {
        if (res.code == 0) {
          this.domain = res.data
        }
      })
    },

    getSiteConfigs() {
      GetConfigs(this.$route.params.id).then(res => {
        if (res.code == 0) {
          this.pre_cdn = Boolean(res.data.is_real_ip_from_header)
          this.form.real_ip_header = res.data.real_ip_header
          this.form.is_real_ip_from_header = res.data.is_real_ip_from_header
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
    if (this.domain == '') {
      this.getDomain()
    }
    this.getSiteConfigs()
  },
  activated() {
    // 在首次挂载、
    // 以及每次从缓存中被重新插入的时候调用
    this.getSiteConfigs()
    this.domain = store.state.site.domain
    if (this.domain == '') {
      this.getDomain()
    }
  }
}
</script>

<style scoped>
.list-lable {
  color: rgba(0, 0, 0, 0.65);
  font-size: 14px;
  /* line-height: 30px; */
  /* font-weight: 400; */
}
.item {
  /* height: 10px; */
  line-height: 35px;
}

.right {
  text-align: right;
}
</style>
