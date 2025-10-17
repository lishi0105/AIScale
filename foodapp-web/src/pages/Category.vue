<template>
  <div class="page-category">
    <aside class="category-panel">
      <div class="panel-header">
        <h2>商品品类</h2>
        <el-button size="small" type="primary" @click="openCreate" :disabled="!isAdmin">+ 新增商品</el-button>
      </div>
      <div class="panel-search">
        <el-input
          v-model="keyword"
          size="small"
          clearable
          placeholder="请输入"
          @clear="onSearch"
          @keyup.enter="onSearch"
        >
          <template #suffix>
            <el-button link @click="onSearch">查询</el-button>
          </template>
        </el-input>
      </div>
      <div class="panel-list" v-loading="treeLoading">
        <!-- 全部商品选项 -->
        <div 
          class="category-item all-item"
          :class="{ active: selectedId === '' }"
          @click="onSelectAll"
        >
          <span class="item-name">全部商品</span>
        </div>
        <!-- 品类列表 -->
        <div
          v-for="item in categories"
          :key="item.ID"
          class="category-item"
          :class="{ active: item.ID === selectedId }"
          @click="onNodeClick(item)"
        >
          <span class="item-name">{{ item.Name }}</span>
          <el-dropdown
            v-if="isAdmin"
            trigger="click"
            @command="(command: DropdownCommand) => onNodeCommand(command, item)"
            @click.stop
          >
            <el-button link class="item-more">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">编辑</el-dropdown-item>
                <el-dropdown-item command="delete">删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        <el-empty v-if="!categories.length && !treeLoading" description="暂无品类" :image-size="80" />
      </div>
    </aside>

    <section class="category-content">
      <div v-if="!selectedId" class="detail-card">
        <h3 class="detail-title">全部商品</h3>
        <div class="all-products-info">
          <p>当前共有 <strong>{{ categories.length }}</strong> 个商品品类</p>
        </div>
      </div>
      <div v-else-if="selectedCategory" class="detail-card">
        <h3 class="detail-title">品类详情</h3>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="品类名称">{{ selectedCategory.Name }}</el-descriptions-item>
          <el-descriptions-item label="拼音">
            {{ selectedCategory.Pinyin || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="编码">
            {{ selectedCategory.Code || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="排序值">{{ selectedCategory.Sort }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(selectedCategory.CreatedAt) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(selectedCategory.UpdatedAt) }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <el-empty v-else description="请选择左侧品类" />
    </section>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增品类' : '编辑品类'"
      width="520px"
      @closed="onDialogClosed"
    >
      <el-form label-width="96px" :model="form" v-loading="submitLoading">
        <el-form-item label="品类名称">
          <el-input
            v-model="form.name"
            maxlength="64"
            :placeholder="namePlaceholder"
          />
        </el-form-item>
        <el-form-item label="拼音">
          <el-input
            v-model="form.pinyin"
            maxlength="64"
            :placeholder="pinyinPlaceholder"
            clearable
          />
        </el-form-item>
        <el-form-item label="编码">
          <el-input
            v-model="form.code"
            maxlength="64"
            :placeholder="codePlaceholder"
            clearable
          />
        </el-form-item>
        <el-form-item label="排序值">
          <el-input-number
            v-model="form.sort"
            :min="0"
            :max="999999"
            controls-position="right"
            :placeholder="sortPlaceholder"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false" :disabled="submitLoading">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled } from '@element-plus/icons-vue'
import { CategoryAPI } from '@/api/category'
import type { CategoryListParams, CategoryRow, CategoryUpdatePayload } from '@/api/category'
import { notifyError } from '@/utils/notify'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'

const keyword = ref('')
const categories = ref<CategoryRow[]>([])
const selectedId = ref('')
const treeLoading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const deletingId = ref('')
const editingCategory = ref<CategoryRow | null>(null)

interface CategoryForm {
  id: string
  name: string
  pinyin: string
  code: string
  sort: number | null
}

const form = reactive<CategoryForm>({
  id: '',
  name: '',
  pinyin: '',
  code: '',
  sort: null,
})

type DropdownCommand = 'edit' | 'delete'

const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})

const teamId = computed(() => jwtPayload.value?.team_id || '')
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

const selectedCategory = computed(() =>
  categories.value.find(item => item.ID === selectedId.value) || null
)

const namePlaceholder = computed(() =>
  dialogMode.value === 'edit'
    ? editingCategory.value?.Name || '请输入品类名称'
    : '请输入品类名称'
)
const pinyinPlaceholder = computed(() =>
  dialogMode.value === 'edit'
    ? editingCategory.value?.Pinyin || '留空自动生成'
    : '留空自动生成'
)
const codePlaceholder = computed(() =>
  dialogMode.value === 'edit'
    ? editingCategory.value?.Code || '留空自动生成'
    : '留空自动生成'
)
const sortPlaceholder = computed(() =>
  dialogMode.value === 'edit'
    ? String(editingCategory.value?.Sort ?? '') || '自动生成'
    : '自动生成'
)

const resetForm = () => {
  form.id = ''
  form.name = ''
  form.pinyin = ''
  form.code = ''
  form.sort = null
}

const onDialogClosed = () => {
  resetForm()
  editingCategory.value = null
}

const fetchCategories = async () => {
  if (!teamId.value) {
    categories.value = []
    selectedId.value = ''
    return
  }
  treeLoading.value = true
  try {
    const params: CategoryListParams = {
      team_id: teamId.value,
      page: 1,
      page_size: 999,
    }
    if (keyword.value.trim()) {
      params.keyword = keyword.value.trim()
    }
    const { data } = await CategoryAPI.list(params)
    categories.value = data?.items || []
  } catch (error) {
    notifyError(error)
  } finally {
    treeLoading.value = false
  }
}

watch(
  () => teamId.value,
  () => {
    fetchCategories()
  },
  { immediate: true }
)

watch(
  () => categories.value,
  (list: CategoryRow[]) => {
    if (!list.length) {
      selectedId.value = ''
      return
    }
    // 如果当前选中的品类不在列表中，默认选中"全部商品"
    if (selectedId.value && !list.some(item => item.ID === selectedId.value)) {
      selectedId.value = ''
    }
  },
  { deep: true }
)

const onSelectAll = () => {
  selectedId.value = ''
}

const onNodeClick = (item: CategoryRow) => {
  selectedId.value = item.ID
}

const onNodeCommand = (command: DropdownCommand, row: CategoryRow) => {
  selectedId.value = row.ID
  if (command === 'edit') {
    openEdit(row)
  } else if (command === 'delete') {
    confirmDelete(row)
  }
}

const openCreate = () => {
  dialogMode.value = 'create'
  editingCategory.value = null
  resetForm()
  dialogVisible.value = true
}

const openEdit = (row: CategoryRow) => {
  dialogMode.value = 'edit'
  editingCategory.value = row
  form.id = row.ID
  form.name = row.Name
  form.pinyin = row.Pinyin || ''
  form.code = row.Code || ''
  form.sort = row.Sort
  dialogVisible.value = true
}

const onSearch = () => {
  fetchCategories()
}

const formatDate = (value: string) => {
  if (!value) return '—'
  try {
    return new Date(value).toLocaleString()
  } catch (error) {
    return value
  }
}

const onSubmit = async () => {
  const name = form.name.trim()
  if (!name) {
    ElMessage.warning('请输入品类名称')
    return
  }
  submitLoading.value = true
  try {
    if (dialogMode.value === 'create') {
      const { data } = await CategoryAPI.create({
        name,
        team_id: teamId.value,
        pinyin: form.pinyin.trim() ? form.pinyin.trim() : undefined,
        code: form.code.trim() ? form.code.trim() : undefined,
      })
      ElMessage.success('新增品类成功')
      await fetchCategories()
      if (data?.ID) {
        selectedId.value = data.ID
      }
    } else {
      const payload: CategoryUpdatePayload = {
        id: form.id,
        name,
      }
      if (form.pinyin.trim() || editingCategory.value?.Pinyin) {
        payload.pinyin = form.pinyin.trim() ? form.pinyin.trim() : null
      }
      if (form.code.trim() || editingCategory.value?.Code) {
        payload.code = form.code.trim() ? form.code.trim() : null
      }
      if (form.sort !== null && form.sort !== undefined) {
        payload.sort = Number(form.sort)
      }
      await CategoryAPI.update(payload)
      ElMessage.success('更新品类成功')
      await fetchCategories()
      selectedId.value = form.id
    }
    dialogVisible.value = false
  } catch (error) {
    notifyError(error)
  } finally {
    submitLoading.value = false
  }
}

const confirmDelete = async (row: CategoryRow) => {
  if (deletingId.value) return
  try {
    await ElMessageBox.confirm(
      `确认删除品类“${row.Name}”吗？`,
      '删除确认',
      {
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch {
    return
  }

  deletingId.value = row.ID
  try {
    await CategoryAPI.remove(row.ID)
    ElMessage.success('删除成功')
    if (selectedId.value === row.ID) {
      selectedId.value = ''
    }
    await fetchCategories()
  } catch (error) {
    notifyError(error)
  } finally {
    deletingId.value = ''
  }
}
</script>

<style scoped>
.page-category {
  display: flex;
  gap: 16px;
  height: calc(100vh - 120px);
  min-height: 520px;
}

.category-panel {
  width: 320px;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 0;
  display: flex;
  flex-direction: column;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #e5e7eb;
}

.panel-header h2 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #1f2937;
}

.panel-search {
  padding: 12px 16px;
  border-bottom: 1px solid #e5e7eb;
}

.panel-search :deep(.el-input) {
  width: 100%;
}

.panel-search :deep(.el-input__suffix) {
  padding-right: 4px;
}

.panel-search :deep(.el-button) {
  font-size: 13px;
  padding: 0 8px;
  color: #409eff;
}

.panel-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 8px 0;
}

.category-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  margin: 2px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.category-item:hover {
  background: #f3f4f6;
}

.category-item.active {
  background: #3b82f6;
  color: #fff;
  font-weight: 500;
}

.category-item.active .item-name {
  color: #fff;
}

.category-item .item-name {
  flex: 1;
  font-size: 14px;
  color: #374151;
  transition: color 0.2s ease;
}

.category-item.active:hover {
  background: #2563eb;
}

.category-item .item-more {
  padding: 4px;
  font-size: 16px;
  color: #9ca3af;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.category-item:hover .item-more {
  opacity: 1;
}

.category-item.active .item-more {
  color: #fff;
  opacity: 1;
}

.category-item.all-item {
  margin-bottom: 4px;
}

.category-content {
  flex: 1;
  min-width: 0;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.detail-card {
  width: 100%;
}

.detail-title {
  margin: 0 0 20px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.all-products-info {
  padding: 20px;
  background: #f9fafb;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
}

.all-products-info p {
  margin: 0;
  font-size: 14px;
  color: #6b7280;
}

.all-products-info strong {
  color: #3b82f6;
  font-size: 16px;
}

/* 滚动条样式 */
.panel-list::-webkit-scrollbar {
  width: 6px;
}

.panel-list::-webkit-scrollbar-track {
  background: transparent;
}

.panel-list::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.panel-list::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
</style>
