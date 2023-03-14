<template>
  <page-header-wrapper>
    <!-- 添加按钮 -->
    <template v-slot:extra>
      <a-button type="primary" @click="onAdd">添加用户</a-button>
    </template>

    <a-card title="用户列表">
      <!-- 表格 -->
      <a-table :columns="columns" :dataSource="list" :rowKey="record => record.id" :pagination="false">
        <template slot="status" slot-scope="text">
          <template v-if="text === 0">
            <a-badge status="success" text="正常" />
          </template>
          <template v-else>
            <a-badge status="error" text="封禁" />
          </template>
        </template>
        <span slot="created_at" slot-scope="text">{{ text | moment }}</span>
        <template slot="last_sign_in_at" slot-scope="text">
          <template v-if="text === 0">
            <span>-</span>
          </template>
          <template v-else>
            <span>{{ text | moment }}</span>
          </template>
        </template>
        <template slot="operation" slot-scope="text, record">
          <!-- <a-button
            type="link"
            @click="updateItemStatus(record)"
          >{{ record.status === 1 ? '停用' : '启用' }}</a-button>-->

          <a-button type="link" size="small" @click="toDetail(record.id)">详情</a-button>
          <!-- <a-button type="link" size="small" @click="deleteItem(record)">删除</a-button> -->
          <a-dropdown :trigger="['click']">
            <a class="ant-dropdown-link" @click="e => e.preventDefault()">
              设置
              <a-icon type="down" />
            </a>
            <a-menu slot="overlay" @click="handleMenuClick">
              <a-menu-item key="1">
                <span @click="editItem(record)">编辑</span>
              </a-menu-item>
              <a-menu-item key="2">
                <span @click="blockItem(record.id)">{{ record.blocked_at === 0 ? '封禁' : '解除封禁' }}</span>
              </a-menu-item>
              <a-menu-divider />
              <a-menu-item key="3">
                <span class="text-danger" @click="deleteItem(record.id)">删除此用户</span>
              </a-menu-item>
            </a-menu>
            <a-button size="small" style="margin-left: 8px"> <a-icon type="setting" />设置 </a-button>
          </a-dropdown>
        </template>
      </a-table>
      <!-- 表格end -->
      <!-- 分页 -->
      <a-row :style="{ marginTop: '10px' }" v-if="meta.total">
        <a-col :span="24">
          <a-pagination
            style="float: right"
            show-size-changer
            show-quick-jumper
            show-less-items
            :show-total="total => `共 ${total} 条记录 第${meta.page_num}/${meta.pages}页`"
            :total="meta.total"
            :pageSize="queryParams.page_size"
            @showSizeChange="onShowSizeChange"
            @change="onChange"
          />
        </a-col>
      </a-row>
      <!-- 分页end -->
    </a-card>

    <!-- 新增 Drawer -->
    <a-drawer
      :title="(operateType == 'add' ? '添加' : '编辑') + '用户'"
      :width="700"
      placement="right"
      :closable="false"
      :visible="visible"
      :after-visible-change="afterVisibleChange"
      @close="onClose"
    >
      <!-- form -->
      <a-form-model ref="form" :model="form" :rules="formRules" :label-col="labelCol" :wrapper-col="wrapperCol">
        <a-form-model-item label="用户名" prop="username">
          <a-input :disabled="operateType == 'edit'" placeholder="请输入登陆用户名" v-model="form.username"></a-input>
        </a-form-model-item>
        <a-form-model-item label="密码" prop="password">
          <a-input-password placeholder="请输入登陆密码" v-model="form.password" />
        </a-form-model-item>
        <a-form-model-item label="真实名" prop="name">
          <a-input placeholder="请输入真实名" v-model="form.name" />
        </a-form-model-item>
        <a-form-model-item label="手机号" prop="phone">
          <a-input placeholder="请输入手机号" v-model="form.phone" />
        </a-form-model-item>
        <a-form-model-item label="邮箱" prop="email">
          <a-input placeholder="请输入邮箱" v-model="form.email" />
        </a-form-model-item>
      </a-form-model>
      <!-- from-model END -->
      <!-- 取消,确认按钮 -->
      <div class="drawer-submit-button">
        <a-button :style="{ marginRight: '8px' }" @click="onClose">取消</a-button>
        <a-button type="primary" @click="onSubmit">确认</a-button>
      </div>
    </a-drawer>
    <!-- MoDal END -->

    <!-- 编辑 Modal -->
    <a-modal :width="800" v-model="editVisible" title="修改用户信息" @ok="handleOk">
      <!-- form -->
      <a-form-model ref="form" :model="form" :rules="editFormRules" :label-col="labelCol" :wrapper-col="wrapperCol">
        <a-form-model-item label="用户名" prop="username">
          <a-input disabled placeholder="请输入用户名" v-model="form.username"></a-input>
        </a-form-model-item>
        <a-form-model-item label="密码" prop="password">
          <a-input-password placeholder="不重置密码请留空" v-model="form.password" />
        </a-form-model-item>
        <a-form-model-item label="真实名" prop="name">
          <a-input placeholder="请输入真实名" v-model="form.name" />
        </a-form-model-item>
        <a-form-model-item label="手机号" prop="phone">
          <a-input placeholder="请输入手机号" v-model="form.phone" />
        </a-form-model-item>
        <a-form-model-item label="邮箱" prop="email">
          <a-input placeholder="请输入邮箱" v-model="form.email" />
        </a-form-model-item>
      </a-form-model>
      <!-- form END -->
    </a-modal>
    <!-- MoDal END -->
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { InsertUser, ListUsers, UpdateUser, BlockUser, DeleteUser } from '@/api/user'
import templates from '@/views/templates.vue'
const columns = [
  {
    title: '用户名',
    dataIndex: 'username'
  },
  // {
  //   title: '真实名',
  //   dataIndex: 'name'
  // },
  {
    title: '状态',
    dataIndex: 'blocked_at',
    scopedSlots: { customRender: 'status' }
  },
  {
    title: '最后登陆时间',
    dataIndex: 'last_sign_in_at',
    scopedSlots: { customRender: 'last_sign_in_at' }
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    scopedSlots: { customRender: 'created_at' }
  },
  {
    title: '操作',
    dataIndex: 'operation',
    scopedSlots: { customRender: 'operation' },
    fixed: 'right'
  }
]

