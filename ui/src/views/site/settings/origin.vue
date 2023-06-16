<template>
  <div>
    <a-row>
      <div class="item">
        <a-col :span="22"></a-col>
        <a-col :span="2">
          <a @click="onEdit">修改</a>
        </a-col>
      </div>
    </a-row>

    <!-- 添加源站 -->
    <!-- <a-modal
      title="新增源站"
      :visible="visible"
      :confirm-loading="confirmLoading"
      @ok="handleOk"
      @cancel="handleCancel"
      :width="800"
    > -->
    <a-form-model :model="addForm" :label-col="labelCol" :wrapper-col="wrapperCol">
      <a-form-model-item label="回源host">
        <a-input v-if="editable" style="width: 200px" placeholder="" />
        <template v-else>www.xnile.cn</template>
      </a-form-model-item>
      <a-form-model-item label="回源协议">
        <a-radio-group v-if="editable">
          <a-radio value="http">HTTP</a-radio>
          <a-radio value="https">HTTPS</a-radio>
        </a-radio-group>
        <template v-else>
          http
        </template>
      </a-form-model-item>
      <a-form-model-item label="源站地址" prop="origins">
        <!-- <a-table
          :columns="originColumns"
          :dataSource="addForm.origins"
          :rowKey="record => record.key"
          :pagination="false"
        >
          <template slot="host" slot-scope="v, r, index">
            <a-input type="text" placeholder="请输入IP或域名" v-model="addForm.origins[index].host" />
          </template>
          <template slot="http_port" slot-scope="v, r, index">
            <a-input type="number" :placeholder="80" v-model="addForm.origins[index].http_port" />
          </template>
          <template slot="weight" slot-scope="v, r, index">
            <a-input type="number" v-model="addForm.origins[index].weight" />
          </template>
          <template slot="operation" slot-scope="v, record, index">
            <a-space>
              <a-button type="link" @click="onDelOrigin(record, index)">删除</a-button>
            </a-space>
          </template>
        </a-table> -->

        <a-table :columns="columns" :data-source="data" :pagination="false" :rowKey="record => record.id">
          <template v-for="col in ['addr', 'port', 'weight']" :slot="col" slot-scope="text, record">
            <div :key="col">
              <a-input
                v-if="editable"
                style="margin: -5px 0"
                :placeholder="originPlaceholder[col]"
                :value="text"
                @change="e => handleChange(e.target.value, record.id, col)"
              />
              <template v-else>
                {{ text }}
              </template>
            </div>
          </template>
          <template slot="operation" slot-scope="text, record">
            <div class="editable-row-operations">
              <!-- <span v-if="record.editable">
                <a-popconfirm title="确认更改?" @confirm="() => save(record.id)">
                  <a>保存</a>
                </a-popconfirm>
                <a-popconfirm title="确认取消?" @confirm="() => cancel(record.id)">
                  <a>取消</a>
                </a-popconfirm>
              </span> -->
              <span v-if="editable">
                <a :disabled="editingKey !== ''" @click="() => del(record.id)">删除</a>
              </span>
            </div>
          </template>
        </a-table>

        <a-button style="width: 100%; margin: 20px 0" type="dashed" @click="onAddOriginItem">+ 新增</a-button>
      </a-form-model-item>

      <a-form-model-item :wrapper-col="{ span: 14, offset: 1 }">
        <a-button type="primary">
          保存
        </a-button>
        <a-button style="margin-left: 10px;" @click="onCancel">
          取消
        </a-button>
      </a-form-model-item>
    </a-form-model>
    <!-- </a-modal> -->
    <!-- 添加源站 END -->

    <!-- <a-table :columns="columns" :data-source="data" :pagination="false" :rowKey="record => record.id">
      <template v-for="col in ['host', 'http_port', 'weight']" :slot="col" slot-scope="text, record">
        <div :key="col">
          <a-input
            v-if="record.editable"
            style="margin: -5px 0"
            :value="text"
            @change="e => handleChange(e.target.value, record.id, col)"
          />
          <template v-else>
            {{ text }}
          </template>
        </div>
      </template>
      <template slot="operation" slot-scope="text, record">
        <div class="editable-row-operations">
          <span v-if="record.editable">
            <a-popconfirm title="确认更改?" @confirm="() => save(record.id)">
              <a>保存</a>
            </a-popconfirm>
            <a-popconfirm title="确认取消?" @confirm="() => cancel(record.id)">
              <a>取消</a>
            </a-popconfirm>
          </span>
          <span v-else>
            <a :disabled="editingKey !== ''" @click="() => edit(record.id)">编辑</a>
            <a :disabled="editingKey !== ''" @click="() => del(record.id)">删除</a>
          </span>
        </div>
      </template>
    </a-table>
    <a-button type="primary" @click="onAddOrigin">新增源站</a-button> -->
  </div>
</template>
<script>
import { GetOrigins, UpdateOrigin, AddOrigins, DelOrigin } from '@/api/site/origin'
import templates from '@/views/templates.vue'
const columns = [
  {
    title: '回源地址',
    dataIndex: 'addr',
    width: '60%',
    scopedSlots: { customRender: 'addr' }
  },
  {
    title: '端口',
    dataIndex: 'port',
    width: '15%',
    scopedSlots: { customRender: 'port' }
  },
  {
    title: '权重',
    dataIndex: 'weight',
    width: '15%',
    scopedSlots: { customRender: 'weight' }
  },
  {
    title: '操作',
    dataIndex: 'operation',
    scopedSlots: { customRender: 'operation' },
    width: '20%'
  }
]
// const originColumns = [
//   {
//     title: '回源地址',
//     dataIndex: 'host',
//     width: 300,
//     scopedSlots: { customRender: 'host' }
//   },
//   {
//     title: '端口',
//     dataIndex: 'http_port',
//     width: 100,
//     scopedSlots: { customRender: 'http_port' }
//   },
//   {
//     title: '权重',
//     dataIndex: 'weight',
//     width: 100,
//     scopedSlots: { customRender: 'weight' }
//   },
//   {
//     title: '操作',
//     dataIndex: 'action',
//     scopedSlots: { customRender: 'operation' }
//   }
// ]

