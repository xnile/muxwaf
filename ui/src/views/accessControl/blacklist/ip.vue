<template>
  <page-header-wrapper>
    <template v-slot:extra>
      <a-button type="primary" @click="showModal">添加IP黑名单</a-button>
    </template>

    <a-card>
      <div class="table-page-search-wrapper">
        <a-form layout="inline" labelAlign="right">
          <a-row>
            <a-col :span="5">
              <a-form-item label="">
                <a-range-picker :showTime="showTime" v-model="time" @change="onTimeChange" />
              </a-form-item>
            </a-col>
            <a-col :span="2" style="margin-left: 5px">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.status" @change="onSearchStatusChange">
                  <a-select-option value="">全部</a-select-option>
                  <a-select-option :value="0">未启用</a-select-option>
                  <a-select-option :value="1">已启用</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="4" style="margin-left: 8px">
              <a-form-item label="">
                <a-input placeholder="请输入IP或CIDR" style="width: 100%" v-model="queryParams.ip"></a-input>
              </a-form-item>
            </a-col>
            <a-col :span="3" style="margin-left: 8px">
              <a-space>
                <a-button type="primary" @click="onQuery">查询</a-button>
                <a-button type="primary" @click="onCheckIfExist">检测IP是否已经在库中</a-button>
                <!-- <a-button @click="onReset">重置</a-button> -->
              </a-space>
            </a-col>
          </a-row>
        </a-form>
      </div>

      <a-table
        :columns="columns"
        :data-source="list"
        :scroll="{ x: 1300 }"
        :rowKey="record => record.id"
        :pagination="false"
        :row-selection="{ selectedRowKeys: selectedRowKeys, onChange: onSelectChange }"
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
        <template slot="action" slot-scope="text, record">
          <a-button type="link" @click="updateItemStatus(record)">{{ record.status === 1 ? '停用' : '启用' }}</a-button>
          <a-button type="link" size="small" @click="updateItem(record)">编辑</a-button>
          <a-button type="link" size="small" @click="deleteItem(record.id)">删除</a-button>
        </template>
      </a-table>

      <a-row :style="{ marginTop: '10px' }" v-if="meta.total">
        <a-col :span="4">
          <a-space>
            <a-button :disabled="selectedRowKeys.length == 0" @click="onBatchDel">删除</a-button>
            <a-button @click="onBatchAdd">批量添加</a-button>
          </a-space>
        </a-col>
        <a-col :span="20">
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
    <!-- <a-modal v-model="visible" :title="(modalType == 'create' ? '添加' : '编辑') + 'IP黑名单'" @ok="handleItem">
      <a-form :form="form" :label-col="{ span: 5 }" :wrapper-col="{ span: 15 }">
        <a-form-item label="ID" v-if="modalType == 'edit'" style="display: none">
          <a-input :disabled="modalType == 'edit'" v-decorator="['id', {}]"></a-input>
        </a-form-item>
        <a-form-item label="IP" prop="ip">
          <a-input
            placeholder="请输入IP"
            v-decorator="[
              'ip',
              {
                rules: [{ required: true, message: '请输入IP!' }]
              }
            ]"
          ></a-input>
        </a-form-item>
        <a-form-item label="备注">
          <a-textarea
            placeholder="请输入备注"
            v-decorator="[
              'remark',
              {
                rules: [{ required: false, message: '请输入备注!' }]
              }
            ]"
          ></a-textarea>
        </a-form-item>
      </a-form>
    </a-modal> -->
    <a-modal
      :width="700"
      v-model="visible"
      :title="operateType == 'add' ? '添加IP黑名单' : '修改备注'"
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

    <!-- 批量添加 -->
    <a-modal
      :width="700"
      v-model="batchAddVisible"
      title="批量添加IP黑名单"
      @ok="onBatchAddOK"
      @cancel="onBatchAddCancel"
    >
      <a-form-model
        ref="batchAddForm"
        :model="batchAddForm"
        :rules="batchAddRules"
        :label-col="{ span: 5 }"
        :wrapper-col="{ span: 15 }"
      >
        <a-form-model-item label="IP" prop="ipList">
          <a-textarea :rows="5" placeholder="请输入IP" v-model="batchAddForm.ipList"></a-textarea>
        </a-form-model-item>
        <a-form-model-item label="备注" prop="remark">
          <a-textarea :rows="1" placeholder="请输入备注" v-model="batchAddForm.remark"></a-textarea>
        </a-form-model-item>
      </a-form-model>
    </a-modal>
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import {
  ListBlacklistIP,
  InsertBlacklistIP,
  UpdateBlacklistIP,
  DeleteBlacklistIP,
  UpdateBlacklistIPStatus,
  IsIncluded,
  BatchAdd
} from '@/api/blacklist/ip'

