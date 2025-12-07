<template>
  <div class="users">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="showDialog()">新增用户</el-button>
        </div>
      </template>
      
      <el-table :data="users" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="full_name" label="姓名" width="120" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="group_name" label="用户组" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'danger'">
              {{ scope.row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="showDialog(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteUser(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="currentUser.id ? '编辑用户' : '新增用户'"
      width="600px"
    >
      <el-form :model="currentUser" label-width="120px">
        <el-form-item label="用户名">
          <el-input v-model="currentUser.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="currentUser.password" type="password" placeholder="留空则不修改" />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="currentUser.full_name" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="currentUser.email" />
        </el-form-item>
        <el-form-item label="用户组">
          <el-select v-model="currentUser.group_id" placeholder="请选择用户组">
            <el-option
              v-for="group in groups"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="自定义路由">
          <el-input v-model="currentUser.custom_routes" placeholder="留空则继承用户组" />
        </el-form-item>
        <el-form-item label="自定义策略">
          <el-input v-model="currentUser.custom_policies" placeholder="留空则继承用户组" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="currentUser.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveUser">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

const users = ref([])
const groups = ref([])
const dialogVisible = ref(false)
const currentUser = ref({
  username: '',
  password: '',
  full_name: '',
  email: '',
  group_id: null,
  custom_routes: '',
  custom_policies: '',
  enabled: true
})

const fetchUsers = async () => {
  try {
    const response = await axios.get('/api/users')
    users.value = response.data.data || []
  } catch (error) {
    ElMessage.error('获取用户失败')
  }
}

const fetchGroups = async () => {
  try {
    const response = await axios.get('/api/groups')
    groups.value = response.data.data || []
  } catch (error) {
    ElMessage.error('获取用户组失败')
  }
}

const showDialog = (user = null) => {
  if (user) {
    currentUser.value = { ...user, password: '' }
  } else {
    currentUser.value = {
      username: '',
      password: '',
      full_name: '',
      email: '',
      group_id: null,
      custom_routes: '',
      custom_policies: '',
      enabled: true
    }
  }
  dialogVisible.value = true
}

const saveUser = async () => {
  try {
    if (currentUser.value.id) {
      await axios.put(`/api/users/${currentUser.value.id}`, currentUser.value)
      ElMessage.success('更新成功')
    } else {
      await axios.post('/api/users', currentUser.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const deleteUser = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除此用户吗?', '警告', {
      type: 'warning'
    })
    await axios.delete(`/api/users/${id}`)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  fetchUsers()
  fetchGroups()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