export default {
  components: { templates },
  data() {
    return {
      columns,
      list: [],
      meta: {},
      queryParams: {
        page_num: 1,
        page_size: 10
      },
      visible: false,
      editVisible: false,
      operateType: '',
      form: {
        id: '',
        username: '',
        password: '',
        name: '',
        phone: '',
        email: '',
        remark: ''
      },
      formRules: {
        username: [{ required: true, message: '请输入用户名！' }],
        password: [{ required: true, message: '请输入密码！' }]
      },
      editFormRules: {
        // name: [{ required: true, message: '请输入姓名！' }],
        // phone: [{ required: true, message: '手机号不能为空！' }]
      },
      labelCol: { span: 4 },
      wrapperCol: { span: 15 }
      // END
    }
  },

  methods: {
    onAdd() {
      this.operateType = 'add'
      this.visible = true
      this.form = {}
    },
    onShowSizeChange() {},
    onChange() {},

    onClose() {
      this.visible = false
    },

    handleMenuClick() {},

    afterVisibleChange(val) {
      console.log('visible', val)
    },

    editItem(record) {
      this.operateType = 'edit'
      // this.visible = true
      // this.$nextTick(() => {
      //   this.formRules.password = null
      // })
      this.editVisible = true
      this.form = record
    },

    blockItem(uid) {
      this.block(uid)
    },

    deleteItem(uid) {
      this.del(uid)
    },

    onSubmit() {
      const _this = this
      this.$refs.form.validate(valid => {
        if (valid) {
          // var parsedobj = JSON.parse(JSON.stringify(this.form))
          // console.log(parsedobj)
          // this.visible = false

          console.log(_this.form)
          switch (_this.operateType) {
            case 'add':
              _this.add(_this.form)
              break
            // case 'edit':
            //   _this.update(_this.form)
            //   break
          }
        }
      })
      // console.log('submit!', this.form)
    },

    handleOk() {
      const _this = this
      this.$refs.form.validate(valid => {
        if (valid) {
          // var parsedobj = JSON.parse(JSON.stringify(this.form))
          // console.log(parsedobj)
          // this.visible = false

          console.log(_this.form)
          switch (_this.operateType) {
            // case 'add':
            //   _this.add(_this.form)
            //   break
            case 'edit':
              _this.update(_this.form)
              break
          }
        }
      })
    },

    add(data) {
      const _this = this
      InsertUser(data)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('添加成功！')
            _this.getList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {})
    },

    update(data) {
      const _this = this
      UpdateUser(data.id, data)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('更新成功！')
            _this.getList()
            _this.editVisible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(() => {
          _this.$message.error('网络异常，请稍候再试')
        })
    },

    block(uid) {
      const _this = this
      BlockUser(uid)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('操作成功！')
            _this.getList()
          }
        })
        .catch(err => {})
    },

    del(uid) {
      const _this = this
      DeleteUser(uid)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('删除成功！')
            _this.getList()
          }
        })
        .catch(err => {})
    },

    getList() {
      ListUsers(this.queryParams).then(res => {
        this.list = res.data.list
        this.meta = res.data.meta
      })
    },

    toDetail(userID) {
      this.$router.push(`/user/profile/${userID}`)
    }
  },

  mounted() {
    this.getList()
  }
}
</script>
