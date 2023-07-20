<template>
  <page-header-wrapper>
    <template v-slot:extra>
      <a-button type="primary" @click="showModal">新增URL白名单</a-button>
    </template>

    <a-card>
      <div class="table-page-search-wrapper">
        <a-form-model layout="inline" labelAlign="right">
          <a-row>
            <a-col :span="4">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.site_id" @change="onChangeSite">
                  <a-select-option :value="0">全部</a-select-option>
                  <a-select-option v-for="item in domains" :value="item.id" :key="item.id">{{
                    item.domain
                  }}</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="2" style="margin-left: 10px">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.status" @change="onChangeStatus">
                  <a-select-option value="">全部</a-select-option>
                  <a-select-option :value="0">未启用</a-select-option>
                  <a-select-option :value="1">已启用</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="6" style="margin-left: 10px">
              <a-form-model-item label="">
                <a-input-search placeholder="请输入URL" v-model="queryParams.url" enter-button @search="onSearch" />
              </a-form-model-item>
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
        :row-selection="{ selectedRowKeys: selectedRowKeys, onChange: onSelectChange }"
      >
        <template slot="match_mode" slot-scope="text">
          <template v-if="text === 1">
            <a-tag color="green">
              前缀
            </a-tag>
          </template>
          <template v-else>
            <a-tag color="blue">
              精准
            </a-tag>
          </template>
        </template>
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
          <a-button type="link" @click="updateItemStatus(record)">{{ record.status === 1 ? '停用' : '启用' }}</a-button>
          <a-button type="link" size="small" @click="updateItem(record)">编辑</a-button>
          <a-button type="link" size="small" @click="deleteItem(record)">删除</a-button>
        </template>
      </a-table>

      <a-row :style="{ marginTop: '10px' }" v-if="meta.total">
        <!-- 批量操作 -->
        <a-col :span="4">
          <a-space>
            <a-button :disabled="selectedRowKeys.length == 0" @click="onBatchDel">删除</a-button>
            <!-- <a-button @click="onBatchAdd">批量添加</a-button> -->
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
    <a-modal
      :width="700"
      v-model="visible"
      :title="(operateType == 'add' ? '添加' : '编辑') + 'URL白名单'"
      @ok="handleOk"
      @cancel="onCancel"
      :afterClose="
        () => {
          this.$refs.form.resetFields()
        }
      "
    >
      <a-form-model ref="form" :model="form" :rules="rules" :label-col="{ span: 5 }" :wrapper-col="{ span: 15 }">
        <!-- <a-form-model-item label="ID" v-if="operateType == 'edit'" style="display:none;">
          <a-input :disabled="operateType == 'edit'"></a-input>
        </a-form-model-item> -->
        <a-form-model-item label="网站" prop="domain">
          <a-select v-model="form.site_id" placeholder="请选择网站">
            <a-select-option v-for="item in domains" :value="item.id" :key="item.id">{{ item.domain }}</a-select-option>
          </a-select>
        </a-form-model-item>
        <a-form-model-item label="URL" prop="path">
          <a-input placeholder="请输入URL" v-model="form.path"></a-input>
        </a-form-model-item>
        <a-form-model-item label="匹配模式">
          <a-radio-group v-model="form.match_mode">
            <a-radio :value="1">前缀</a-radio>
            <a-radio :value="2">精准</a-radio>
          </a-radio-group>
        </a-form-model-item>
        <!-- <a-form-model-item label="Method">
          <a-select plceholder="请选择">
            <a-select-option v-for="item in methods" :value="item" :key="item">{{ item }}</a-select-option>
          </a-select>
        </a-form-model-item> -->
        <a-form-model-item label="添加备注" prop="remark">
          <a-textarea :rows="5" placeholder="请输入备注" v-model="form.remark"></a-textarea>
        </a-form-model-item>
      </a-form-model>
    </a-modal>
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { GetURLList, AddURL, UpdateURL, UpdateURLStatus, DeleteURL } from '@/api/whitelist/url'
import { ListSite } from '@/api/site'

