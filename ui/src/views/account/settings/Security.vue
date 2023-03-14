<template>
  <div>
    <a-list itemLayout="horizontal" :dataSource="data">
      <a-list-item slot="renderItem" slot-scope="item, index" :key="index">
        <a-list-item-meta>
          <a slot="title">{{ item.title }}</a>
          <span slot="description">
            <span class="security-list-description">{{ item.description }}</span>
            <span v-if="item.value"> : </span>
            <span class="security-list-value">{{ item.value }}</span>
          </span>
        </a-list-item-meta>
        <template v-if="item.actions">
          <a slot="actions" @click="item.actions.callback">{{ item.actions.title }}</a>
        </template>
      </a-list-item>
    </a-list>
    <!-- 新增 Modal -->
    <a-modal :width="600" v-model="visible" title="修改密码" @ok="onOk" @cancel="onCancel">
      <!-- form -->
      <a-form-model ref="form" :model="form" :rules="formRules" :label-col="labelCol" :wrapper-col="wrapperCol">
        <a-form-model-item label="当前密码" prop="password">
          <a-input-password placeholder="请输入当前密码" v-model="form.password" />
        </a-form-model-item>
        <a-form-model-item label="新密码" prop="newPassword">
          <a-input-password placeholder="请输入新密码" v-model="form.newPassword" />
        </a-form-model-item>
        <a-form-model-item label="再输入一次新密码" prop="newPassword2">
          <a-input-password placeholder="再输入一次新密码" v-model="form.newPassword2" />
        </a-form-model-item>
      </a-form-model>
      <!-- from-model END -->
      <!-- 取消,确认按钮 -->
      <!-- <div class="drawer-submit-button">
        <a-button :style="{ marginRight: '8px' }" @click="onClose">取消</a-button>
        <a-button type="primary" @click="onSubmit">确认</a-button>
      </div> -->
    </a-modal>
    <!-- Modal END -->
  </div>
</template>

<script>
import md5 from 'md5'
import { mapGetters } from 'vuex'
import { ResetPassword } from '@/api/user'
export default {
  data() {
    return {
      visible: false,
      labelCol: { span: 6 },
      wrapperCol: { span: 15 },
      form: {
        password: '',
        newPassword: '',
        newPassword2: ''
      },
      formRules: {
        password: [{ required: true, message: '请输入当前密码！' }],
        newPassword: [{ required: true, min: 8, message: '请输入8位以上密码' }],
        newPassword2: [{ required: true, min: 8, message: '请输入8位以上密码' }]
      }
    }
  },
  methods: {
    onCancel() {
      this.visible = false
    },
    onEditPassword() {
      this.visible = true
    },
    onOk() {
      const _this = this
      this.$refs.form.validate(valid => {
        if (valid) {
          // var parsedobj = JSON.parse(JSON.stringify(this.form))
          // console.log(parsedobj)
          // this.visible = false
          if (_this.form.newPassword != _this.form.newPassword2) {
            _this.$message.error('两次输入密码不一致')
          } else {
            let payload = {
              old_password: md5(_this.form.password),
              new_password: md5(_this.form.newPassword)
            }
            ResetPassword(payload)
              .then(res => {
                if (res.code == 0) {
                  _this.$message.success('修改成功！')
                  _this.visible = false
                } else {
                  _this.$message.error(res.msg)
                }
              })
              .catch(() => {
                _this.$message.error('网络异常，请稍候再试')
              })
          }
        }
      })
    }
  },
  computed: {
    data() {
      return [
        {
          title: '账号密码',
          description: '当前密码强度',
          value: '强',
          actions: {
            title: '修改',
            callback: this.onEditPassword
          }
        }
      ]
    },
    ...mapGetters(['nickname', 'email'])
  }
}
</script>

<style scoped></style>
