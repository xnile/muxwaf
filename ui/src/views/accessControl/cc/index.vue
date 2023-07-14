<template>
  <page-header-wrapper>
    <template v-slot:extra>
      <a-button type="primary" @click="showModal">添加频率限制</a-button>
    </template>

    <a-card>
      <div class="table-page-search-wrapper">
        <a-form-model layout="inline" labelAlign="right">
          <a-row>
            <a-col :span="3">
              <a-form-model-item label="">
                <a-select placeholder="请选择网站" v-model="queryParams.site_id" @change="onSiteChange">
                  <a-select-option :value="0">全部网站</a-select-option>
                  <a-select-option v-for="item in sites" :value="item.id" :key="item.id">{{
                    item.domain
                  }}</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="2" style="margin-left: 10px">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.match_mode" @change="onMatchModeChange">
                  <a-select-option :value="0">全部模式</a-select-option>
                  <a-select-option :value="1">前缀</a-select-option>
                  <a-select-option :value="2">精准</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="2" style="margin-left: 10px">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.status" @change="onStatusChange">
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
      :title="(operateType == 'add' ? '添加' : '编辑') + 'CC防护'"
      @ok="onOk"
      @cancel="onCancel"
    >
      <!-- form -->
      <a-form-model ref="form" :model="form" :label-col="{ span: 4 }" :wrapper-col="{ span: 16 }">
        <a-form-model-item label="网站" prop="site_id">
          <a-select plceholder="请选择网站" v-model="form.site_id">
            <a-select-option v-for="item in sites" :value="item.id" :key="item.id">{{ item.domain }}</a-select-option>
          </a-select>
        </a-form-model-item>
        <a-form-model-item label="路径" prop="path">
          <a-input placeholder="请输入路径" v-model="form.path"></a-input>
        </a-form-model-item>
        <a-form-model-item label="请求次数" prop="limit">
          <a-input-number placeholder="阈值" v-model.number="form.limit"></a-input-number>
        </a-form-model-item>
        <a-form-model-item label="时间(秒)" prop="window">
          <a-input-number placeholder="统计时间" v-model.number="form.window"></a-input-number>
        </a-form-model-item>
        <a-form-model-item label="匹配模式" prop="match_mode">
          <a-radio-group v-model="form.match_mode">
            <a-radio :value="1">前缀</a-radio>
            <a-radio :value="2">精准</a-radio>
          </a-radio-group>
        </a-form-model-item>
        <a-form-model-item label="规则描述" prop="remark">
          <a-textarea :rows="5" placeholder="请输入规则描述" v-model="form.remark"></a-textarea>
        </a-form-model-item>
      </a-form-model>
      <!-- from END -->
    </a-modal>
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { GetList, Add, Update, UpdateStatus, Delete } from '@/api/rateLimit'
import { GetALLSite } from '@/api/site'

const columns = [
  {
    title: '网站',
    dataIndex: 'host',
    width: '10%'
  },
  {
    title: '路径',
    dataIndex: 'path',
    width: '15%'
  },
  {
    title: '匹配模式',
    dataIndex: 'match_mode',
    scopedSlots: { customRender: 'match_mode' },
    width: '8%'
  },
  {
    title: '阈值',
    dataIndex: 'limit',
    width: '8%'
  },
  {
    title: '时长(秒)',
    dataIndex: 'window',
    width: '8%'
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
        page_num: 1,
        page_size: 10,
        site_id: 0,
        url: '',
        match_mode: 0,
        status: ''
      },
      operateType: '',
      visible: false,
      time: ['', ''],
      sites: [],
      form: {
        site_id: null,
        path: '',
        limit: 1,
        window: 60,
        match_mode: 1,
        remark: ''
      },
      rules: {},
      // methods,
      status: 0,
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
      this.$refs.form && this.$refs.form.resetFields()
    },

    onShowSizeChange() {},

    onChange() {},

    // onTimeChange() {},
    onSiteChange() {
      this.doGetList()
    },

    onMatchModeChange() {
      this.doGetList()
    },

    onStatusChange() {
      this.doGetList()
    },

    query() {
      this.queryParams.page_num = 1
      this.doGetList()
    },

    onSearch() {
      this.doGetList()
    },

    onUpdate(record) {
      this.operateType = 'edit'
      this.visible = true

      this.$nextTick(() => {
        this.form = record
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
          UpdateStatus(item.id)
            .then(res => {
              if (res.code === 0) {
                _this.$message.success('操作成功!')
                _this.query()
              } else {
                _this.$message.error(res.msg)
              }
            })
            .catch(err => {
              this.$message.error(err.message)
            })
        },
        onCancel() {}
      })
    },

    doGetList() {
      GetList(this.queryParams)
        .then(res => {
          this.list = res.data.list
          this.meta = res.data.meta
        })
        .catch(err => {
          this.$message.error(err.message)
        })
    },

    doAdd(values) {
      const _this = this
      Add(values)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('添加成功！')
            _this.doGetList()
            _this.$refs.form.resetFields()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {
          _this.$message.error(err.message)
        })
    },

    doUpdate(values) {
      const _this = this
      Update(values.id, values)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('修改成功！')
            _this.doGetList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {
          _this.$message.error(err.message)
        })
    },

    doDelete(id) {
      const _this = this
      _this.$confirm({
        title: '确认删除？',
        onOk() {
          Delete(id)
            .then(res => {
              if (res.code == 0) {
                _this.$message.success('删除成功！')
                _this.doGetList()
                _this.visible = false
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

    doGetAllSite() {
      GetALLSite().then(res => {
        this.sites = res.data
      })
    }
  },

  beforeCreate() {
    // this.form = this.$form.createForm(this, { name: 'form' })
  },

  mounted() {
    this.doGetList()
    this.doGetAllSite()
  }
}
</script>
