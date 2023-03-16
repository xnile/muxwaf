<template>
  <page-header-wrapper>
    <a-card title="阻断管理">
      <!-- 添加按钮 -->
      <a-button slot="extra" type="primary" @click="add">添加</a-button>

      <!-- 表格 -->
      <a-table
        :columns="columns"
        :dataSource="list"
        :rowKey="record => record.id"
        :pagination="false"
      >
        <span slot="started_at" slot-scope="text">{{ text | moment }}</span>
        <span slot="ended_at" slot-scope="text">{{ text | moment }}</span>
        <template slot="status" slot-scope="text">
          <template v-if="text === 1">
            <a-badge status="success" text="已启用" />
          </template>
          <template v-else>
            <a-badge status="error" text="已停用" />
          </template>
        </template>
        <template slot="action" slot-scope="text, record">
          <a-button type="link" @click="updateItemStatus(record)">解除</a-button>
          <a-button type="link" @click="updateItemStatus(record)">拉黑</a-button>
        </template>
      </a-table>

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
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { ListBlock } from '@/api/block'

const columns = [
  {
    title: 'IP',
    dataIndex: 'ip'
  },
  {
    title: '开始时间',
    dataIndex: 'started_at',
    scopedSlots: { customRender: 'started_at' }
  },
  {
    title: '结束时间',
    dataIndex: 'ended_at',
    scopedSlots: { customRender: 'ended_at' }
  },
  {
    title: '状态',
    dataIndex: 'status',
    scopedSlots: { customRender: 'status' }
  },
  {
    title: '阻断原因',
    dataIndex: 'reason'
  },
  {
    title: '操作',
    key: 'action',
    scopedSlots: { customRender: 'action' }
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
      }
    }
  },
  methods: {
    add () {
    },

    getList () {
      ListBlock(this.queryParmas).then(res => {
        this.list = res.data.list
        this.meta = res.data.meta
      })
    }
  },

  mounted () {
    this.getList()
  }
}
</script>
