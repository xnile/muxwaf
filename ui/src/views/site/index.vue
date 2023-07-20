<template>
  <page-header-wrapper>
    <!-- 添加按钮 -->
    <template v-slot:extra>
      <a-button type="primary" @click="add">新增防护网站</a-button>
    </template>
    <a-card>
      <div class="table-page-search-wrapper">
        <a-form-model layout="inline" labelAlign="right">
          <a-row>
            <a-col :span="2">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.status" @change="onStatusChange">
                  <a-select-option value="">全部</a-select-option>
                  <a-select-option :value="0">未启用</a-select-option>
                  <a-select-option :value="1">已启用</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="10" style="margin-left: 10px">
              <a-form-model-item label="">
                <!-- <a-input-group compact>
                  <a-select placeholder="请选择" v-model="queryParams.status">
                    <a-select-option value="">全部</a-select-option>
                    <a-select-option :value="0">未启用</a-select-option>
                    <a-select-option :value="1">已启用</a-select-option>
                  </a-select>
                  <a-input-search
                    placeholder="请输入域名"
                    v-model="queryParams.domain"
                    enter-button
                    @search="onSearch"
                  />
                </a-input-group> -->
                <a-input-group compact>
                  <a-select v-model="queryParams.is_fuzzy">
                    <a-select-option :value="0">
                      精准匹配
                    </a-select-option>
                    <a-select-option :value="1">
                      模糊匹配
                    </a-select-option>
                  </a-select>
                  <!-- <a-input style="width: 50%" default-value="input content" />
                   -->
                  <a-input-search
                    style="width: 50%"
                    placeholder="请输入域名"
                    v-model="queryParams.domain"
                    @search="onSearch"
                  />
                </a-input-group>
              </a-form-model-item>
            </a-col>
          </a-row>
        </a-form-model>
      </div>

      <!-- 表格 -->
      <a-table
        :columns="columns"
        :dataSource="list"
        :rowKey="record => record.id"
        :pagination="false"
        :row-selection="{ selectedRowKeys: selectedRowKeys, onChange: onSelectChange }"
      >
        <span slot="created_at" slot-scope="text">{{ text | moment }}</span>
        <template slot="status" slot-scope="text">
          <template v-if="text === 1">
            <a-badge status="success" text="正常" />
          </template>
          <template v-else>
            <a-badge status="error" text="已停用" />
          </template>
        </template>
        <template slot="https_status" slot-scope="text">
          <template v-if="text === 1">
            <a-badge status="success" text="已启用" />
          </template>
          <template v-else>
            <a-badge status="warning" text="未启用" />
          </template>
        </template>
        <template slot="pre_cdn" slot-scope="text">
          <template v-if="text === 1">
            <span>是</span>
          </template>
          <template v-else>
            <span>否</span>
          </template>
        </template>
        <template slot="operation" slot-scope="text, record">
          <a-button type="link" size="small" @click="onSettings(record.id, record.domain)">管理</a-button>
          <a-dropdown :trigger="['click']">
            <a class="ant-dropdown-link" @click="e => e.preventDefault()">
              更多
              <a-icon type="down" />
            </a>
            <a-menu slot="overlay" @click="handleMenuClick">
              <a-menu-item key="1">
                <span @click="updateStatus(record.id)">{{ record.status === 1 ? '停用' : '启用' }}</span>
              </a-menu-item>
              <a-menu-item key="2">
                <span class="text-danger" @click="deleteItem(record.id)">删除</span>
              </a-menu-item>
            </a-menu>
          </a-dropdown>
        </template>
      </a-table>
      <!-- 表格end -->
      <!-- 分页 -->
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
      <!-- 分页end -->
    </a-card>
    <a-drawer
      :title="(operateType == 'add' ? '添加' : '修改') + '防护网站'"
      :width="900"
      placement="right"
      :closable="false"
      :visible="visible"
      :after-visible-change="afterVisibleChange"
      @close="onClose"
    >
      <a-form-model
        ref="form"
        labelAlign="left"
        :model="form"
        :rules="rules"
        :label-col="labelCol"
        :wrapper-col="wrapperCol"
        layout="vertical"
      >
        <a-form-model-item label="防护域名" prop="domain" :wrapper-col="{ span: 15 }">
          <a-input placeholder="请输入域名" v-model="form.domain" :disabled="operateType == 'edit'" />
        </a-form-model-item>
        <a-form-model-item label="源站地址" prop="origins">
          <a-table
            :columns="srcColumns"
            :dataSource="form.origins"
            :rowKey="
              (record, index) => {
                return index
              }
            "
            :pagination="false"
          >
            <template slot="addr" slot-scope="v, r, index">
              <a-input type="text" placeholder="请输入源站地址（IP/域名）" v-model="form.origins[index].addr" />
            </template>
            <template slot="port" slot-scope="v, r, index">
              <a-input
                type="number"
                min="0"
                max="65535"
                placeholder="1-65535"
                v-model.number="form.origins[index].port"
              />
            </template>
            <template slot="weight" slot-scope="v, r, index">
              <a-input
                type="number"
                min="0"
                max="100"
                placeholder="0-100"
                v-model.number="form.origins[index].weight"
              />
            </template>
            <template slot="operation" slot-scope="v, record, index">
              <a-space>
                <a-button type="link" @click="delSrc(record, index)">删除</a-button>
              </a-space>
            </template>
          </a-table>
          <a-button style="width: 100%; margin: 20px 0" type="dashed" @click="addSrc">+ 新增</a-button>
        </a-form-model-item>
        <a-form-model-item label="源站协议">
          <a-radio-group v-model="form.origin_protocol">
            <a-radio value="http">HTTP</a-radio>
            <a-radio value="https">HTTPS</a-radio>
          </a-radio-group>
        </a-form-model-item>
        <!-- form END -->
      </a-form-model>

      <!-- 取消,确认按钮 -->
      <div
        :style="{
          position: 'absolute',
          right: 0,
          bottom: 0,
          width: '100%',
          borderTop: '1px solid #e9e9e9',
          padding: '10px 16px',
          background: '#fff',
          textAlign: 'right',
          zIndex: 1
        }"
      >
        <a-button :style="{ marginRight: '8px' }" @click="onClose">取消</a-button>
        <a-button type="primary" @click="onSubmit">确认</a-button>
      </div>
      <!-- END -->
    </a-drawer>
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import store from '@/store'
import { AddSite, ListSite, DelSite, UpdateStatus } from '@/api/site'

