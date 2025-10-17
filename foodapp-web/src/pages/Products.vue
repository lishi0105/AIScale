<!-- src/pages/Products.vue -->
<template>
  <div class="products-page">
    <div class="products-layout">
      <!-- 左侧品类树 -->
      <aside class="category-sidebar">
        <div class="sidebar-header">
          <h3>商品品类</h3>
          <el-button size="small" type="primary" @click="handleAddCategory">
            <el-icon><Plus /></el-icon>
            新增品类
          </el-button>
        </div>
        <div class="category-list">
          <el-scrollbar>
            <div
              v-for="cat in categories"
              :key="cat.ID"
              class="category-item"
              :class="{ active: selectedCategory?.ID === cat.ID }"
              @click="selectCategory(cat)"
              @contextmenu.prevent="showContextMenu($event, cat)"
            >
              <span class="category-name">{{ cat.Name }}</span>
              <span class="category-code">{{ cat.Code || '' }}</span>
            </div>
            <div v-if="categories.length === 0" class="empty-hint">
              暂无品类数据
            </div>
          </el-scrollbar>
        </div>
      </aside>

      <!-- 右侧商品列表 -->
      <main class="products-main">
        <div class="main-header">
          <h2>商品库列表</h2>
          <div class="actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索商品"
              style="width: 200px"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary">
              <el-icon><Plus /></el-icon>
              新增商品
            </el-button>
          </div>
        </div>
        <div class="products-content">
          <el-empty
            v-if="!selectedCategory"
            description="请从左侧选择品类查看商品"
          />
          <div v-else>
            <p>当前品类：{{ selectedCategory.Name }}</p>
            <!-- 这里后续添加商品表格 -->
          </div>
        </div>
      </main>
    </div>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="context-menu"
        :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
        @click="closeContextMenu"
      >
        <div class="menu-item" @click="handleEditCategory">
          <el-icon><Edit /></el-icon>
          编辑
        </div>
        <div class="menu-item danger" @click="handleDeleteCategory">
          <el-icon><Delete /></el-icon>
          删除
        </div>
      </div>
    </Teleport>

    <!-- 编辑品类弹窗 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="editingCategory ? '编辑品类' : '新增品类'"
      width="500px"
    >
      <el-form ref="editFormRef" :model="editForm" label-width="100px">
        <el-form-item label="品类名称" required>
          <el-input
            v-model="editForm.name"
            :placeholder="editingCategory?.Name || '请输入品类名称'"
          />
        </el-form-item>
        <el-form-item label="品类拼音">
          <el-input
            v-model="editForm.pinyin"
            :placeholder="editingCategory?.Pinyin || '自动生成拼音'"
          />
        </el-form-item>
        <el-form-item label="品类编码">
          <el-input
            v-model="editForm.code"
            :placeholder="editingCategory?.Code || '自动生成编码'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Edit, Delete } from '@element-plus/icons-vue'
import { CategoryAPI, type CategoryRow } from '@/api/category'
import { parseJwt } from '@/utils/jwt'
import { getToken } from '@/api/http'

// 品类列表
const categories = ref<CategoryRow[]>([])
const selectedCategory = ref<CategoryRow | null>(null)
const searchKeyword = ref('')

// 获取当前用户的 team_id
const currentTeamId = computed(() => {
  const token = getToken()
  if (!token) return ''
  const payload = parseJwt(token)
  return payload?.team_id || ''
})

// 加载品类列表
const loadCategories = async () => {
  if (!currentTeamId.value) {
    ElMessage.error('未获取到团队信息')
    return
  }
  try {
    const { data } = await CategoryAPI.list({ team_id: currentTeamId.value })
    categories.value = (data.items || []) as CategoryRow[]
  } catch (err: any) {
    ElMessage.error(err.message || '加载品类失败')
  }
}

// 选择品类
const selectCategory = (cat: CategoryRow) => {
  selectedCategory.value = cat
}

