<template>
  <div class="user-groups">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户组配置</span>
          <el-button type="primary" @click="showDialog()">新增用户组</el-button>
        </div>
      </template>
      
      <el-table :data="groups" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="组名" width="150" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="routes" label="路由策略" width="200" />
        <el-table-column prop="policies" label="访问策略" width="200" />
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="showDialog(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteGroup(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="currentGroup.id ? '编辑用户组' : '新增用户组'"
      width="600px"
    >
      <el-form :model="currentGroup" label-width="100px">
        <el-form-item label="组名">
          <el-input v-model="currentGroup.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="currentGroup.description" type="textarea" />
        </el-form-item>
        <el-form-item label="路由策略">
          <el-input v-model="currentGroup.routes" placeholder="例如: 192.168.10.0/24,10.0.0.0/8" />
        </el-form-item>
        <el-form-item label="访问策略">
          <el-input v-model="currentGroup.policies" placeholder='例如: {"allow_internet":true}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveGroup">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

const groups = ref([])
const dialogVisible = ref(false)
const currentGroup = ref({
  name: '',
  description: '',
  routes: '',
  policies: ''
})

const fetchGroups = async () => {
  try {
    const response = await axios.get('/api/groups')
    groups.value = response.data.data || []
  } catch (error) {
    ElMessage.error('获取用户组失败')
  }
}

const showDialog = (group = null) => {
  if (group) {
    currentGroup.value = { ...group }
  } else {
    currentGroup.value = { name: '', description: '', routes: '', policies: '' }
  }
  dialogVisible.value = true
}

const saveGroup = async () => {
  try {
    if (currentGroup.value.id) {
      await axios.put(`/api/groups/${currentGroup.value.id}`, currentGroup.value)
      ElMessage.success('更新成功')
    } else {
      await axios.post('/api/groups', currentGroup.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchGroups()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const deleteGroup = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除此用户组吗?', '警告', {
      type: 'warning'
    })
    await axios.delete(`/api/groups/${id}`)
    ElMessage.success('删除成功')
    fetchGroups()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
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
