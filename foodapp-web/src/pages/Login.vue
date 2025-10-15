<!-- src/pages/Login.vue -->
<template>
  <div class="login-container">
    <div class="login-card">
      <h1 class="title">登录</h1>

      <el-form :model="form" @keyup.enter="onSubmit" label-position="top">
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            clearable
            autocomplete="username"
            size="large"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
            clearable
            autocomplete="current-password"
            size="large"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="onSubmit"
            size="large"
            style="width: 100%"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <p v-if="error" class="error-message">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { login } from '@/api/auth'
import { ElMessage } from 'element-plus'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const route = useRoute()

const form = reactive({
  username: '',
  password: ''
})

const loading = ref(false)
const error = ref('')

// 从 localStorage 读取上次的用户名
onMounted(() => {
  const savedUsername = localStorage.getItem('lastUsername')
  if (savedUsername) {
    form.username = savedUsername
  }
})

const onSubmit = async () => {
  error.value = ''
  if (!form.username || !form.password) {
    error.value = '请输入用户名和密码'
    return
  }

  try {
    loading.value = true
    await login(form.username, form.password)
    ElMessage.success('登录成功')
    // ✅ 登录成功后保存用户名
    localStorage.setItem('lastUsername', form.username)

    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch (e: any) {
    // error.value = e.message || '登录失败，请检查用户名或密码'
    notifyError(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  padding: 32px;
  backdrop-filter: blur(10px);
  transition: transform 0.3s ease;
}

.login-card:hover {
  transform: translateY(-2px);
}

.title {
  font-size: 28px;
  font-weight: bold;
  color: #333;
  margin-bottom: 24px;
  text-align: center;
}

.el-form-item {
  margin-bottom: 20px;
}

.error-message {
  color: #e43;
  font-size: 14px;
  margin-top: 12px;
  text-align: center;
}

.el-input__wrapper {
  border-radius: 8px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
}

.el-button--primary {
  background-color: #409eff;
  border-color: #409eff;
}

.el-button--primary:hover {
  background-color: #66b1ff;
  border-color: #66b1ff;
}
</style>