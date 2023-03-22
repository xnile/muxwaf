<template>
  <div>
    <a-row>
      <div class="item">
        <a-col :span="22"></a-col>
        <a-col :span="2">
          <!-- <a @click="onEdit">修改</a> -->
        </a-col>
      </div>
    </a-row>

    <!-- 添加源站 -->
    <a-modal
      title="新增源站"
      :visible="visible"
      :confirm-loading="confirmLoading"
      @ok="handleOk"
      @cancel="handleCancel"
      :width="800"
    >
      <a-form-model :model="addForm" :label-col="labelCol" :wrapper-col="wrapperCol">
        <!-- <a-form-model-item label="回源协议">
          <a-radio-group v-model="form.origin_protocol" @change="onChange">
            <a-radio :value="0">HTTP</a-radio>
            <a-radio :value="1">HTTPS</a-radio>
            <a-radio :value="2">跟随</a-radio>
          </a-radio-group>
        </a-form-model-item> -->
        <a-form-model-item label="" prop="origins">
          <a-table
            :columns="originColumns"
            :dataSource="addForm.origins"
            :rowKey="record => record.key"
            :pagination="false"
          >
            <!-- <template slot="index" slot-scope="v, r, index">{{ index + 1 }}</template> -->
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
          </a-table>
          <a-button style="width: 100%; margin: 20px 0" type="dashed" @click="onAddOriginItem">+ 新增</a-button>
        </a-form-model-item>
      </a-form-model>
    </a-modal>
    <!-- 添加源站 END -->

    <a-table :columns="columns" :data-source="data" :pagination="false" :rowKey="record => record.id">
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
            <!-- <a @click="() => save(record.key)">保存</a> -->
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
    <!-- <a-button style="width: 100%; margin: 20px 0" type="dashed" @click="addSrc">+ 新增</a-button> -->
    <a-button type="primary" @click="onAddOrigin">新增源站</a-button>
  </div>
</template>
<script>
import { GetOrigins, UpdateOrigin, AddOrigins, DelOrigin } from '@/api/site/origin'
const columns = [
  {
    title: '源站',
    dataIndex: 'host',
    width: '20%',
    scopedSlots: { customRender: 'host' }
  },
  {
    title: '端口',
    dataIndex: 'http_port',
    width: '15%',
    scopedSlots: { customRender: 'http_port' }
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
    scopedSlots: { customRender: 'operation' }
  }
]
const originColumns = [
  {
    title: '源站',
    dataIndex: 'host',
    width: 300,
    scopedSlots: { customRender: 'host' }
  },
  {
    title: '端口',
    dataIndex: 'http_port',
    width: 100,
    scopedSlots: { customRender: 'http_port' }
  },
  {
    title: '权重',
    dataIndex: 'weight',
    width: 100,
    scopedSlots: { customRender: 'weight' }
  },
  {
    title: '操作',
    dataIndex: 'action',
    scopedSlots: { customRender: 'operation' }
  }
]

// const data = []
// for (let i = 0; i < 2; i++) {
//   data.push({
//     key: i.toString(),
//     host: `Edrward ${i}`,
//     port: 80,
//     weight: 32
//   })
// }
export default {
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
      originColumns,
      visible: false,
      confirmLoading: false,
      addForm: {
        origins: [
          {
            key: 0,
            host: '',
            http_port: 80,
            weight: 100
          }
        ]
      }
    }
  },
  methods: {
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
    edit(key) {
      const newData = [...this.data]
      const target = newData.find(item => key === item.id)
      this.editingKey = key
      if (target) {
        target.editable = true
        this.data = newData
      }
    },
    del(key) {
      let newData = [...this.data]
      const target = newData.find(item => key === item.id)
      if (target) {
        newData.map((value, idx) => {
          if (value.id == key) {
            DelOrigin(key).then(res => {
              if (res.code == 0) {
                newData.splice(idx, 1)
                this.data = newData
                this.$message.success('删除成功！')
              } else {
                this.$message.error(res.msg)
              }
            })
          }
        })
      }
    },
    save(key) {
      console.log(key)
      const newData = [...this.data]
      const newCacheData = [...this.cacheData]
      const target = newData.find(item => key === item.id)
      const targetCache = newCacheData.find(item => key === item.id)
      if (target && targetCache) {
        this.doUpdateOrigin(key, target).then(res => {
          if (res) {
            delete target.editable
            this.data = newData
            Object.assign(targetCache, target)
            this.cacheData = newCacheData
          } else {
          }
        })
      }
      this.editingKey = ''
    },
    cancel(key) {
      console.log(this.cacheData)
      const newData = [...this.data]
      const target = newData.find(item => key === item.id)
      this.editingKey = ''
      if (target) {
        Object.assign(
          target,
          this.cacheData.find(item => key === item.id)
        )
        delete target.editable
        this.data = newData
      }
    },

    // add
    onAddOrigin() {
      // this.$parent.$emit('changeOperation', '编辑')
      this.visible = true
    },
    onAddOriginItem() {
      const newData = {
        key: this.count,
        host: '',
        http_port: 80,
        weight: 100
      }
      this.addForm.origins = [...this.addForm.origins, newData]
      // this.addForm.origins.push({ host: '', port: 80, weight: 100 })
    },
    onDelOrigin(record, index) {
      this.addForm.origins.splice(index, 1)
    },
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