// ===== 右键菜单 =====
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextCategory = ref<CategoryRow | null>(null)

const showContextMenu = (event: MouseEvent, cat: CategoryRow) => {
  contextCategory.value = cat
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
  
  // 点击其他地方关闭菜单
  document.addEventListener('click', closeContextMenu, { once: true })
}

const closeContextMenu = () => {
  contextMenuVisible.value = false
}

// ===== 编辑品类 =====
const editDialogVisible = ref(false)
const editingCategory = ref<CategoryRow | null>(null)
const editForm = ref({
  name: '',
  pinyin: '',
  code: ''
})
const submitting = ref(false)

const handleAddCategory = () => {
  editingCategory.value = null
  editForm.value = { name: '', pinyin: '', code: '' }
  editDialogVisible.value = true
}

const handleEditCategory = () => {
  if (!contextCategory.value) return
  editingCategory.value = contextCategory.value
  editForm.value = {
    name: contextCategory.value.Name || '',
    pinyin: contextCategory.value.Pinyin || '',
    code: contextCategory.value.Code || ''
  }
  editDialogVisible.value = true
  closeContextMenu()
}

const submitEdit = async () => {
  if (!editForm.value.name.trim()) {
    ElMessage.warning('请输入品类名称')
    return
  }

  submitting.value = true
  try {
    if (editingCategory.value) {
      // 更新
      await CategoryAPI.update({
        id: editingCategory.value.ID,
        name: editForm.value.name,
        code: editForm.value.code || undefined,
        pinyin: editForm.value.pinyin || undefined
      })
      ElMessage.success('品类更新成功')
    } else {
      // 新增
      await CategoryAPI.create({
        name: editForm.value.name,
        team_id: currentTeamId.value,
        code: editForm.value.code || undefined,
        pinyin: editForm.value.pinyin || undefined
      })
      ElMessage.success('品类创建成功')
    }
    editDialogVisible.value = false
    await loadCategories()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

// ===== 删除品类 =====
const handleDeleteCategory = async () => {
  if (!contextCategory.value) return
  
  closeContextMenu()
  
  try {
    await ElMessageBox.confirm(
      `确定要删除品类"${contextCategory.value.Name}"吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await CategoryAPI.delete(contextCategory.value.ID)
    ElMessage.success('品类删除成功')
    
    // 如果删除的是当前选中的品类，清空选中
    if (selectedCategory.value?.ID === contextCategory.value.ID) {
      selectedCategory.value = null
    }
    
    await loadCategories()
  } catch (err: any) {
    if (err !== 'cancel') {
      ElMessage.error(err.message || '删除失败')
    }
  }
}

// 页面加载时获取品类列表
onMounted(() => {
  loadCategories()
})
</script>

<style scoped>
.products-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.products-layout {
  flex: 1;
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 16px;
  min-height: 0;
}

/* 左侧品类栏 */
.category-sidebar {
  background: #fff;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.category-list {
  flex: 1;
  overflow: hidden;
}

.category-item {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f3f4f6;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background-color 0.2s;
}

.category-item:hover {
  background-color: #f9fafb;
}

.category-item.active {
  background-color: #e0f2fe;
  border-left: 3px solid #0ea5e9;
}

.category-name {
  font-weight: 500;
  color: #1f2937;
}

.category-code {
  font-size: 12px;
  color: #9ca3af;
}

.empty-hint {
  padding: 32px 16px;
  text-align: center;
  color: #9ca3af;
}

/* 右侧主内容 */
.products-main {
  background: #fff;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.main-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.main-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.products-content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

/* 右键菜单 */
.context-menu {
  position: fixed;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 4px 0;
  min-width: 120px;
  z-index: 9999;
}

.menu-item {
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #374151;
  transition: background-color 0.2s;
}

.menu-item:hover {
  background-color: #f3f4f6;
}

.menu-item.danger {
  color: #ef4444;
}

.menu-item.danger:hover {
  background-color: #fef2f2;
}
</style>
