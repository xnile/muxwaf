<template>
  <page-header-wrapper>
    <template v-slot:extra>
      <a-button type="primary" @click="showModal">新增IP白名单</a-button>
    </template>

    <a-card>
      <div class="table-page-search-wrapper">
        <a-form-model layout="inline" labelAlign="right">
          <a-row>
            <a-col :span="5">
              <a-form-item label="">
                <a-range-picker :showTime="showTime" v-model="time" @change="onTimeChange" />
              </a-form-item>
            </a-col>
            <a-col :span="2" style="margin-left: 5px">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.status">
                  <a-select-option value="">全部</a-select-option>
                  <a-select-option :value="0">未启用</a-select-option>
                  <a-select-option :value="1">已启用</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="4" style="margin-left: 8px">
              <a-form-model-item label="">
                <a-input placeholder="请输入IP" style="width:100%" v-model="queryParams.ip"></a-input>
              </a-form-model-item>
            </a-col>
            <a-col :span="3" style="margin-left: 8px">
              <a-space>
                <a-button type="primary" @click="onQuery">查询</a-button>
                <a-button type="primary" @click="onCheckIfExist">检测IP是否已经在库中</a-button>
              </a-space>
            </a-col>
          </a-row>
        </a-form-model>
      </div>

      <a-table
        :columns="columns"
        :data-source="list"
        :scroll="{ x: 1300 }"
        :rowKey="record => record.id"
        :pagination="false"
      >
        <template slot="index" slot-scope="text, record, index">{{ index + 1 }}</template>
        <span slot="created_at" slot-scope="text">{{ text | moment }}</span>
        <template slot="status" slot-scope="text">
          <template v-if="text === 1">
            <a-badge status="success" text="已启用" />
          </template>
          <template v-else>
            <a-badge status="error" text="已停用" />
          </template>
        </template>
        <template slot="operation" slot-scope="text, record">
          <a-button type="link" @click="onUpdateStatus(record)">{{ record.status === 1 ? '停用' : '启用' }}</a-button>
          <a-button type="link" size="small" @click="onUpdate(record)">编辑</a-button>
          <a-button type="link" size="small" @click="onDelete(record)">删除</a-button>
        </template>
      </a-table>

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
    </a-card>

    <!-- 添加对话框 -->
    <a-modal
      :width="700"
      v-model="visible"
      :title="operateType == 'add' ? '添加IP白名单' : '修改备注'"
      @ok="onOk"
      @cancel="onCancel"
    >
      <a-form-model ref="form" :model="form" :rules="rules" :label-col="{ span: 5 }" :wrapper-col="{ span: 15 }">
        <!-- <a-form-model-item label="ID" v-if="operateType == 'edit'" style="display:none;">
          <a-input :disabled="operateType == 'edit'"></a-input>
        </a-form-model-item> -->
        <a-form-model-item label="IP" prop="ip" v-show="operateType == 'add'">
          <a-input placeholder="请输入IP" v-model="form.ip"></a-input>
        </a-form-model-item>
        <a-form-model-item label="备注" prop="remark">
          <a-textarea :rows="5" placeholder="请输入备注" v-model="form.remark"></a-textarea>
        </a-form-model-item>
      </a-form-model>
    </a-modal>
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { GetIPList, AddIP, UpdateIP, UpdateIPStatus, DeleteIP, IsIncluded } from '@/api/whitelist/ip'

const columns = [
  {
    title: 'IP',
    dataIndex: 'ip'
  },
  {
    title: '状态',
    dataIndex: 'status',
    scopedSlots: { customRender: 'status' }
  },
  {
    title: '添加时间',
    dataIndex: 'created_at',
    scopedSlots: { customRender: 'created_at' }
  },
  {
    title: '备注',
    dataIndex: 'remark'
  },
  {
    title: '操作',
    dataIndex: 'operation',
    scopedSlots: { customRender: 'operation' }
  }
]

// const methods = ['GET', 'POST', 'PUT', 'DELETE', 'HEAD']

