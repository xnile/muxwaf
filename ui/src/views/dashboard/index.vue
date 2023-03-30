<template>
  <div>
    <a-row :gutter="24">
      <a-col :sm="24" :md="24" :xl="24" style="margin-bottom: 24px;">
        <a-card>
          <a-space>
            <a-select placeholder="请选择" style="width: 200px;">
              <a-select-option value="">全部</a-select-option>
              <a-select-option :value="0">未启用</a-select-option>
              <a-select-option :value="1">已启用</a-select-option>
            </a-select>
            <span></span>
          </a-space>
          <a-radio-group>
            <a-radio-button value="今天">
              今天
            </a-radio-button>
            <a-radio-button value="7天">
              7天
            </a-radio-button>
            <a-radio-button value="30天">
              30天
            </a-radio-button>
          </a-radio-group>
          <a-range-picker />
        </a-card>
      </a-col>
    </a-row>
    <a-row :gutter="24">
      <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '4px' }">
        <chart-card :loading="loading" title="总拦截数" total="126,560"> </chart-card>
      </a-col>
      <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '24px' }">
        <chart-card :loading="loading" title="CC防护" :total="8846 | NumberFormat"> </chart-card>
      </a-col>
      <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '24px' }">
        <chart-card :loading="loading" title="IP黑名单" :total="6560 | NumberFormat"> </chart-card>
      </a-col>
      <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '24px' }">
        <chart-card :loading="loading" title="地域级IP黑名单" :total="6560 | NumberFormat"> </chart-card>
      </a-col>
    </a-row>

    <a-row>
      <a-col :span="24">
        <div class="gutter-box">
          <a-card :loading="loading" :bordered="false" :body-style="{ padding: '10px' }">
            <div>
              <a-tabs
                default-active-key="1"
                size="large"
                :tab-bar-style="{ marginBottom: '24px', paddingLeft: '16px' }"
              >
                <a-tab-pane loading="true" tab="QPS" key="1">
                  <!-- eslint-disable-next-line -->
                  <!-- <a-card :bordered="false" :body-style="{ padding: '0 15px 0 15px' }"> -->
                  <a-card :bordered="false" :body-style="{ padding: '0 15px 0 15px' }">
                    <div v-if="showChart">
                      <v-chart :forceFit="true" height="421" :data="data" :scale="scale" :padding="padding">
                        <v-tooltip :crosshairs="false"></v-tooltip>
                        <v-legend position="top-center"></v-legend>
                        <v-line position="Time*qps" color="#1890ff"></v-line>
                        <v-axis dataKey="Time" :label="labelFormater"></v-axis>
                        <v-axis dataKey="qps" :label="qpsLabelFormater"></v-axis>
                      </v-chart>
                    </div>
                  </a-card>
                </a-tab-pane>
                <!-- <a-tab-pane tab="访问量" key="2">
                  <a-card :bordered="false" :body-style="{ padding: '0' }">
                    <div v-if="showChart">
                      <v-chart :forceFit="true" height="400" :data="data" :scale="scale" :padding="padding">
                        <v-tooltip :crosshairs="false"></v-tooltip>
                        <v-legend position="top-center"></v-legend>
                        <v-line position="Time*qps" color="#2fc25b"></v-line>
                        <v-axis dataKey="Time" :label="label"></v-axis>
                        <v-axis dataKey="qps" :label="label"></v-axis>
                      </v-chart>
                    </div>
                  </a-card>
                </a-tab-pane> -->
              </a-tabs>
            </div>
          </a-card>
        </div>
      </a-col>
      <!-- <a-col :span="8">
        <div class="gutter-box">
          <a-card :loading="loading" :bordered="false" :body-style="{ padding: '10px' }">
            <div class="salesCard">
              <a-tabs
                default-active-key="1"
                size="large"
                :tab-bar-style="{ marginBottom: '24px', paddingLeft: '16px' }"
              >
                <a-tab-pane loading="true" tab="拦截TOP IP" key="1">
                  <rank-list title="" :list="rankList" />
                </a-tab-pane>
              </a-tabs>
            </div>
          </a-card>
        </div>
      </a-col> -->
    </a-row>

    <a-row :gutter="24" type="flex" :style="{ marginTop: '24px' }">
      <a-col :xl="12" :lg="24" :md="24" :sm="24" :xs="24">
        <a-card :loading="loading" :bordered="false" :style="{ minHeight: '520px' }" :body-style="{ padding: '10px' }">
          <a-tabs default-active-key="1" size="large" :tab-bar-style="{ marginBottom: '24px', paddingLeft: '16px' }">
            <a-tab-pane loading="true" tab="响应码占比" key="1">
              <!-- style="width: calc(100% - 240px);" -->
              <div>
                <v-chart :force-fit="true" :height="405" :data="pieData" :scale="pieScale">
                  <v-tooltip :showTitle="false" dataKey="item*percent" />
                  <v-axis />
                  <!-- position="right" :offsetX="-140" -->
                  <v-legend dataKey="item" />
                  <v-pie position="percent" color="item" :vStyle="pieStyle" />
                  <v-coord type="theta" :radius="0.85" :innerRadius="0.6" />
                </v-chart>
              </div>
            </a-tab-pane>
            <a-tab-pane loading="true" tab="响应码" key="2">
              <!-- <div class="rank">
                <ul class="list">
                  <li :key="index" v-for="(item, index) in responseCode">
                    <span :class="index < 3 ? 'active' : null"></span>
                    <span>{{ item.name }}</span>
                    <span>{{ item.total }}</span>
                  </li>
                </ul>
              </div> -->
              <div>
                <v-chart :forceFit="true" :height="405" :data="responseCode">
                  <v-coord type="rect" direction="LB" />
                  <v-tooltip />
                  <v-axis dataKey="code" :label="label" />
                  <v-bar position="code*total" />
                </v-chart>
              </div>
            </a-tab-pane>
          </a-tabs>
        </a-card>
      </a-col>
      <a-col :xl="12" :lg="24" :md="24" :sm="24" :xs="24">
        <a-card :loading="loading" :bordered="false" :body-style="{ padding: '10px' }">
          <a-tabs default-active-key="1" size="large" :tab-bar-style="{ marginBottom: '24px', paddingLeft: '16px' }">
            <a-tab-pane loading="true" tab="拦截TOP IP" key="1">
              <rank-list title="" :list="rankList" />
            </a-tab-pane>
          </a-tabs>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script>