// const data = []
// for (let i = 0; i < 2; i++) {
//   data.push({
//     key: i.toString(),
//     host: `Edrward ${i}`,
//     port: 80,
//     weight: 32
//   })
// }

const originPlaceholder = {
  addr: '请输入源站地址（IP/域名）',
  port: '1-65535',
  weight: '0-100'
}
export default {
  components: { templates },
  data() {
    // this.cacheData = data.map(item => ({ ...item }))
    return {
      data: [],
      cacheData: [],
      columns,
      editingKey: '',
      count: 1,

      // add
      labelCol: { span: 2 },
      wrapperCol: { span: 20 },
      // originColumns,
      visible: false,
      confirmLoading: false,
      form: {
        origin_host_header: '',
        origin_protocol: 'http',
        origins: [
          {
            key: 0,
            host: '',
            http_port: 80,
            weight: 100
          }
        ]
      },

      editable: false,
      originPlaceholder
    }
  },
  methods: {
    onEdit() {
      this.editable = true
    },

    onCancel() {
      this.editable = false
    },

    onChange() {},

    handleChange(value, key, column) {
      const newData = [...this.data]
      const target = newData.find(item => key === item.id)
      if (target) {
        if (column == 'http_port' || column == 'weight') {
          value = Number(value)
        }
        target[column] = value
        this.data = newData
      }
    },
    // edit(key) {
    //   const newData = [...this.data]
    //   const target = newData.find(item => key === item.id)
    //   this.editingKey = key
    //   if (target) {
    //     target.editable = true
    //     this.data = newData
    //   }
    // },
    del(key) {
      let newData = [...this.data]
      console.log(newData)
      const target = newData.find(item => key === item.id)
      if (target) {
        newData.map((value, idx) => {
          if (value.id == key) {
            console.log('key: ', key)
            console.log('idx: ', idx)
            newData.splice(idx, 1)
            this.data = newData
            // DelOrigin(key).then(res => {
            //   if (res.code == 0) {
            //     newData.splice(idx, 1)
            //     this.data = newData
            //     this.$message.success('删除成功！')
            //   } else {
            //     this.$message.error(res.msg)
            //   }
            // })
          }
        })
      }
    },
    // save(key) {
    //   console.log(key)
    //   const newData = [...this.data]
    //   const newCacheData = [...this.cacheData]
    //   const target = newData.find(item => key === item.id)
    //   const targetCache = newCacheData.find(item => key === item.id)
    //   if (target && targetCache) {
    //     this.doUpdateOrigin(key, target).then(res => {
    //       if (res) {
    //         delete target.editable
    //         this.data = newData
    //         Object.assign(targetCache, target)
    //         this.cacheData = newCacheData
    //       } else {
    //       }
    //     })
    //   }
    //   this.editingKey = ''
    // },
    // cancel(key) {
    //   console.log(this.cacheData)
    //   const newData = [...this.data]
    //   const target = newData.find(item => key === item.id)
    //   this.editingKey = ''
    //   if (target) {
    //     Object.assign(
    //       target,
    //       this.cacheData.find(item => key === item.id)
    //     )
    //     delete target.editable
    //     this.data = newData
    //   }
    // },

    // // add
    // onAddOrigin() {
    //   // this.$parent.$emit('changeOperation', '编辑')
    //   this.visible = true
    // },
    onAddOriginItem() {
      const newData = {
        key: this.count,
        id: 13,
        addr: '',
        port: null,
        weight: null
      }
      this.data = [...this.data, newData]
      console.log(this.data)
      // this.addForm.origins.push({ host: '', port: 80, weight: 100 })
    },
    // onDelOrigin(record, index) {
    //   this.addForm.origins.splice(index, 1)
    // },
    handleOk() {
      this.doAddOrigins()
    },
    handleCancel() {
      this.visible = false
    },

    doGetOrigins() {
      let id = this.$route.params.id
      GetOrigins(id).then(res => {
        if (res.code == 0) {
          this.data = res.data
          this.cacheData = this.data.map(item => ({ ...item }))
        }
      })
    },

    doUpdateOrigin(id, data) {
      return new Promise(resolve => {
        UpdateOrigin(id, data)
          .then(res => {
            if (res.code == 0) {
              this.$message.success('更新成功！')
              resolve(true)
            } else {
              this.$message.error(res.msg)
              resolve(false)
            }
          })
          .catch(err => {
            console.log(err)
            this.$message.error(err.msg)
            resolve(false)
          })
      })
    },

    doAddOrigins() {
      const id = this.$route.params.id
      const data = this.addForm.origins
      AddOrigins(id, data)
        .then(res => {
          if (res.code == 0) {
            this.$message.success('添加成功')
            this.visible = false
            this.doGetOrigins()
          } else {
            this.$message.error(res.msg)
          }
        })
        .catch(err => {
          this.$message.error(err.msg)
        })
    }
  },
  mounted() {
    // this.doGetOrigins()
  },
  activated() {
    // 在首次挂载、
    // 以及每次从缓存中被重新插入的时候调用
    this.doGetOrigins()
  }
}
</script>
<style scoped>
.editable-row-operations a {
  margin-right: 8px;
}
</style>
