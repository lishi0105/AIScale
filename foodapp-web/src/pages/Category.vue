<template>
  <div class="page-category">
    <aside class="category-panel">
      <div class="panel-header">
        <div class="panel-title">
          <h2>商品品类</h2>
          <p class="panel-sub">点击左侧分类查看详情</p>
        </div>
        <el-button size="small" type="primary" @click="openCreate" :disabled="!isAdmin">+ 新增品类</el-button>
      </div>
      <div class="panel-search">
        <el-input
          v-model="keyword"
          size="small"
          clearable
          placeholder="搜索品类名称/编码/拼音"
          @clear="onSearch"
          @keyup.enter="onSearch"
        />
        <el-button size="small" @click="onSearch">搜索</el-button>
      </div>
      <div class="category-list" v-loading="treeLoading">
        <el-empty v-if="!treeLoading && !categories.length" description="暂无品类" />
        <el-scrollbar v-else>
          <div
            v-for="item in cardCategories"
            :key="item.id"
            class="category-item"
            :class="{ active: item.id === selectedId }"
            @click="selectCategory(item.id)"
          >
            <div class="item-info">
              <div class="item-avatar">{{ item.avatar }}</div>
              <div class="item-text">
                <div class="item-name" :title="item.name">{{ item.name }}</div>
              </div>
            </div>
            <div class="item-meta">
              <el-dropdown
                trigger="click"
                placement="bottom-end"
                @command="(command: DropdownCommand) => onCardCommand(command, item.raw)"
              >
                <span class="item-actions" @click.stop>
                  <el-icon><MoreFilled /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit" :disabled="!isAdmin">编辑</el-dropdown-item>
                    <el-dropdown-item command="delete" :disabled="!isAdmin">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-scrollbar>
      </div>
    </aside>

    <section class="category-content">
      <div v-if="selectedCategory" class="detail-card">
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
import { CategoryAPI } from '@/api/category'
import type { CategoryListParams, CategoryRow, CategoryUpdatePayload } from '@/api/category'
import { notifyError } from '@/utils/notify'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'
import { MoreFilled } from '@element-plus/icons-vue'

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

const organId = computed(() => jwtPayload.value?.org_id || '')
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

const selectedCategory = computed(() =>
  categories.value.find(item => item.ID === selectedId.value) || null
)

const cardCategories = computed(() =>
  categories.value.map(item => ({
    id: item.ID,
    name: item.Name,
    avatar: item.Name?.charAt(0)?.toUpperCase() || '#',
    raw: item,
  }))
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
  if (!organId.value) {
    categories.value = []
    selectedId.value = ''
    return
  }
  treeLoading.value = true
  try {
    const params: CategoryListParams = {
      org_id: organId.value,
      page: 1,
      page_size: 15,
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
  () => organId.value,
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
    if (!list.some(item => item.ID === selectedId.value)) {
      const first = list[0]
      if (first) {
        selectedId.value = first.ID
      }
    }
  },
  { deep: true }
)

const selectCategory = (id: string) => {
  selectedId.value = id
}

const onCardCommand = (command: DropdownCommand, row: CategoryRow) => {
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
        org_id: organId.value,
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
  width: 280px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}


.panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.panel-title h2 {
  font-size: 18px;
  margin: 0 0 4px;
}

.panel-sub {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.panel-search {
  display: flex;
  gap: 8px;
}

.category-list {
  flex: 1;
  min-height: 0;
}

.category-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border: 1px solid transparent;
  border-radius: 10px;
  transition: all 0.2s ease;
  cursor: pointer;
  background: #f8fafc;
  margin-bottom: 10px;
}

.category-item:last-child {
  margin-bottom: 0;
}

.category-item:hover,
.category-item.active {
  border-color: #409eff;
  background: rgba(64, 158, 255, 0.08);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.12);
}

.item-info {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.item-avatar {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: linear-gradient(135deg, #409eff, #66b1ff);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  box-shadow: 0 4px 10px rgba(64, 158, 255, 0.35);
  flex-shrink: 0;
}

.item-text {
  min-width: 0;
}

.item-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-sub {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.item-actions {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  transition: background 0.2s;
  color: #606266;
}

.item-actions:hover {
  background: rgba(64, 158, 255, 0.15);
  color: #409eff;
}

.category-content {
  flex: 1;
  min-width: 0;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 20px 24px;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.detail-card {
  width: 100%;
}

.detail-title {
  margin: 0 0 16px;
  font-size: 18px;
  font-weight: 600;
}
</style>