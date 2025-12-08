<template>
  <div class="settings">
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统设置</span>
              <el-button type="primary" @click="saveSettings" :loading="saving">保存设置</el-button>
            </div>
          </template>

          <el-form :model="settings" label-width="150px">
            <el-divider content-position="left">VPN 网络配置</el-divider>
            
            <el-form-item label="默认IP地址池">
              <el-input v-model="settings.default_ip_pool" placeholder="例如: 192.168.100.0/24" />
              <div class="form-tip">VPN客户端分配的IP地址范围，用户组可单独配置</div>
            </el-form-item>

            <el-form-item label="DNS服务器1">
              <el-input v-model="settings.default_dns1" placeholder="例如: 8.8.8.8" />
            </el-form-item>

            <el-form-item label="DNS服务器2">
              <el-input v-model="settings.default_dns2" placeholder="例如: 8.8.4.4" />
            </el-form-item>

            <el-form-item label="MTU">
              <el-input-number v-model.number="settings.default_mtu" :min="500" :max="1500" />
              <div class="form-tip">最大传输单元，建议值: 1400</div>
            </el-form-item>

            <el-divider content-position="left">连接限制</el-divider>

            <el-form-item label="最大客户端数">
              <el-input-number v-model.number="settings.max_clients" :min="1" :max="10000" />
              <div class="form-tip">允许同时连接的最大客户端数量</div>
            </el-form-item>

            <el-form-item label="空闲超时(秒)">
              <el-input-number v-model.number="settings.idle_timeout" :min="60" :max="86400" />
              <div class="form-tip">客户端空闲多久后自动断开，建议值: 3600 (1小时)</div>
            </el-form-item>

            <el-divider content-position="left">高级设置</el-divider>

            <el-form-item label="VPN域名">
              <el-input v-model="settings.vpn_domain" placeholder="例如: edge-vpn.local" />
              <div class="form-tip">VPN服务器的域名标识</div>
            </el-form-item>

            <el-form-item label="虚拟网卡名称">
              <el-input v-model="settings.vpn_device" placeholder="例如: vpns" />
              <div class="form-tip">VPN虚拟网络设备名称前缀</div>
            </el-form-item>

            <el-alert
              title="提示"
              type="warning"
              :closable="false"
              style="margin-top: 20px"
            >
              配置修改后需要重启 Edge Server 服务才能生效
            </el-alert>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card>
          <template #header>
            <span>修改密码</span>
          </template>

          <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="100px">
            <el-form-item label="原密码" prop="old_password">
              <el-input 
                v-model="passwordForm.old_password" 
                type="password" 
                show-password
                placeholder="请输入原密码"
              />
            </el-form-item>

            <el-form-item label="新密码" prop="new_password">
              <el-input 
                v-model="passwordForm.new_password" 
                type="password" 
                show-password
                placeholder="请输入新密码(至少6位)"
              />
            </el-form-item>

            <el-form-item label="确认密码" prop="confirm_password">
              <el-input 
                v-model="passwordForm.confirm_password" 
                type="password" 
                show-password
                placeholder="请再次输入新密码"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="changePassword" :loading="changingPassword" style="width: 100%">
                修改密码
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'

const settings = ref({
  default_ip_pool: '192.168.100.0/24',
  default_dns1: '8.8.8.8',
  default_dns2: '8.8.4.4',
  default_mtu: 1400,
  max_clients: 100,
  idle_timeout: 3600,
  vpn_domain: 'edge-vpn.local',
  vpn_device: 'vpns'
})

const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const passwordFormRef = ref(null)
const saving = ref(false)
const changingPassword = ref(false)

const validateConfirmPassword = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('请再次输入新密码'))
  } else if (value !== passwordForm.value.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  old_password: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const fetchSettings = async () => {
  try {
    const response = await axios.get('/api/config')
    const config = response.data.data
    
    settings.value.default_ip_pool = config.default_ip_pool || '192.168.100.0/24'
    settings.value.default_dns1 = config.default_dns1 || '8.8.8.8'
    settings.value.default_dns2 = config.default_dns2 || '8.8.4.4'
    settings.value.default_mtu = parseInt(config.default_mtu) || 1400
    settings.value.max_clients = parseInt(config.max_clients) || 100
    settings.value.idle_timeout = parseInt(config.idle_timeout) || 3600
    settings.value.vpn_domain = config.vpn_domain || 'edge-vpn.local'
    settings.value.vpn_device = config.vpn_device || 'vpns'
  } catch (error) {
    ElMessage.error('获取系统配置失败')
  }
}

const saveSettings = async () => {
  saving.value = true
  try {
    const payload = {
      default_ip_pool: settings.value.default_ip_pool,
      default_dns1: settings.value.default_dns1,
      default_dns2: settings.value.default_dns2,
      default_mtu: String(settings.value.default_mtu),
      max_clients: String(settings.value.max_clients),
      idle_timeout: String(settings.value.idle_timeout),
      vpn_domain: settings.value.vpn_domain,
      vpn_device: settings.value.vpn_device
    }
    
    await axios.put('/api/config', payload)
    ElMessage.success('配置保存成功，请重启服务使配置生效')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存配置失败')
  } finally {
    saving.value = false
  }
}

const changePassword = async () => {
  if (!passwordFormRef.value) return
  
  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    changingPassword.value = true
    try {
      await axios.post('/api/change-password', {
        old_password: passwordForm.value.old_password,
        new_password: passwordForm.value.new_password
      })
      
      ElMessage.success('密码修改成功')
      
      passwordForm.value.old_password = ''
      passwordForm.value.new_password = ''
      passwordForm.value.confirm_password = ''
      passwordFormRef.value.resetFields()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '密码修改失败')
    } finally {
      changingPassword.value = false
    }
  })
}

onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.settings {
  background-color: #fff;
}
</style>