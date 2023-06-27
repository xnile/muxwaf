<template>
  <page-header-wrapper>
    <!-- 添加按钮 -->
    <template v-slot:extra>
      <a-button type="primary" @click="add">添加证书</a-button>
    </template>

    <a-card title="证书管理">
      <!-- 表格 -->
      <a-table :columns="columns" :dataSource="list" :rowKey="record => record.id" :pagination="false">
        <span slot="end_time" slot-scope="text">{{ text | moment }}</span>
        <template slot="sans" slot-scope="items">
          <div v-for="item in items" :key="item">{{ item }}</div>
        </template>
        <template slot="sites" slot-scope="items">
          <div v-for="item in items" :key="item.id">{{ item.domain }}</div>
          <div v-if="items.length == 0">--</div>
        </template>
        <template slot="operation" slot-scope="text, record">
          <a-button type="link" size="small" @click="onUpdateItem(record)">更新</a-button>
          <a-button type="link" size="small" @click="onDelItem(record)">删除</a-button>
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

    <!-- 新增 Modal -->
    <a-modal :width="700" v-model="visible" :title="(operateType == 'add' ? '添加' : '更新') + '证书'" @ok="handleOk">
      <!-- form -->
      <a-form-model
        ref="form"
        labelAlign="left"
        :model="form"
        :rules="rules"
        :label-col="{ span: 5 }"
        :wrapper-col="{ span: 15 }"
        layout="vertical"
      >
        <a-form-model-item label="证书名称" prop="name">
          <a-input placeholder="请输入证书名称" v-model="form.name"></a-input>
        </a-form-model-item>
        <a-form-model-item label="证书文件" prop="cert">
          <a-textarea :rows="7" :placeholder="placeholderCert" v-model="form.cert"></a-textarea>
        </a-form-model-item>
        <a-form-model-item label="证书私钥" prop="key">
          <a-textarea :rows="7" :placeholder="placeholderKey" v-model="form.key"></a-textarea>
        </a-form-model-item>
        <!-- </a-form> -->
      </a-form-model>
      <!-- form END -->
    </a-modal>
    <!-- MoDal END -->
  </page-header-wrapper>
</template>

<script>
import { AddCert, DelCert, ListCert, UpdateCert } from '@/api/certificate'
import moment from 'moment'

const columns = [
  {
    title: '证书名称',
    dataIndex: 'name'
  },
  {
    title: '域名',
    dataIndex: 'sans',
    scopedSlots: { customRender: 'sans' }
  },
  {
    title: '证书品牌',
    dataIndex: 'cn'
  },
  {
    title: '绑定站点',
    dataIndex: 'sites',
    scopedSlots: { customRender: 'sites' }
  },
  {
    title: '到期时间',
    dataIndex: 'end_time',
    scopedSlots: { customRender: 'end_time' }
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
      placeholderCert: '证书格式以"-----BEGIN CERTIFICATE-----"开头，以"-----END CERTIFICATE-----"结尾。',
      placeholderKey:
        '证书私钥格式以"-----BEGIN (RSA|EC) PRIVATE KEY-----"开头，以"-----END(RSA|EC) PRIVATE KEY-----"结尾。',
      columns,
      list: [],
      meta: {},
      queryParams: {
        page_num: 1,
        page_size: 10
      },
      visible: false,
      operateType: '',
      form: {
        id: 0,
        name: '',
        cert: '',
        key: ''
      },
      rules: {}
    }
  },

  methods: {
    add() {
      this.visible = true
      this.operateType = 'add'
      this.$nextTick(() => {
        this.$refs.form && this.$refs.form.resetFields()
      })
    },

    onUpdateItem(record) {
      this.visible = true
      this.operateType = 'edit'
      this.form.name = record.name
      this.form.id = record.id
    },

    onDelItem(record) {
      const _this = this
      _this.$confirm({
        title: `确定删除`,
        okText: '确认风险并删除',
        content: h => (
          <div>
            <span>删除证书将可能面临业务中断的风险，请预先排查该证书是否有站点在使用。</span>
          </div>
        ),
        onOk() {
          _this.delete(record.id)
        },
        onCancel() {}
      })
    },

    onShowSizeChange() {},
    onChange() {},

    handleOk() {
      // e.preventDefault()
      const _this = this
      _this.$refs.form.validate(valid => {
        if (valid) {
          switch (_this.operateType) {
            case 'add':
              // console.log(this.form)
              _this.insert(_this.form)
              break
            case 'edit':
              _this.update(_this.form.id, _this.form)
              break
          }
        }
      })
    },

    insert(data) {
      const _this = this
      AddCert(data)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('添加成功！')
            _this.getList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {
          _this.$message.error(err.msg)
        })
    },

    update(id, data) {
      const _this = this
      UpdateCert(id, data)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('更新成功！')
            _this.getList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {
          _this.$message.error(err.msg)
        })
    },

    delete(id) {
      const _this = this
      DelCert(id)
        .then(res => {
          if (res.code == 0) {
            _this.$message.success('删除成功！')
            _this.getList()
            _this.visible = false
          } else {
            _this.$message.error(res.msg)
          }
        })
        .catch(err => {
          _this.$message.error('网络异常')
        })
    },

    getList() {
      ListCert(this.queryParams).then(res => {
        this.list = res.data.list || []
        this.meta = res.data.meta
      })
    }
  },

  beforeCreate() {
    // this.form = this.$form.createForm(this, { name: 'form' })
  },

  mounted() {
    this.getList()
  }
}
</script>
