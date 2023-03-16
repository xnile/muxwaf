<template>
  <a-card>
    <div class="table-page-search-wrapper">
      <a-form layout="inline" :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-row>
          <a-col :span="8">
            <a-form-item label="操作时间">
              <a-range-picker :showTime="showTime" :value="time" @change="onTimeChange" />
            </a-form-item>
          </a-col>
          <a-col :span="7" :offset="1">
            <a-space>
              <a-button type="primary" @click="query">查询</a-button>
              <a-button @click="reset">重置</a-button>
            </a-space>
          </a-col>
        </a-row>
      </a-form>
    </div>
    <a-table
      :pagination="false"
      :columns="columns"
      :rowKey="record => record.id"
      :data-source="data"
      :scroll="{ x: 1300 }"
    >
      <template slot="index" slot-scope="text, record, index">{{ index + 1 }}</template>
      <span slot="loginType" slot-scope="text">{{ typeList[text - 1].labelName }}</span>
      <span slot="createTime" slot-scope="text">{{ text | moment }}</span>
    </a-table>

    <!-- 分页 -->
    <a-row :style="{ paddingTop: '10px' }" v-if="meta.total">
      <a-col :span="24">
        <a-pagination
          style="float:right"
          show-size-changer
          show-quick-jumper
          show-less-items
          :show-total="total => `共 ${total} 条记录 第${meta.pageNum}/${meta.pages}页`"
          :total="meta.total"
          :pageSize="queryParams.pageSize"
          @showSizeChange="onShowSizeChange"
          @change="onChange"
        />
      </a-col>
    </a-row>
  </a-card>
</template>

<script>
import moment from 'moment'
import 'moment/locale/zh-cn'
moment.locale('zh-cn')
const columns = [
  {
    title: '序号',
    dataIndex: 'keyId',
    key: 'keyId',
    scopedSlots: { customRender: 'index' },
    width: 100,
    fixed: 'left'
  },
  {
    title: '登录方式',
    dataIndex: 'loginType',
    scopedSlots: { customRender: 'loginType' }
  },
  {
    title: '来源',
    dataIndex: 'clientName',
    width: 100
  },
  {
    title: '系统',
    dataIndex: 'osName'
  },
  {
    title: '设备',
    dataIndex: 'deviceMfrs'
  },
  {
    title: '客户端版本',
    dataIndex: 'clientVersion'
  },
  {
    title: 'IP',
    dataIndex: 'ip'
  },
  {
    title: '操作时间',
    dataIndex: 'createTime',
    scopedSlots: { customRender: 'createTime' }
  }
]

export default {
  data () {
    return {
      queryParams: {
        startTime: '',
        endTime: '',
        pageNum: 1,
        pageSize: 10
      },
      meta: {},
      total: {},
      columns,
      data: [],
      selectItem: {},
      visible: false,
      typeList: [
        {
          labelName: '手机号+验证码',
          labelId: '1'
        },
        {
          labelName: '微信',
          labelId: '2'
        },
        {
          labelName: 'qq',
          labelId: '3'
        },
        {
          labelName: '微博',
          labelId: '4'
        },
        {
          labelName: '一键登录',
          labelId: '5'
        },
        {
          labelName: '用户名密码',
          labelId: '6'
        }
      ],
      time: null,
      showTime: {
        defaultValue: [moment('00:00:00', 'HH:mm:ss'), moment('23:59:59', 'HH:mm:ss')]
      }
    }
  },
  methods: {
    onTimeChange (date, dateString) {
      this.time = date
      this.queryParams.startTime = date[0].valueOf()
      this.queryParams.endTime = date[1].valueOf()
    },
    onShowSizeChange (current, pageSize) {
      this.queryParams.pageSize = pageSize
      this.queryParams.pageNum = 1
      this.getList()
    },
    onChange (page, pageSize) {
      this.queryParams.pageNum = page
      this.getList()
    },
    query () {
      this.queryParams.pageNum = 1
      this.getList()
    },
    reset () {
      this.queryParams = {
        startTime: null,
        endTime: null,
        keyword: this.$route.query.id,
        pageNum: 1,
        pageSize: 10
      }
      this.time = null
      this.getList()
    },
    getList () {

    },
    mounted () {
      this.queryParams.keyword = this.$route.query.id
      this.$nextTick(() => {
        this.getList()
      })
    }
  }
}
</script>