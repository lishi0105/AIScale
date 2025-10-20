<template>
  <div class="goods-demo-page">
    <header class="goods-header">
      <div class="header-left">
        <div class="brand">食品库管理系统</div>
        <div class="sub">商品库管理</div>
      </div>
      <div class="header-right">
        <el-input
          v-model="headerSearch"
          placeholder="搜索"
          :prefix-icon="Search"
          size="small"
          class="header-search"
          clearable
        />
        <el-icon><Bell /></el-icon>
        <el-avatar size="small" class="header-avatar">宁</el-avatar>
      </div>
    </header>

    <div class="goods-body">
      <aside class="goods-sidebar">
        <nav>
          <div v-for="group in navGroups" :key="group.title" class="sidebar-group">
            <div class="sidebar-title">{{ group.title }}</div>
            <ul>
              <li
                v-for="item in group.items"
                :key="item"
                :class="{ active: activeNav === item }"
                @click="activeNav = item"
              >
                <span>{{ item }}</span>
              </li>
            </ul>
          </div>
        </nav>
      </aside>

      <main class="goods-main">
        <div class="goods-breadcrumb">当前位置：基础资料 / 商品库管理</div>
        <section class="goods-card">
          <header class="card-header">
            <div class="card-title">
              <h1>商品库管理</h1>
              <p>管理商品品类、规格和基础信息示例页面</p>
            </div>
            <div class="card-actions">
              <el-button type="primary">导入商品</el-button>
              <el-button>导出模板</el-button>
            </div>
          </header>

          <div class="card-toolbar">
            <el-input v-model="keyword" placeholder="搜索商品名称/拼音" clearable />
            <el-select v-model="selectedCategory" placeholder="商品品类" clearable>
              <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
            </el-select>
            <el-select v-model="selectedStatus" placeholder="使用状态" clearable>
              <el-option label="使用中" value="使用中" />
              <el-option label="停用" value="停用" />
            </el-select>
            <el-button type="primary">搜索</el-button>
            <el-button>重置</el-button>
            <div class="toolbar-spacer" />
            <el-button type="primary">+ 新增商品</el-button>
          </div>

          <div class="table-wrapper">
            <table class="goods-table">
              <thead>
                <tr>
                  <th class="check-col">
                    <el-checkbox
                      :model-value="allChecked"
                      :indeterminate="indeterminate"
                      @change="toggleAll"
                    />
                  </th>
                  <th>序号</th>
                  <th>商品名称</th>
                  <th>所属品类</th>
                  <th>规格</th>
                  <th>拼音首字母</th>
                  <th>排序值</th>
                  <th>使用状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, index) in demoGoods" :key="row.id">
                  <td class="check-col"><el-checkbox v-model="row.checked" /></td>
                  <td>{{ index + 1 }}</td>
                  <td>
                    <div class="goods-name">
                      <span class="badge">{{ row.type }}</span>
                      <div class="name-text">{{ row.name }}</div>
                    </div>
                  </td>
                  <td>{{ row.category }}</td>
                  <td>{{ row.spec }}</td>
                  <td>{{ row.pinyin }}</td>
                  <td>{{ row.sort }}</td>
                  <td>
                    <el-tag :type="row.status === '使用中' ? 'success' : 'info'" size="small">
                      {{ row.status }}
                    </el-tag>
                  </td>
                  <td>
                    <el-button link>编辑</el-button>
                    <el-button link type="danger">删除</el-button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <footer class="card-footer">
            <div class="footer-info">已选择 {{ selectedCount }} 个商品，共 {{ demoGoods.length }} 个商品</div>
            <el-pagination
              layout="prev, pager, next, sizes, ->, total"
              :total="demoGoods.length"
              :page-size="10"
              small
            />
          </footer>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { Bell, Search } from '@element-plus/icons-vue'

const headerSearch = ref('')
const keyword = ref('')
const selectedCategory = ref('')
const selectedStatus = ref('')
const activeNav = ref('商品库管理')

const navGroups = reactive([
  {
    title: '集采管理',
    items: ['采购订单', '入库管理', '出库管理', '库存盘点', '报表分析'],
  },
  {
    title: '基础资料',
    items: ['供应商管理', '商品库管理', '单位设置', '仓库设置', '人员设置'],
  },
  {
    title: '系统配置',
    items: ['组织管理', '角色权限', '日志审计'],
  },
])

const categories = ['蔬菜类', '肉禽类', '水产类', '干货类', '调味品']

interface DemoGoodsRow {
  id: number
  name: string
  category: string
  spec: string
  pinyin: string
  sort: number
  status: string
  type: string
  checked: boolean
}