import * as echarts from 'echarts'
import moment from 'moment'
import { ChartCard, MiniProgress, RankList, Bar, Trend, NumberInfo, MiniSmoothArea } from '@/components'
import { baseMixin } from '@/store/app-mixin'

const DataSet = require('@antv/data-set')

const rankList = [
  { name: '114.11.7.8', total: 5600 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 },
  { name: '114.11.7.8', total: 1000 }
]

const sourceResponseCode = [
  { code: '200', total: 45000 },
  { code: '400', total: 1000 },
  { code: '403', total: 1000 },
  { code: '404', total: 1000 },
  { code: '413', total: 1000 },
  { code: '499', total: 1000 },
  { code: '500', total: 1000 },
  { code: '502', total: 1000 },
  { code: '其它', total: 1000 }
]

const sourceData = [
  { item: '200', count: 32.2 },
  { item: '404', count: 21 },
  { item: '413', count: 17 },
  { item: '502', count: 13 },
  { item: '499', count: 9 },
  { item: '其他', count: 7.8 }
]

const pieScale = [
  {
    dataKey: 'percent',
    min: 0,
    formatter: '.0%'
  }
]

const dv = new DataSet.View().source(sourceData)
dv.transform({
  type: 'percent',
  field: 'count',
  dimension: 'item',
  as: 'percent'
})
const pieData = dv.rows

