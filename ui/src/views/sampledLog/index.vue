<template>
  <page-header-wrapper>
    <a-card>
      <div class="table-page-search-wrapper">
        <a-form-model layout="inline" labelAlign="right">
          <a-row>
            <a-col :span="5">
              <a-form-item label="">
                <a-range-picker :showTime="showTime" v-model="time" @change="onTimeChange" />
              </a-form-item>
            </a-col>
            <a-col :span="3" style="margin-left: 3px">
              <a-form-model-item label="">
                <a-select placeholder="请选择网站" v-model="queryParams.site_id">
                  <a-select-option :value="0">全部网站</a-select-option>
                  <a-select-option v-for="item in sites" :value="item.id" :key="item.id">{{
                    item.domain
                  }}</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="2" style="margin-left: 5px">
              <a-form-model-item label="">
                <a-select placeholder="请选择" v-model="queryParams.action">
                  <a-select-option value="">全部动作</a-select-option>
                  <a-select-option :value="1">放行</a-select-option>
                  <a-select-option :value="2">拦截</a-select-option>
                </a-select>
              </a-form-model-item>
            </a-col>
            <a-col :span="8" style="margin-left: 5px">
              <a-form-model-item label="">
                <a-input
                  placeholder="请输入内容，可以是IP、请求ID、URL"
                  style="width:100%"
                  v-model="queryParams.content"
                ></a-input>
              </a-form-model-item>
            </a-col>
            <a-col :span="3" style="margin-left: 8px">
              <a-space>
                <a-button type="primary" @click="onSearch">查询</a-button>
                <a-button type="primary" @click="onReset">重置</a-button>
              </a-space>
            </a-col>
          </a-row>
        </a-form-model>
      </div>
      <a-table :columns="columns" :dataSource="list" :rowKey="record => record.id" :pagination="false">
        <span slot="time" slot-scope="text">{{ text | moment }}</span>
        <template slot="action" slot-scope="text">
          <template v-if="text === 1">
            <a-badge status="success" text="放行" />
          </template>
          <template v-else>
            <a-badge status="error" text="拦截" />
          </template>
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
    </a-card>

    <!-- 详情 -->
    <!-- <a-drawer
      title="详情"
      :width="700"
      placement="right"
      :closable="false"
      :visible="detailVisible"
      :after-visible-change="afterVisibleChange"
      @close="onClose"
    >
      <a-descriptions :column="1" title="攻击日志详情">
        <a-descriptions-item label="UserName">Zhou Maomao</a-descriptions-item>
        <a-descriptions-item label="Telephone">1810000000</a-descriptions-item>
        <a-descriptions-item label="Live">Hangzhou, Zhejiang</a-descriptions-item>
        <a-descriptions-item label="Remark">empty</a-descriptions-item>
        <a-descriptions-item label="Address"
          >No. 18, Wantang Road, Xihu District, Hangzhou, Zhejiang, China</a-descriptions-item
        >
      </a-descriptions>
    </a-drawer> -->
    <!-- 详情END -->
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { ListAttackLog } from '@/api/sampledLog'
import { ListSite } from '@/api/site'

const columns = [
  {
    title: '网站',
    dataIndex: 'host',
    width: '10%'
  },
  {
    title: '路径',
    dataIndex: 'request_path',
    width: '15%'
  },
  {
    title: '方法',
    dataIndex: 'request_method',
    width: '5%'
  },
  {
    title: '规则',
    dataIndex: 'rule_type',
    width: '8%'
  },
  {
    title: '防护动作',
    dataIndex: 'action',
    width: '8%',
    scopedSlots: { customRender: 'action' }
  },
  {
    title: 'IP',
    dataIndex: 'real_client_ip',
    width: '10%'
  },
  {
    title: '请求ID',
    dataIndex: 'request_id'
  },
  {
    title: '时间',
    dataIndex: 'request_time',
    scopedSlots: { customRender: 'time' }
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
        start_time: '',
        end_time: '',
        site_id: 0,
        action: '',
        content: ''
      },
      // detailVisible: false
      time: ['', ''],
      showTime: {
        defaultValue: [moment('00:00:00', 'HH:mm:ss'), moment('23:59:59', 'HH:mm:ss')]
      },
      sites: []
    }
  },
  methods: {
    onClose() {
      this.detailVisible = false
    },

    // prettier-ignore
    onReset() {
      this.time = ['', ''],
      this.queryParams.page_num = 1,
      this.queryParams = {
        page_num: 1,
        page_size: 10,
        start_time: '',
        end_time: '',
        site_id: 0,
        action: '',
        content: ''
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

    onSearch() {
      this.queryParams.page_num = 1
      this.getList()
    },

    // toDetail(id) {
    //   this.detailVisible = true
    // },
    getAllSites() {
      ListSite().then(res => {
        this.sites = res.data.list
      })
    },

    getList() {
      ListAttackLog(this.queryParams)
        .then(res => {
          if (res.code == 0) {
            this.list = res.data.list || []
            this.meta = res.data.meta
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(() => {
          this.$message.error('网络异常，请稍候再试')
        })
    }
  },
  mounted() {
    this.getList()
    this.getAllSites()
  }
}
</script>
