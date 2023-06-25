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

    <a-form-model
      ref="form"
      :model="form"
      :rules="rules"
      :label-col="labelCol"
      :wrapper-col="wrapperCol"
      label-align="left"
    >
      <a-form-model-item label="回源host" prop="origin_host_header">
        <a-input v-if="editable" style="width: 200px" placeholder="" v-model="form.origin_host_header" />
        <template v-else>{{ form.origin_host_header }}</template>
      </a-form-model-item>
      <a-form-model-item label="回源协议">
        <a-radio-group v-if="editable" v-model="form.origin_protocol">
          <a-radio value="http">HTTP</a-radio>
          <a-radio value="https">HTTPS</a-radio>
        </a-radio-group>
        <template v-else>
          {{ form.origin_protocol }}
        </template>
      </a-form-model-item>
      <a-form-model-item label="源站地址">
        <a-table :columns="columns" :data-source="form.origins" :pagination="false" :rowKey="record => record.id">
          <template v-for="col in ['addr', 'port', 'weight']" :slot="col" slot-scope="text, r, idx">
            <div :key="col" class="origin">
              <template v-if="editable">
                <template v-if="col == 'addr'">
                  <a-form-model-item :prop="col">
                    <a-input
                      style="margin: -5px 0"
                      :placeholder="originPlaceholder[col]"
                      v-model="form.origins[idx][col]"
                    />
                  </a-form-model-item>
                </template>
                <template v-else>
                  <a-form-model-item :prop="col">
                    <a-input-number
                      style="margin: -5px 0"
                      :placeholder="originPlaceholder[col]"
                      v-model="form.origins[idx][col]"
                    />
                  </a-form-model-item>
                </template>
              </template>
              <template v-else>
                {{ text }}
              </template>
            </div>
          </template>

          <!-- <template slot="addr" slot-scope="v, r, index">
            <a-input type="text" placeholder="请输入IP或域名" v-model="addForm.origins[index].host" />
          </template>
          <template slot="http_port" slot-scope="v, r, index">
            <a-input type="number" :placeholder="80" v-model="addForm.origins[index].http_port" />
          </template>
          <template slot="weight" slot-scope="v, r, index">
            <a-input type="number" v-model="addForm.origins[index].weight" />
          </template> -->

          <template slot="operation" slot-scope="text, record">
            <div class="origin">
              <a-form-model-item v-if="editable">
                <a :disabled="editingKey !== ''" @click="() => onDelItem(record.id)">删除</a>
              </a-form-model-item>
            </div>
          </template>
        </a-table>

        <!-- eslint-disable-next-line -->
        <a-button v-if="editable" type="dashed" @click="onAddOriginItem" style="width: 100%; margin: 20px 0"
          >+ 新增</a-button
        >
      </a-form-model-item>

      <template v-if="editable">
        <a-row>
          <a-col>
            <a-button style="margin-left: 10px;" type="primary" @click="onOK">
              保存
            </a-button>
            <a-button style="margin-left: 10px;" @click="onCancel">
              取消
            </a-button>
          </a-col>
        </a-row>
      </template>
    </a-form-model>
  </div>
</template>
<script>
import { GetOrigins, UpdateOriginCfg } from '@/api/site/origin'
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
      columns,
      editingKey: '',

      // add
      labelCol: { span: 2 },
      wrapperCol: { span: 20 },
      form: {
        origin_host_header: '',
        origin_protocol: '',
        origins: []
      },

      editable: false,
      originPlaceholder,
      rules: {
        // addr: [{ required: true, message: '请输入源站地址（IP/域名）', trigger: 'blur' }],
        // port: [{ required: true, message: '请输入源站端口', trigger: 'blur' }],
        // weight: [{ required: true, message: '请输入源站权重', trigger: 'blur' }],
        // origin_host_header: [{ required: true, message: '请输入源站权重', trigger: 'blur' }]
      }
      // rules: {}
    }
  },
  methods: {
    onEdit() {
      this.editable = true
    },

    onCancel() {
      this.doGetOrigins()
      this.editable = false
    },

    onOK() {
      // this.$refs.form.validate(valid => {
      //   if (valid) {
      //     console.log('ok')
      //   }
      //   console.log(valid)
      // })

      const siteID = this.$route.params.id
      this.doUpdateOrigin(siteID, this.form)
    },

    // handleChange(value, key, column) {
    //   const newData = [...this.data]
    //   const target = newData.find(item => key === item.id)
    //   if (target) {
    //     if (column == 'http_port' || column == 'weight') {
    //       value = Number(value)
    //     }
    //     target[column] = value
    //     this.data = newData
    //   }
    // },
    // edit(key) {
    //   const newData = [...this.data]
    //   const target = newData.find(item => key === item.id)
    //   this.editingKey = key
    //   if (target) {
    //     target.editable = true
    //     this.data = newData
    //   }
    // },
    onDelItem(key) {
      let newData = [...this.form.origins]
      const target = newData.find(item => key === item.id)
      if (target) {
        newData.map((value, idx) => {
          if (value.id == key) {
            newData.splice(idx, 1)
            this.form.origins = newData
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
      let max = 0
      this.form.origins.forEach(item => {
        if (item.id > max) {
          max = item.id
        }
      })

      const newData = {
        id: max + 1,
        addr: '',
        port: null,
        weight: null
      }
      this.form.origins = [...this.form.origins, newData]
      console.log(this.form.origins)
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
          this.form = res.data
          // this.cacheData = this.data.map(item => ({ ...item }))
        }
      })
    },

    // doUpdateOrigin(id, data) {
    //   return new Promise(resolve => {
    //     UpdateOrigin(id, data)
    //       .then(res => {
    //         if (res.code == 0) {
    //           this.$message.success('更新成功！')
    //           resolve(true)
    //         } else {
    //           this.$message.error(res.msg)
    //           resolve(false)
    //         }
    //       })
    //       .catch(err => {
    //         this.$message.error(err.msg)
    //         resolve(false)
    //       })
    //   })
    // },

    doUpdateOrigin(id, data) {
      UpdateOriginCfg(id, data).then(res => {
        if (res.code == 0) {
          this.$message.success('更新成功！')
          this.doGetOrigins()
          this.editable = false
        } else {
          this.$message.error(res.msg)
        }
      })
    }

    // doAddOrigins() {
    //   const id = this.$route.params.id
    //   const data = this.form.origins
    //   AddOrigins(id, data)
    //     .then(res => {
    //       if (res.code == 0) {
    //         this.$message.success('添加成功')
    //         this.visible = false
    //         this.doGetOrigins()
    //       } else {
    //         this.$message.error(res.msg)
    //       }
    //     })
    //     .catch(err => {
    //       this.$message.error(err.msg)
    //     })
    // }
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
.ant-form-item {
  margin-bottom: 5px;
}

.origin .ant-form-item {
  /* -webkit-box-sizing: border-box;
  box-sizing: border-box;
  margin: 0;
  padding: 0;
  color: rgba(0, 0, 0, 0.65);
  font-size: 14px;
  font-variant: tabular-nums;
  line-height: 1.5;
  list-style: none;
  -webkit-font-feature-settings: 'tnum';
  font-feature-settings: 'tnum'; */
  margin-bottom: 0;
  /* vertical-align: top; */
}
</style>