const columns = [
  {
    title: '网站',
    dataIndex: 'host'
  },
  {
    title: '路径',
    dataIndex: 'path',
    width: 300
  },
  {
    title: '匹配模式',
    dataIndex: 'match_mode',
    scopedSlots: { customRender: 'match_mode' }
  },
  // {
  //   title: '方法',
  //   dataIndex: 'method'
  // },
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
        page_num: 1,
        page_size: 10,
        site_id: 0,
        status: '',
        url: ''
      },
      operateType: '',
      visible: false,
      domains: [],
      form: {
        id: null,
        site_id: undefined,
        path: '',
        match_mode: 1,
        remark: ''
      },
      rules: {},
      selectedRowKeys: []
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

    // 选中第几页
    onShowSizeChange(current, pageSize) {
      this.queryParams.page_size = pageSize
      this.queryParams.page_num = 1
      this.getList()
    },

    // 跳转到第几页
    onChange(page, pageSize) {
      this.queryParams.page_num = page
      this.getList()
    },

    onCancel() {
      this.$refs.form.resetFields()
    },

    onSearch() {
      this.queryParams.page_num = 1
      this.getList()
    },

    onSelectChange(selectedRowKeys) {
      this.selectedRowKeys = selectedRowKeys
    },

    onBatchDel() {
      let _this = this
      this.$confirm({
        title: '确定要删除所选的' + _this.selectedRowKeys.length + '个URL白名单规则？',
        onOk() {
          _this.selectedRowKeys.forEach(item => {
            DeleteURL(item)
              .then(res => {
                if (res.code == 0) {
                  _this.$message.success('删除成功!')
                  _this.getList()
                } else {
                  _this.$message.error(res.msg)
                }
              })
              .catch(err => {
                _this.$message.error(err.message)
              })
          })
          _this.selectedRowKeys = []
        },
        onCancel() {}
      })
    },

    onChangeSite() {
      this.getList()
    },

    onChangeStatus() {
      this.getList()
    },

    updateItem(record) {
      console.log(record)
      this.operateType = 'edit'
      this.visible = true
      this.form.id = record.id
      this.form.site_id = record.site_id
      this.form.path = record.path
      this.form.remark = record.remark
      this.form.match_mode = record.match_mode
    },

    handleOk() {
      this.$refs.form.validate(valid => {
        if (valid) {
          var parsedobj = JSON.parse(JSON.stringify(this.form))
          console.log(parsedobj)
          // this.visible = false
          switch (this.operateType) {
            case 'add':
              this.addUrl(this.form)
              break
            case 'edit':
              this.updateUrl(this.form)
              break
          }
        }
      })

      // e.preventDefault()
      // const _this = this
      // _this.$refs.form.validate((err, values) => {
      //   console.log(values)
      //   switch (_this.operateType) {
      //     case 'add':
      //       console.log(_this.form)
      //       _this.addUrl(values)
      //       break
      //     case 'edit':
      //       _this.updateUrl(values)
      //       break
      //   }
      // })
    },

    updateItemStatus(item) {
      const _this = this
      this.$confirm({
        title: `确定${item.status === 1 ? '停用' : '启用'}？`,
        onOk() {
          UpdateURLStatus(item.id)
            .then(res => {
              if (res.code === 0) {
                _this.$message.success('操作成功!')
                _this.onSearch()
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

    deleteItem(item) {
      this.deleteUrl(item.id)
    },

    getList() {
      GetURLList(this.queryParams)
        .then(res => {
          this.list = res.data.list
          this.meta = res.data.meta
        })
        .catch(() => {
          this.$message.error('网络异常，请稍后再试')
        })
    },

    getAllDomains() {
      ListSite().then(res => {
        this.domains = res.data.list
      })
    },

    addUrl(values) {
      const _this = this
      AddURL(values)
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

    updateUrl(values) {
      const _this = this
      UpdateURL(values.id, values)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('修改成功！')
            _this.getList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {})
    },

    deleteUrl(id) {
      const _this = this
      DeleteURL(id)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('删除成功！')
            _this.getList()
            _this.visible = false
          }
        })
        .catch(err => {})
    }
  },

  beforeCreate() {
    // this.form = this.$form.createForm(this, { name: 'form' })
  },

  mounted() {
    this.getList()
    this.getAllDomains()
  }
}
</script>
