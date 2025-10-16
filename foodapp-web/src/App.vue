<!-- src/App.vue -->
<template>
  <!-- 登录页：全屏，无布局 -->
  <div v-if="isLoginRoute" class="login-only">
    <router-view />
  </div>

  <!-- 主应用布局 -->
  <div v-else class="layout">
    <!-- 左侧导航 -->
    <aside class="sider" :class="{ collapsed }">
      <div class="logo">
        <span v-if="!collapsed">食品品控管理系统</span>
        <span v-else class="logo-icon">🍱</span>
      </div>
      <el-menu
        class="menu"
        router
        :default-active="activeMenu"
        :collapse="collapsed"
        background-color="#1f1f1f"
        text-color="#d9d9d9"
        active-text-color="#fff"
      >
        <el-sub-menu index="acl">
          <template #title>
            <span class="menu-icon">🔐</span>
            <span>权限管理</span>
          </template>
          <el-menu-item index="/acl/accounts">
            <span class="menu-icon">👤</span>
            <span>账户管理</span>
          </el-menu-item>
          <el-menu-item index="/acl/organs">
            <span class="menu-icon">🏢</span>
            <span>中队管理</span>
          </el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="dict">
          <template #title>
            <span class="menu-icon">📚</span>
            <span>字典数据管理</span>
          </template>
          <el-menu-item index="/dict/units">
            <span class="menu-icon">⚖️</span>
            <span>商品单位</span>
          </el-menu-item>
          <el-menu-item index="/dict/specs">
            <span class="menu-icon">📦</span>
            <span>商品规格</span>
          </el-menu-item>
          <el-menu-item index="/dict/mealtimes">
            <span class="menu-icon">🍽️</span>
            <span>菜单餐次</span>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>

      <div class="sider-bottom">
        <el-button size="small" text @click="toggleCollapse">
          <span class="menu-icon">{{ collapsed ? '➕' : '➖' }}</span>
          {{ collapsed ? '展开' : '收起' }}
        </el-button>
      </div>
    </aside>

    <!-- 右侧内容区 -->
    <main class="content">
      <!-- 顶部工具条 -->
      <div class="topbar">
        <div class="topbar-left">
          <el-breadcrumb separator=">">
            <el-breadcrumb-item v-for="(c, i) in crumbs" :key="i">{{ c }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="topbar-right">
          <span class="user">欢迎，{{ usernameDisplay }}</span>
          <el-button size="small" @click="onLogout">退出登录</el-button>
        </div>
      </div>

      <!-- 页面内容 -->
      <div class="page">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { clearAuth, getToken } from '@/api/http'
import { roleLabel } from '@/utils/role'
import { parseJwt} from '@/utils/jwt'
import type { JwtPayload } from '@/utils/jwt'

const route = useRoute()
const router = useRouter()

// 面包屑
const crumbs = computed(() => {
  if (route.path.startsWith('/login')) return []
  const section = (route.meta.section as string) || ''
  const title = (route.meta.title as string) || ''
  return [section, title].filter(Boolean)
})

// 当前是否为登录页
const isLoginRoute = computed(() => route.path === '/login')

// 从 Token 解析用户信息
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})

// 显示的用户名（来自 token 的 usr 字段）
const usernameDisplay = computed(() => {
  if (!jwtPayload.value) return '未登录'
  const role = roleLabel(jwtPayload.value.role ?? 1)
  const username = jwtPayload.value.usr || '用户'
  return `${role} · ${username}`
})

// 菜单折叠状态
const collapsed = ref(false)
const toggleCollapse = () => {
  collapsed.value = !collapsed.value
}

// 激活菜单项（根据路由）
const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/acl/')) return path
  if (path.startsWith('/dict/')) return path
  return '/dict/units' // 默认高亮
})

// 退出登录
const onLogout = () => {
  clearAuth()
  router.replace('/login')
}

// 页面加载时检查登录状态
onMounted(() => {
  if (!getToken() && !isLoginRoute.value) {
    router.replace({ path: '/login', query: { redirect: route.fullPath } })
  }
})
</script>

<style scoped>
.layout {
  height: 100vh;
  display: grid;
  grid-template-columns: var(--sider-width) 1fr;
  background: #f5f7fa;
  --sider-width: 240px;
}

.layout .sider.collapsed {
  --sider-width: 64px;
}

.login-only {
  min-height: 100vh;
  background: #f5f7fa;
}

.sider {
  background: #1f1f1f;
  color: #d9d9d9;
  display: flex;
  flex-direction: column;
  transition: width 0.25s ease;
  width: var(--sider-width);
  overflow: hidden;
}

.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 16px;
  white-space: nowrap;
}

.logo-icon {
  font-size: 20px;
  line-height: 1;
}

.menu {
  border-right: none;
  flex: 1;
  padding-top: 4px;
}

.menu-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  margin-right: 6px;
}

.sider-bottom {
  padding: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  justify-content: center;
}

.content {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  height: 56px;
  padding: 0 16px;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  gap: 12px;
}

.topbar-left {
  flex: 1;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user {
  color: #666;
  font-size: 13px;
}

.page {
  padding: 16px;
  overflow: auto;
}

/* 深色菜单样式优化 */
:deep(.el-sub-menu__title:hover),
:deep(.el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.06);
}
:deep(.el-menu-item.is-active) {
  background-color: rgba(255, 255, 255, 0.1);
  color: #fff !important;
}
:deep(.el-menu--collapse) .el-sub-menu .el-sub-menu__title span,
:deep(.el-menu--collapse) .el-menu-item span {
  display: none;
}
</style>