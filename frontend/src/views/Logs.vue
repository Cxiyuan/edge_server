<template>
  <div class="logs">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="认证日志" name="auth">
        <el-table :data="authLogs" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="remote_ip" label="远端IP" width="150" />
          <el-table-column prop="action" label="操作" width="120" />
          <el-table-column label="结果" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.success ? 'success' : 'danger'">
                {{ scope.row.success ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="消息" />
          <el-table-column label="时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.created_at) }}
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-model:current-page="authPage"
          v-model:page-size="authPageSize"
          :total="authTotal"
          @current-change="fetchAuthLogs"
          layout="total, prev, pager, next"
          style="margin-top: 20px; justify-content: center"
        />
      </el-tab-pane>

      <el-tab-pane label="访问日志" name="access">
        <el-table :data="accessLogs" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" width="120" />
          <el-table-column prop="src_ip" label="源IP" width="140" />
          <el-table-column prop="dst_ip" label="目标IP" width="140" />
          <el-table-column prop="dst_port" label="目标端口" width="100" />
          <el-table-column prop="protocol" label="协议" width="100" />
          <el-table-column prop="action" label="动作" width="100" />
          <el-table-column label="发送" width="120">
            <template #default="scope">
              {{ formatBytes(scope.row.bytes_sent) }}
            </template>
          </el-table-column>
          <el-table-column label="接收" width="120">
            <template #default="scope">
              {{ formatBytes(scope.row.bytes_recv) }}
            </template>
          </el-table-column>
          <el-table-column label="时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.created_at) }}
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-model:current-page="accessPage"
          v-model:page-size="accessPageSize"
          :total="accessTotal"
          @current-change="fetchAccessLogs"
          layout="total, prev, pager, next"
          style="margin-top: 20px; justify-content: center"
        />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'

const activeTab = ref('auth')
const authLogs = ref([])
const accessLogs = ref([])
const authPage = ref(1)
const authPageSize = ref(20)
const authTotal = ref(0)
const accessPage = ref(1)
const accessPageSize = ref(20)
const accessTotal = ref(0)

const fetchAuthLogs = async () => {
  try {
    const response = await axios.get('/api/logs/auth', {
      params: { page: authPage.value, pageSize: authPageSize.value }
    })
    authLogs.value = response.data.data || []
    authTotal.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取认证日志失败')
  }
}

const fetchAccessLogs = async () => {
  try {
    const response = await axios.get('/api/logs/access', {
      params: { page: accessPage.value, pageSize: accessPageSize.value }
    })
    accessLogs.value = response.data.data || []
    accessTotal.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取访问日志失败')
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
  fetchAuthLogs()
  fetchAccessLogs()
})
</script>

<style scoped>
.logs {
  background-color: #fff;
  padding: 20px;
  border-radius: 4px;
}
</style>