const columns = [
  {
    title: '防护网站',
    dataIndex: 'domain'
  },
  {
    title: 'HTTPS',
    dataIndex: 'config.is_https',
    scopedSlots: { customRender: 'https_status' }
  },
  {
    title: '前置CDN',
    dataIndex: 'config.is_real_ip_from_header',
    scopedSlots: { customRender: 'pre_cdn' }
  },
  {
    title: '状态',
    dataIndex: 'status',
    scopedSlots: { customRender: 'status' }
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    scopedSlots: { customRender: 'created_at' }
  },
  {
    title: '操作',
    dataIndex: 'operation',
    scopedSlots: { customRender: 'operation' }
  }
]

const srcColumns = [
  // {
  //   title: '序号',
  //   dataIndex: 'index',
  //   scopedSlots: { customRender: 'index' }
  // },
  {
    title: '源站地址',
    dataIndex: 'addr',
    width: '57%',
    scopedSlots: { customRender: 'addr' }
  },
  {
    title: '端口',
    dataIndex: 'port',
    width: '20%',
    scopedSlots: { customRender: 'port' }
  },
  {
    title: '权重',
    dataIndex: 'weight',
    width: '18%',
    scopedSlots: { customRender: 'weight' }
  },
  {
    title: '操作',
    dataIndex: 'action',
    width: '5%',
    scopedSlots: { customRender: 'operation' }
  }
]