const columns = [
  {
    title: 'IP',
    key: 'ip',
    dataIndex: 'ip'
  },
  {
    title: '状态',
    key: 'status',
    dataIndex: 'status',
    scopedSlots: { customRender: 'status' }
  },
  {
    title: '描述',
    key: 'remark',
    dataIndex: 'remark'
  },
  {
    title: '创建时间',
    key: 'created_at',
    dataIndex: 'created_at',
    scopedSlots: { customRender: 'created_at' }
  },
  {
    title: '操作',
    key: 'action',
    scopedSlots: { customRender: 'action' }
  }
]
export default {
  data() {
    return {
      modalType: null,
      operateType: 'add',
      time: ['', ''],
      columns,
      list: [],
      meta: {},
      visible: false,
      queryParams: {
        start_time: '',
        end_time: '',
        page_num: 1,
        page_size: 10,
        ip: '',
        status: ''
      },
      form: {
        ip: '',
        remark: ''
      },
      rules: { ip: [{ required: true, message: '请输入IP', trigger: 'blur' }] },
      keyword: '',
      showTime: {
        defaultValue: [moment('00:00:00', 'HH:mm:ss'), moment('23:59:59', 'HH:mm:ss')]
      },
      selectedRowKeys: [],

      batchAddVisible: false,
      batchAddForm: {
        ipList: '',
        remark: ''
      },
      batchAddRules: {}
    }
  },
  methods: {
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
    onSearchStatusChange() {
      // console.log(this.queryParams.status)
      this.doGetList()
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
              this.$message.success('已存在')
            } else {
              this.$message.warning('未找到')
            }
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(err => {
          this.$message.error(err.msg)
        })
    },
    // prettier-ignore
    onReset() {
      this.time = ['', ''],
      this.queryParams.page_num = 1,
      this.queryParams = {
        page_num: 1,
        page_size: 10,
        ip: '',
        start_time: '',
        end_time: ''
      }
      this.doGetList()
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

    onSelectChange(selectedRowKeys) {
      // console.log('selectedRowKeys changed: ', selectedRowKeys)
      this.selectedRowKeys = selectedRowKeys
    },

    onBatchAdd() {
      this.batchAddVisible = true
    },
    onBatchAddOK() {
      let data = {
        ip_list: this.batchAddForm.ipList.replace(/^\s*$(?:\r\n?|\n)/gm, '').split(/,|\n|\r\n/),
        remark: this.batchAddForm.remark
      }
      // let ipList = this.batchAddForm.ipList.replace(/^\s*$(?:\r\n?|\n)/gm, '').split(/,|\n|\r\n/)

      BatchAdd(data)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('添加成功')
            this.batchAddVisible = false
            this.doGetList()
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(err => {
          this.$message.error(err.message)
        })
    },
    onBatchAddCancel() {},

    onBatchDel() {
      let _this = this
      this.$confirm({
        title: '确定要删除所选的' + _this.selectedRowKeys.length + '个IP黑名单',
        onOk() {
          _this.selectedRowKeys.forEach(item => {
            DeleteBlacklistIP(item)
              .then(res => {
                if (res.code == 0) {
                  _this.$message.success('删除成功!')
                  _this.doGetList()
                } else {
                  _this.$message.error(res.msg)
                }
              })
              .catch(err => {
                _this.$message.error(err.message)
              })
          })
          _this.selectedRowKeys = []
          // _this.doGetList()
        },
        onCancel() {}
      })
    },

    deleteItem(id) {
      const _this = this
      this.$confirm({
        title: '确定删除？',
        onOk() {
          DeleteBlacklistIP(id).then(res => {
            if (res.code === 0) {
              _this.$message.success('操作成功!')
              _this.doGetList()
            } else {
              _this.$message.error(res.msg)
            }
          })
        },
        onCancel() {
          // console.log('Cancel')
        }
      })
    },

    showModal() {
      this.operateType = 'add'
      this.visible = true
      this.$nextTick(() => {
        this.$refs.form && this.$refs.form.resetFields()
      })
    },

    onCancel() {
      this.$refs.form.resetFields()
    },

    // onOk() {
    //   if (this.modalType == 'create') {
    //     this.form.validateFields((err, values) => {
    //       if (!err) {
    //         InsertBlacklistIP(values)
    //           .then(res => {
    //             // console.log(res)
    //             if (res.code === 0) {
    //               this.visible = false
    //               this.doGetList()
    //             } else {
    //               this.$message.error(res.msg)
    //             }
    //           })
    //           .catch(err => {
    //             this.$message.error(err.message)
    //           })
    //       }
    //     })
    //   } else {
    //     this.form.validateFields((err, values) => {
    //       if (!err) {
    //         const id = values.id
    //         delete values.id
    //         UpdateBlacklistIP(id, values)
    //           .then(res => {
    //             if (res.code === 0) {
    //               this.$message.success('修改成功！')
    //               this.visible = false
    //               this.doGetList()
    //             } else {
    //               this.$message.error(res.msg)
    //             }
    //           })
    //           .catch(() => {
    //             this.$message.error('网络异常，请稍后再试')
    //           })
    //       }
    //     })
    //   }
    // },
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

    updateItem(record) {
      this.operateType = 'edit'
      this.visible = true
      this.$nextTick(() => {
        this.form.id = record.id
        this.form.ip = record.ip
        this.form.remark = record.remark
      })
    },

    doAdd(values) {
      const _this = this
      InsertBlacklistIP(values)
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
      UpdateBlacklistIP(id, values)
        .then(res => {
          if (res.code == 0) {
            _this.visible = false
            _this.$message.success('修改成功！')
            _this.doGetList()
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {})
    },

    updateItemStatus(item) {
      const _this = this
      this.$confirm({
        title: `确定${item.status === 1 ? '停用' : '启用'}？`,
        onOk() {
          UpdateBlacklistIPStatus(item.id)
            .then(res => {
              if (res.code === 0) {
                _this.$message.success('操作成功!')
                _this.doGetList()
              } else {
                _this.$message.error(res.msg)
              }
            })
            .catch(err => {
              _this.$message.error(err.message)
            })
        },
        onCancel() {}
      })
    },
    doGetList() {
      ListBlacklistIP(this.queryParams)
        .then(res => {
          this.list = res.data.list
          this.meta = res.data.meta
        })
        .catch(() => {
          this.$message.error('网络异常，请稍后再试')
        })
    }
  },
  beforeCreate() {
    this.form = this.$form.createForm(this, { name: 'form' })
  },
  mounted() {
    this.doGetList()
  }
}
</script>

<style></style>