const demoGoods = reactive<DemoGoodsRow[]>([
  { id: 1, name: '大白菜', category: '蔬菜类', spec: '500g/袋', pinyin: 'd', sort: 1001, status: '使用中', type: 'F', checked: false },
  { id: 2, name: '土豆', category: '蔬菜类', spec: '500g/袋', pinyin: 't', sort: 1002, status: '使用中', type: 'F', checked: false },
  { id: 3, name: '西红柿', category: '蔬菜类', spec: '500g/袋', pinyin: 'x', sort: 1003, status: '使用中', type: 'F', checked: false },
  { id: 4, name: '大葱', category: '蔬菜类', spec: '500g/袋', pinyin: 'd', sort: 1004, status: '使用中', type: 'F', checked: false },
  { id: 5, name: '青椒', category: '蔬菜类', spec: '500g/袋', pinyin: 'q', sort: 1005, status: '使用中', type: 'F', checked: false },
  { id: 6, name: '胡萝卜', category: '蔬菜类', spec: '500g/袋', pinyin: 'h', sort: 1006, status: '使用中', type: 'F', checked: false },
  { id: 7, name: '黄瓜', category: '蔬菜类', spec: '500g/袋', pinyin: 'h', sort: 1007, status: '使用中', type: 'F', checked: false },
  { id: 8, name: '香菜', category: '蔬菜类', spec: '500g/袋', pinyin: 'x', sort: 1008, status: '使用中', type: 'F', checked: false },
  { id: 9, name: '香葱', category: '蔬菜类', spec: '500g/袋', pinyin: 'x', sort: 1009, status: '使用中', type: 'F', checked: false },
  { id: 10, name: '菠菜', category: '蔬菜类', spec: '500g/袋', pinyin: 'b', sort: 1010, status: '使用中', type: 'F', checked: false },
  { id: 11, name: '韭菜', category: '蔬菜类', spec: '500g/袋', pinyin: 'j', sort: 1011, status: '使用中', type: 'F', checked: false },
  { id: 12, name: '冬瓜', category: '蔬菜类', spec: '500g/袋', pinyin: 'd', sort: 1012, status: '使用中', type: 'F', checked: false },
  { id: 13, name: '娃娃菜', category: '蔬菜类', spec: '500g/袋', pinyin: 'w', sort: 1013, status: '使用中', type: 'F', checked: false },
  { id: 14, name: '南瓜', category: '蔬菜类', spec: '500g/袋', pinyin: 'n', sort: 1014, status: '使用中', type: 'F', checked: false },
  { id: 15, name: '生菜', category: '蔬菜类', spec: '500g/袋', pinyin: 's', sort: 1015, status: '使用中', type: 'F', checked: false },
])

const selectedCount = computed(() => demoGoods.filter(item => item.checked).length)
const allChecked = computed(() => demoGoods.length > 0 && demoGoods.every(item => item.checked))
const indeterminate = computed(() => selectedCount.value > 0 && !allChecked.value)

const toggleAll = (val: boolean) => {
  const checked = Boolean(val)
  demoGoods.forEach(row => {
    row.checked = checked
  })
}
</script>

<style scoped>
:deep(.el-input__wrapper) {
  border-radius: 6px;
}

.goods-demo-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f5f7fa;
  color: #303133;
  font-size: 14px;
}

.goods-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: #1f2937;
  color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.brand {
  font-size: 18px;
  font-weight: 600;
}

.sub {
  font-size: 12px;
  opacity: 0.85;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-search {
  width: 200px;
}

.header-avatar {
  background: #409eff;
  color: #fff;
  font-size: 12px;
}

.goods-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.goods-sidebar {
  width: 220px;
  background: #111827;
  color: #d1d5db;
  padding: 24px 16px;
  overflow-y: auto;
}

.sidebar-group + .sidebar-group {
  margin-top: 24px;
}

.sidebar-title {
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #9ca3af;
  margin-bottom: 8px;
}

.goods-sidebar ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.goods-sidebar li {
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background 0.2s ease, color 0.2s ease;
}

.goods-sidebar li:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.goods-sidebar li.active {
  background: #2563eb;
  color: #fff;
}

.goods-main {
  flex: 1;
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: auto;
}

.goods-breadcrumb {
  color: #6b7280;
  font-size: 13px;
}

.goods-card {
  background: #fff;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.08);
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.card-title h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
}

.card-title p {
  margin: 4px 0 0;
  color: #6b7280;
}

.card-actions {
  display: flex;
  gap: 12px;
}

.card-toolbar {
  display: grid;
  grid-template-columns: 220px 160px 160px 100px 100px auto 140px;
  gap: 12px;
  align-items: center;
}

.toolbar-spacer {
  flex: 1;
}

.table-wrapper {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
}

.goods-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
}

.goods-table thead {
  background: #f9fafb;
  color: #4b5563;
}

.goods-table th,
.goods-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #f1f5f9;
}

.goods-table tbody tr:hover {
  background: #f8fafc;
}

.goods-name {
  display: flex;
  align-items: center;
  gap: 10px;
}

.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 10px;
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
  color: #fff;
  font-weight: 600;
}

.name-text {
  font-weight: 500;
  color: #111827;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.footer-info {
  color: #6b7280;
}

.check-col {
  width: 56px;
}

@media (max-width: 1280px) {
  .card-toolbar {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }
}
</style>