export default {
  data() {
    return {
      list: [],
      meta: {},
      columns,
      queryParams: {
        start_time: '',
        end_time: '',
        page_num: 1,
        page_size: 10,
        status: '',
        ip: ''
      },
      operateType: '',
      visible: false,
      form: {
        ip: '',
        remark: ''
      },
      rules: {},
      time: ['', ''],
      showTime: {
        defaultValue: [moment('00:00:00', 'HH:mm:ss'), moment('23:59:59', 'HH:mm:ss')]
      }
    }
  },
  methods: {
    showModal() {
      this.visible = true
      this.operateType = 'add'
      this.$nextTick(() => {
        this.$refs.form && this.$refs.form.resetFields()
      })
    },

    onCancel() {
      this.$refs.form.resetFields()
    },

    // 选中第几页
    onShowSizeChange(current, pageSize) {
      this.queryParams.page_size = pageSize
      this.queryParams.page_num = 1
      this.doGetList()
    },

    // 跳转到第几页
    onChange(page, pageSize) {
      this.queryParams.page_num = page
      this.doGetList()
    },

    onTimeChange(date, dateString) {
      this.time = date
      if (date.length === 0) {
        this.queryParams.start_time = ''
        this.queryParams.end_time = ''
        return
      }
      this.queryParams.start_time = date[0].valueOf() / 1000
      this.queryParams.end_time = date[1].valueOf() / 1000
    },

    onQuery() {
      this.queryParams.page_num = 1
      this.doGetList()
    },

    onCheckIfExist() {
      IsIncluded(this.queryParams.ip)
        .then(res => {
          if (res.code === 0) {
            if (res.data) {
              this.$message.success('IP已存在')
            } else {
              this.$message.warning('IP未找到')
            }
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(err => {
          this.$message.error('网络异常，请稍后再试')
        })
    },

    onUpdate(record) {
      this.operateType = 'edit'
      this.visible = true
      this.$nextTick(() => {
        this.form.id = record.id
        this.form.ip = record.ip
        this.form.remark = record.remark
      })
    },

    onOk() {
      this.$refs.form.validate(valid => {
        if (valid) {
          switch (this.operateType) {
            case 'add':
              this.doAdd(this.form)
              break
            case 'edit':
              this.doUpdate(this.form)
              break
          }
        }
      })
    },

    onDelete(item) {
      this.doDelete(item.id)
    },

    onUpdateStatus(item) {
      const _this = this
      this.$confirm({
        title: `确定${item.status === 1 ? '停用' : '启用'}？`,
        onOk() {
          UpdateIPStatus(item.id)
            .then(res => {
              if (res.code === 0) {
                _this.$message.success('操作成功!')
                _this.doGetList()
              } else {
                _this.$message.error(res.msg)
              }
            })
            .catch(err => {
              _this.$message.error(err.msg)
            })
        },
        onCancel() {}
      })
    },

    doGetList() {
      GetIPList(this.queryParams)
        .then(res => {
          this.list = res.data.list
          this.meta = res.data.meta
        })
        .catch(() => {
          this.$message.error('网络异常，请稍后再试')
        })
    },

    doAdd(values) {
      const _this = this
      AddIP(values)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('添加成功！')
            _this.doGetList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {})
    },

    doUpdate(values) {
      const _this = this
      const id = values.id
      delete values.id
      UpdateIP(id, values)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('修改成功！')
            _this.doGetList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {})
    },

    doDelete(id) {
      const _this = this
      _this.$confirm({
        title: '确定删除？',
        onOk() {
          DeleteIP(id)
            .then(res => {
              if (res.code == 0) {
                _this.$message.success('删除成功！')
                _this.doGetList()
                _this.visible = false
              } else {
                _this.$message.error(res.msg)
              }
            })
            .catch(err => {})
        },
        onCancel() {}
      })
    }
  },

  beforeCreate() {
    // this.form = this.$form.createForm(this, { name: 'form' })
  },

  mounted() {
    this.doGetList()
  }
}
</script>
