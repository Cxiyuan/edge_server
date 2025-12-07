<template>
  <div class="online-users">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>在线用户</span>
          <el-button type="primary" @click="fetchOnlineUsers" :icon="Refresh">刷新</el-button>
        </div>
      </template>
      
      <el-table :data="onlineUsers" stripe style="width: 100%">
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="group_name" label="用户组" width="120" />
        <el-table-column prop="virtual_ip" label="虚拟IP" width="140" />
        <el-table-column prop="remote_ip" label="远端IP" width="140" />
        <el-table-column prop="mac" label="MAC地址" width="150" />
        <el-table-column prop="protocol" label="协议" width="100" />
        <el-table-column prop="virtual_dev" label="虚拟网卡" width="100" />
        <el-table-column prop="mtu" label="MTU" width="80" />
        <el-table-column label="上行速率" width="120">
          <template #default="scope">
            {{ formatBytes(scope.row.upload_speed) }}/s
          </template>
        </el-table-column>
        <el-table-column label="下行速率" width="120">
          <template #default="scope">
            {{ formatBytes(scope.row.download_speed) }}/s
          </template>
        </el-table-column>
        <el-table-column label="总上行" width="120">
          <template #default="scope">
            {{ formatBytes(scope.row.total_upload) }}
          </template>
        </el-table-column>
        <el-table-column label="总下行" width="120">
          <template #default="scope">
            {{ formatBytes(scope.row.total_download) }}
          </template>
        </el-table-column>
        <el-table-column label="连接时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.connected_at) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'

const onlineUsers = ref([])
let timer = null

const fetchOnlineUsers = async () => {
  try {
    const response = await axios.get('/api/online')
    onlineUsers.value = response.data.data || []
  } catch (error) {
    ElMessage.error('获取在线用户失败')
  }
}

const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

const formatTime = (timeStr) => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchOnlineUsers()
  timer = setInterval(fetchOnlineUsers, 3000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
