<template>
  <page-header-wrapper>
    <template v-slot:extra>
      <a-button type="primary" @click="onAdd">添加节点</a-button>
    </template>
    <a-card>
      <a-table
        :columns="columns"
        :data-source="list"
        :scroll="{ x: 1300 }"
        :rowKey="record => record.id"
        :pagination="false"
      >
        <template slot="isSampledLogUpload" slot-scope="value">
          <template v-if="value === 1">
            <a-badge status="success" text="已启用" />
          </template>
          <template v-else>
            <a-badge status="warning" text="未启用" />
          </template>
        </template>
        <template slot="lastSyncAt" slot-scope="value">
          <template v-if="value === 0">
            <span>无</span>
          </template>
          <template v-else> {{ value | moment }}</template>
        </template>
        <template slot="lastSyncStatus" slot-scope="value">
          <template v-if="value === 1">
            <a-badge status="success" text="成功" />
          </template>
          <template v-else-if="value === -1">
            <a-badge status="error" text="失败" />
          </template>
          <template v-else>
            <a-badge status="default" text="未知" />
          </template>
        </template>
        <template slot="operation" slot-scope="text, record">
          <a-button type="link" size="small" @click="onSync(record.id)">同步</a-button>
          <!-- <a-button type="link" @click="onSwitchLogUploadStatus(record.id)">{{
            record.is_sampled_log_upload === 1 ? '关闭日志上报' : '开启日志上报'
          }}</a-button>
          <a-button type="link" size="small" @click="onDelete(record.id)">删除</a-button> -->
          <template>
            <a-dropdown>
              <a class="ant-dropdown-link" @click="e => e.preventDefault()"> 更多<a-icon type="down" /> </a>
              <a-menu slot="overlay" @click="e => onClickMore(e.key, record.id)">
                <a-menu-item :key="1">
                  {{ record.is_sampled_log_upload === 1 ? '关闭日志上报' : '开启日志上报' }}
                </a-menu-item>
                <a-menu-item :key="2">
                  删除
                </a-menu-item>
              </a-menu>
            </a-dropdown>
          </template>
        </template>
      </a-table>
    </a-card>
    <a-modal
      :width="700"
      v-model="visible"
      :title="(operateType == 'add' ? '添加' : '编辑') + '节点'"
      @ok="onOK"
      @cancel="onCancel"
    >
      <a-form-model ref="form" :model="form" :rules="rules" :label-col="{ span: 5 }" :wrapper-col="{ span: 15 }">
        <a-form-model-item label="地址" prop="ip_or_domain">
          <a-input placeholder="输入guard地址，可以是IP、域名、Hostname" v-model="form.ip_or_domain"></a-input>
        </a-form-model-item>
        <a-form-model-item label="端口" prop="port">
          <a-input-number placeholder="端口" v-model.number="form.port"></a-input-number>
        </a-form-model-item>
      </a-form-model>
    </a-modal>
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { AddNode, ListNodes, DelNode, SyncCfg, SwitchLogUpload } from '@/api/node'
const columns = [
  {
    title: '地址',
    dataIndex: 'ip_or_domain'
  },
  {
    title: '端口',
    dataIndex: 'port'
  },
  {
    title: '日志上报',
    dataIndex: 'is_sampled_log_upload',
    scopedSlots: { customRender: 'isSampledLogUpload' }
  },
  {
    title: '上次同步状态',
    dataIndex: 'last_sync_status',
    scopedSlots: { customRender: 'lastSyncStatus' }
  },
  {
    title: '上次同步时间',
    dataIndex: 'last_sync_at',
    scopedSlots: { customRender: 'lastSyncAt' }
  },
  {
    title: '操作',
    dataIndex: 'operation',
    scopedSlots: { customRender: 'operation' }
  }
]

export default {
  data() {
    return {
      columns,
      list: [],
      visible: false,
      operateType: 'add',
      form: {
        name: '',
        ip_or_domain: '',
        port: 8083
      },
      rules: {
        ip_or_domain: [
          { required: true, message: '请输入节点IP或主机名', trigger: 'blur' }
          // {
          //   // eslint-disable-next-line
          //   // pattern: /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/,
          //   message: 'IP地址无效',
          //   trigger: 'blur'
          // }
        ],
        port: [
          { required: true, message: '请输入主机端口', trigger: 'blur' },
          {
            type: 'number',
            min: 1,
            max: 65535,
            message: '端口无效',
            trigger: 'blur'
          }
        ]
      }
    }
  },
  methods: {
    onAdd() {
      // this.$refs.form.resetFields()
      this.visible = true
    },
    onOK() {
      const _this = this
      _this.$refs.form.validate(valid => {
        if (valid) {
          switch (_this.operateType) {
            case 'add':
              _this.insert(_this.form)
              break
            case 'edit':
              break
          }
        }
      })
    },
    onCancel() {},
    onSync(id) {
      SyncCfg(id)
        .then(res => {
          if (res.code == 0) {
            this.getList()
            this.$message.success('任务添加成功！')
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(err => {
          this.$message.error('网络异常，请稍候再试')
        })
    },

    onClickMore(key, id) {
      if (key == 1) {
        this.doSwitchLogUploadStatus(id)
      }
      if (key == 2) {
        this.doDelete(id)
      }
    },

    doDelete(id) {
      const _this = this
      _this.$confirm({
        title: `确认删除？`,
        okText: '确认',
        // content: h => (
        //   <div>
        //     <span>请确认是否删除</span>
        //   </div>
        // ),
        onOk() {
          DelNode(id)
            .then(res => {
              if (res.code == 0) {
                _this.$message.success('删除成功！')
                _this.getList()
              } else {
                _this.$message.error(res.msg)
              }
            })
            .catch(err => {
              _this.$message.error('网络异常')
            })
        },
        onCancel() {}
      })
    },
    doSwitchLogUploadStatus(id) {
      SwitchLogUpload(id)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('任务添加成功！')
            this.getList()
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(err => {
          this.$message.error('网络异常，请稍候再试')
        })
    },

    insert(data) {
      const _this = this
      AddNode(data)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('添加成功！')
            _this.$refs.form.resetFields()
            _this.visible = false
            _this.getList()
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {
          _this.$message.error('网络异常')
        })
    },
    getList() {
      ListNodes(this.queryParams)
        .then(res => {
          this.list = res.data.list || []
          this.meta = res.data.meta
        })
        .catch(err => {
          this.$message.error('网络异常，获取节点列表失败')
        })
    }
  },
  mounted() {
    this.getList()
  }
}
</script>