export default {
  name: 'Analysis',
  mixins: [baseMixin],
  components: {
    ChartCard,
    // MiniArea,
    // MiniBar,
    // MiniProgress,
    RankList
    // Bar,
    // Trend
    // NumberInfo,
    // MiniSmoothArea
  },
  data() {
    return {
      loading: true,
      rankList,

      // 搜索用户数
      // searchUserData,
      // searchUserScale,
      // searchData,

      // barData,
      // barData2,

      //
      pieScale,
      pieData,
      sourceData,
      pieStyle: {
        stroke: '#fff',
        lineWidth: 1
      },
      // xnile
      responseCode: [],
      showChart: false,
      data: [],
      label: {
        textStyle: {
          fill: '#aaaaaa'
        }
      },
      labelFormater: {
        textStyle: {
          fill: '#aaaaaa'
        },
        density: 0.5,
        formatter: function formatter(text) {
          var dataStrings = text.split(' ')
          return dataStrings[0]
        }
      },
      qpsLabelFormater: {
        textStyle: {
          fill: '#aaaaaa'
        },
        // density: 0.2,
        formatter: function formatter(text) {
          return text.replace(/(\d)(?=(?:\d{3})+$)/g, '$1,')
        }
      },

      scale: [
        {
          dataKey: 'Time',
          alias: '日期'
          // ticks: ['2018-08-09 20:00']
        }
      ],
      // scale: [
      //   // {
      //   //   dataKey: 'time'
      //   //   // tickCount: 12
      //   //   // type: 'time'
      //   // },
      //   {
      //     dataKey: 'Time',
      //     type: 'time',
      //     // tickInterval: 2,
      //     tickCount: 12
      //   }
      //   // {
      //   //   dataKey: 'Time',
      //   //   tickCount: 12,
      //   //   // tickInterval: 3600,
      //   //   type: 'time'
      //   // }
      // ],
      padding: [30, 20, 70, 70]
    }
  },
  methods: {},

  computed: {},
  created() {
    setTimeout(() => {
      this.loading = !this.loading
    }, 1000)
  },
  mounted() {
    let data = [
      {
        Time: '2018-08-09 20:00',
        qps: 1476
      },
      {
        Time: '2018-08-09 21:00',
        qps: 200
      },
      {
        Time: '2019-08-09 22:00',
        qps: 200
      },
      {
        Time: '2019-08-09 22:00',
        qps: 200
      }
    ]
    var ds = new DataSet()
    var dv = ds.createView().source(data)
    dv.transform({
      type: 'map',
      callback: function callback(row) {
        var times = row.Time.split(' ')
        row.date = times[0]
        row.time = times[1]
        return row
      }
    })
    this.$data.data = dv.rows
    this.$nextTick(() => {
      this.showChart = true
    })

    let dv2 = new DataSet.View().source(sourceResponseCode)
    dv2.transform({
      type: 'sort',
      callback(a, b) {
        return a.total - b.total > 0
      }
    })
    this.responseCode = dv2.rows
  }
}
</script>

<style lang="less" scoped>
/deep/ .chart-card-footer {
  border-top: initial;
  padding-top: 9px;
  margin-top: 8px;
}

// .gutter-box {
//   background: #f0f2f5;
//   padding: 0 10px 0 0;
// }

// .rank {
//   padding: 0 32px 32px 72px;
//   min-height: 380px;

//   .list {
//     margin: 25px 0 0;
//     padding: 0;
//     list-style: none;

//     li {
//       margin-top: 16px;

//       span {
//         color: rgba(0, 0, 0, 0.65);
//         font-size: 14px;
//         line-height: 22px;

//         // &:first-child {
//         //   background-color: #f5f5f5;
//         //   border-radius: 20px;
//         //   display: inline-block;
//         //   font-size: 12px;
//         //   font-weight: 600;
//         //   margin-right: 24px;
//         //   height: 20px;
//         //   line-height: 20px;
//         //   width: 20px;
//         //   text-align: center;
//         // }
//         &.active {
//           background-color: #314659;
//           color: #fff;
//         }
//         &:last-child {
//           float: right;
//           padding-right: 30px;
//         }
//       }
//     }
//   }
// }
</style>
