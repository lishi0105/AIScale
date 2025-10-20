// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import Units from '@/pages/Units.vue'
import Specs from '@/pages/Specs.vue'
import MealTimes from '@/pages/MealTimes.vue'
import Login from '@/pages/Login.vue'
import { getToken } from '@/api/http'
import Organs from '@/pages/Organs.vue'
import Accounts from '@/pages/Accounts.vue'
import Category from '@/pages/Category.vue'
import Suppliers from '@/pages/Suppliers.vue'
import Goods from '@/pages/Goods.vue'

const routes = [
  { path: '/login', component: Login },
  { path: '/', redirect: '/dict/units' },
  { path: '/dict/units',
    component: Units,
    meta: { requiresAuth: true, section: '字典数据管理', title: '商品单位' }
  },
  { path: '/dict/specs',
    component: Specs,
    meta: { requiresAuth: true, section: '字典数据管理', title: '商品规格' }
  },
  { path: '/dict/mealtimes',
    component: MealTimes,
    meta: { requiresAuth: true, section: '字典数据管理', title: '菜单餐次' }
  },
  { path: '/base/category',
    component: Category,
    meta: { requiresAuth: true, section: '基础数据管理', title: '商品库管理' }
  },
  { path: '/base/goods',
    component: Goods,
    meta: { requiresAuth: true, section: '基础数据管理', title: '商品价格管理' }
  },
  { path: '/base/suppliers',
    component: Suppliers,
    meta: { requiresAuth: true, section: '基础数据管理', title: '供货商管理' }
  },
  { path: '/acl/orgs',
    component: Organs,
    meta: { requiresAuth: true, section: '权限管理', title: '中队管理' }
  },
  { path: '/acl/accounts',
    component: Accounts,
    meta: { requiresAuth: true, section: '权限管理', title: '账户管理' }
  },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to, from) => {
  const token = getToken()
  const isLogin = to.path === '/login'

  // 未登录：除 /login 外全部重定向到登录，并携带原目标地址
  if (!token && !isLogin) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  // 已登录又访问 /login：直接回到上一次或首页
  if (token && isLogin) {
    const target = (to.query.redirect as string) || from.fullPath || '/'
    return target
  }
})

export default router
