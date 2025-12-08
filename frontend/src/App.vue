<template>
  <div class="layout-container">
    <router-view v-if="$route.path === '/login'" />
    <el-container v-else>
      <el-header class="header">
        <div class="header-left">
          <el-icon :size="28" color="#409EFF"><Connection /></el-icon>
          <span class="title">端点网络接入平台</span>
        </div>
        <div class="header-right">
          <el-icon :size="20"><User /></el-icon>
          <span class="username">{{ username }}</span>
          <el-button type="primary" size="small" @click="handleLogout" style="margin-left: 16px">
            退出登录
          </el-button>
        </div>
      </el-header>
      <el-container>
        <el-aside width="240px" class="sidebar">
          <el-menu
            :default-active="$route.path"
            router
            background-color="#001529"
            text-color="#fff"
            active-text-color="#409EFF"
          >
            <el-menu-item index="/">
              <el-icon><DataLine /></el-icon>
              <span>首页</span>
            </el-menu-item>
            <el-menu-item index="/groups">
              <el-icon><Grid /></el-icon>
              <span>用户组配置</span>
            </el-menu-item>
            <el-menu-item index="/users">
              <el-icon><User /></el-icon>
              <span>用户管理</span>
            </el-menu-item>
            <el-menu-item index="/online">
              <el-icon><Connection /></el-icon>
              <span>在线用户</span>
            </el-menu-item>
            <el-menu-item index="/logs">
              <el-icon><Document /></el-icon>
              <span>日志审计</span>
            </el-menu-item>
            <el-menu-item index="/settings">
              <el-icon><Setting /></el-icon>
              <span>系统设置</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Connection, User, DataLine, Grid, Document, Setting } from '@element-plus/icons-vue'

const router = useRouter()
const username = ref('管理员')

onMounted(() => {
  const storedUsername = localStorage.getItem('username')
  if (storedUsername) {
    username.value = storedUsername
  }
})

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗?', '提示', {
      type: 'warning'
    })
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    router.push('/login')
  } catch (error) {
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.el-container {
  height: 100%;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(90deg, #1e3c72 0%, #2a5298 100%);
  color: #fff;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title {
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 1px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.username {
  font-size: 14px;
}

.sidebar {
  background-color: #001529;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
}

.main-content {
  background-color: #f0f2f5;
  padding: 24px;
  overflow-y: auto;
}
</style>