<template>
  <page-header-wrapper>
    <a-card title="自定义规则（全局有效）">
      <!-- 添加按钮 -->
      <a-button slot="extra" type="primary" @click="add">添加</a-button>

      <!-- 表格 -->
      <a-table
        :columns="columns"
        :dataSource="list"
        :rowKey="record => record.id"
        :pagination="false"
      ></a-table>

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
      <!-- 分页结束 -->
    </a-card>

    <!-- 新增 Modal -->
    <a-modal
      :width="700"
      v-model="visible"
      :title="modalType == 'create' ? '添加':'编辑' + '证书'"
      @ok="handleOk"
    >
      <!-- form -->
      <a-form :form="form" :label-col="{ span: 5 }" :wrapper-col="{ span: 15 }">
        <a-form-item label="ID" v-if="modalType == 'edit'">
          <a-input :disabled="modalType == 'edit'" v-decorator="['id',{}]"></a-input>
        </a-form-item>
        <a-form-item label="添加备注">
          <a-textarea
            placeholder="请输入备注"
            v-decorator="[
              'remark',
              {
                rules: [{ required: true, message: '请输入备注!' }]
              }
            ]"
          ></a-textarea>
        </a-form-item>
      </a-form>
      <!-- from END -->
    </a-modal>
    <!-- Modal END -->
  </page-header-wrapper>
</template>
<script>
const columns = [
  {
    title: '序号',
    dataIndex: 'index'
  }
]
export default {
  data () {
    return {
      columns,
      list: [],
      meta: {},
      queryParams: {
        page_num: 1,
        page_size: 10
      },
      visible: false,
      modalType: null
    }
  },

  methods: {
    add () {
      this.visible = true
    },

    onShowSizeChange () {

    },

    onChange () {

    },

    getList () {

    }

  },
  beforeCreate () {
    this.form = this.$form.createForm(this, { name: 'form' })
  },
  mounted () {
    this.getList()
  }
}
</script>
