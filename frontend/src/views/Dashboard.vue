<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#409EFF"><Monitor /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.cpuUsage.toFixed(1) }}%</div>
              <div class="stat-label">CPU使用率</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#67C23A"><Cpu /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.memoryUsage.toFixed(1) }}%</div>
              <div class="stat-label">内存使用率</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#E6A23C"><Connection /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.onlineUsers }}</div>
              <div class="stat-label">在线用户</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#F56C6C"><Clock /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ formatUptime(stats.uptime) }}</div>
              <div class="stat-label">运行时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统资源监控</span>
            </div>
          </template>
          <div ref="cpuChart" style="height: 300px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>网络连接统计</span>
            </div>
          </template>
          <div ref="networkChart" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Monitor, Cpu, Connection, Clock } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import axios from 'axios'

const stats = ref({
  cpuUsage: 0,
  memoryUsage: 0,
  diskUsage: 0,
  networkConnections: 0,
  onlineUsers: 0,
  uptime: 0
})

const cpuChart = ref(null)
const networkChart = ref(null)
let cpuChartInstance = null
let networkChartInstance = null
let timer = null

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/stats')
    stats.value = response.data.data
    updateCharts()
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const formatUptime = (seconds) => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${days}天${hours}时${minutes}分`
}

const initCharts = () => {
  cpuChartInstance = echarts.init(cpuChart.value)
  cpuChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['CPU', '内存', '磁盘'] },
    xAxis: { type: 'category', data: [] },
    yAxis: { type: 'value', max: 100 },
    series: [
      { name: 'CPU', type: 'line', data: [], smooth: true, itemStyle: { color: '#409EFF' } },
      { name: '内存', type: 'line', data: [], smooth: true, itemStyle: { color: '#67C23A' } },
      { name: '磁盘', type: 'line', data: [], smooth: true, itemStyle: { color: '#E6A23C' } }
    ]
  })

  networkChartInstance = echarts.init(networkChart.value)
  networkChartInstance.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie',
      radius: '60%',
      data: [
        { value: stats.value.onlineUsers, name: '在线用户' },
        { value: stats.value.networkConnections - stats.value.onlineUsers, name: '其他连接' }
      ],
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }]
  })
}

const updateCharts = () => {
  if (cpuChartInstance) {
    const option = cpuChartInstance.getOption()
    const now = new Date().toLocaleTimeString()
    option.xAxis[0].data.push(now)
    option.series[0].data.push(stats.value.cpuUsage)
    option.series[1].data.push(stats.value.memoryUsage)
    option.series[2].data.push(stats.value.diskUsage)
    
    if (option.xAxis[0].data.length > 20) {
      option.xAxis[0].data.shift()
      option.series[0].data.shift()
      option.series[1].data.shift()
      option.series[2].data.shift()
    }
    
    cpuChartInstance.setOption(option)
  }

  if (networkChartInstance) {
    networkChartInstance.setOption({
      series: [{
        data: [
          { value: stats.value.onlineUsers, name: '在线用户' },
          { value: Math.max(0, stats.value.networkConnections - stats.value.onlineUsers), name: '其他连接' }
        ]
      }]
    })
  }
}

onMounted(() => {
  fetchStats()
  initCharts()
  timer = setInterval(fetchStats, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (cpuChartInstance) cpuChartInstance.dispose()
  if (networkChartInstance) networkChartInstance.dispose()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stat-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-card :deep(.el-card__body) {
  padding: 20px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #fff;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

.card-header {
  font-weight: 600;
  font-size: 16px;
}
</style>