export default {
  data() {
    return {
      columns,
      list: [],
      meta: {},
      queryParams: {
        page_num: 1,
        page_size: 10,
        status: '',
        is_fuzzy: 0,
        domain: ''
      },
      operateType: '',
      visible: false,
      visibleHttps: false,
      certificates: [],
      real_ip_header_type: 0,
      form: {
        domain: '',
        origin_protocol: 'http',
        origins: [
          {
            addr: null,
            port: null,
            weight: null
          }
        ]
      },
      rules: {
        // domain: [{ required: true, message: '' }],
        // protocol: [{ required: true, message: '' }]
      },
      labelCol: { span: 3 },
      wrapperCol: { span: 20 },
      srcColumns,
      radioStyle: {
        display: 'block',
        height: '30px',
        lineHeight: '30px'
      },
      selectedRowKeys: []
    }
  },
  methods: {
    onShowSizeChange() {},
    onChange() {},
    handleMenuClick() {},
    afterVisibleChange(val) {},
    showDrawer() {
      this.visible = true
    },
    onClose() {
      this.visible = false
    },

    // prettier-ignore
    onSearch() {
      this.queryParams.page_num = 1,
      this.doGetList()
    },

    onSettings(id, domain) {
      this.$router.push({ path: `/site/${id}/settings` })
      store.commit('SET_DOMAIN', domain)
    },

    onSelectChange(selectedRowKeys) {
      this.selectedRowKeys = selectedRowKeys
    },

    onBatchDel() {
      let _this = this
      this.$confirm({
        title: '确定要删除所选的' + _this.selectedRowKeys.length + '个站点？',
        onOk() {
          _this.selectedRowKeys.forEach(item => {
            DelSite(item)
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
        },
        onCancel() {}
      })
    },

    onStatusChange() {
      this.doGetList()
    },

    add() {
      this.operateType = 'add'
      this.visible = true
      this.$nextTick(() => {
        this.$refs.form.resetFields()
      })
    },

    updateStatus(id) {
      const _this = this
      UpdateStatus(id)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('操作成功')
            _this.doGetList()
          }
        })
        .catch(err => {
          _this.$message.error(err.message)
        })
    },

    deleteItem(id) {
      const _this = this
      _this.$confirm({
        title: `危险操作`,
        okText: '确认风险并删除',
        content: h => (
          <div>
            <span>删除站点将可能面临业务中断的风险，删除前需要先删除站点关联的防护规则。</span>
          </div>
        ),
        onOk() {
          DelSite(id)
            .then(res => {
              if (res.code == 0) {
                _this.$message.success('操作成功')
                _this.doGetList()
              } else {
                _this.$message.error(res.msg)
              }
            })
            .catch(err => {
              _this.$message.error('网络异常请稍后再试')
            })
        },
        onCancel() {}
      })
    },

    onSubmit() {
      console.log('submit!', this.form)
      const _this = this
      _this.$refs.form.validate(valid => {
        if (valid) {
          switch (_this.operateType) {
            case 'add':
              _this.insert(this.form)
              break
            case 'edit':
              _this.doUpdate(this.form)
              break
            default:
              break
          }
        } else {
        }
      })
    },
    addSrc() {
      this.form.origins.push({ host: '', http_port: 80, weight: 100 })
    },
    delSrc(r, i) {
      this.form.origins.splice(i, 1)
    },

    insert(data) {
      const _this = this
      AddSite(data)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('添加成功！')
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

    doGetList() {
      ListSite(this.queryParams)
        .then(res => {
          if (res.code == 0) {
            this.list = res.data.list || []
            this.meta = res.data.meta || {}
          }
        })
        .catch(err => {
          this.$message.error(err.message)
        })
    }
  },

  mounted() {
    this.doGetList()
  }
}
</script>

<style scoped>
.radio-vertical {
  display: 'block';
  height: '30px';
  line-height: '30px';
}
</style>
