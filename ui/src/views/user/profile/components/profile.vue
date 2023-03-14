<template>
  <a-card style="margin-top: 24px" :bordered="false" title="用户信息">
    <a-descriptions>
      <a-descriptions-item label="用户名">{{ data.username }}</a-descriptions-item>
      <a-descriptions-item></a-descriptions-item>
      <a-descriptions-item></a-descriptions-item>
      <a-descriptions-item label="用户姓名">{{ data.name }}</a-descriptions-item>
      <a-descriptions-item label="角色">{{ data.role || '无' }}</a-descriptions-item>
      <a-descriptions-item></a-descriptions-item>
      <a-descriptions-item label="邮箱">{{ data.email || '无' }}</a-descriptions-item>
      <a-descriptions-item label="联系方式">{{ data.phone }}</a-descriptions-item>
    </a-descriptions>
    <a-descriptions title="登录信息">
      <a-descriptions-item label="上次登录IP">{{ data.last_sign_in_ip || '无' }}</a-descriptions-item>
      <a-descriptions-item label="上次登录时间">{{ data.last_sign_in_at || '无' }}</a-descriptions-item>
      <a-descriptions-item></a-descriptions-item>
    </a-descriptions>
  </a-card>
</template>

<script>
import { GetUser } from '@/api/user'
export default {
  data() {
    return {
      data: {}
      // END
    }
  },

  methods: {
    getProfile() {
      GetUser(this.userID).then(res => {
        if (res.code === 0) {
          this.data = res.data
        }
      })
    }
  },

  computed: {
    userID() {
      return this.$route.params.id
    }
  },

  mounted() {
    this.getProfile()
  }
}
</script>
