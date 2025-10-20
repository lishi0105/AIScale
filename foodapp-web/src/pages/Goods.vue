<template>
  <div class="page-goods">
    <!-- 左侧品类面板 -->
    <aside class="category-panel">
      <div class="panel-header">
        <h3>商品品类</h3>
      </div>
      <div class="panel-search">
        <el-input
          v-model="categoryKeyword"
          size="small"
          clearable
          placeholder="搜索品类"
          @clear="onCategorySearch"
          @keyup.enter="onCategorySearch"
        />
        <el-button size="small" @click="onCategorySearch">搜索</el-button>
      </div>
      <div class="category-list" v-loading="categoryLoading">
        <el-scrollbar>
          <div
            v-for="cat in categories"
            :key="cat.ID"
            class="category-item"
            :class="{ active: cat.ID === selectedCategoryId }"
            @click="selectCategory(cat.ID)"
          >
            <div class="category-avatar">{{ cat.Name?.charAt(0) || '#' }}</div>
            <div class="category-name">{{ cat.Name }}</div>
          </div>
          <el-empty v-if="!categoryLoading && !categories.length" description="暂无品类" :image-size="80" />
        </el-scrollbar>
      </div>
    </aside>

    <!-- 右侧商品列表 -->
    <section class="goods-content">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="keyword"
            size="small"
            clearable
            placeholder="搜索商品名称/编码/拼音"
            style="width: 280px"
            @clear="onSearch"
            @keyup.enter="onSearch"
          />
          <el-button size="small" @click="onSearch">查询</el-button>
        </div>
        <div class="toolbar-right">
          <el-button size="small" type="primary" @click="openCreate" :disabled="!isAdmin">+ 新增</el-button>
          <el-button size="small" :disabled="!isAdmin">+ 导入Excel</el-button>
          <el-button size="small">导出</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <div class="table-container">
        <el-table
          :data="goodsList"
          v-loading="tableLoading"
          border
          stripe
          size="small"
          height="100%"
          :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: '600' }"
        >
          <el-table-column type="index" label="序号" width="60" align="center" />
          <el-table-column label="商品图" width="80" align="center">
            <template #default="{ row }">
              <el-avatar v-if="row.ImageURL" :src="row.ImageURL" :size="50" shape="square" />
              <el-avatar v-else :size="50" shape="square">
                <el-icon><Picture /></el-icon>
              </el-avatar>
            </template>
          </el-table-column>
          <el-table-column prop="Name" label="品名" min-width="120" show-overflow-tooltip />
          <el-table-column prop="Pinyin" label="拼音首字母" width="120" show-overflow-tooltip />
          <el-table-column prop="Code" label="编码" width="120" show-overflow-tooltip />
          <el-table-column label="规格标准" width="120" show-overflow-tooltip>
            <template #default="{ row }">
              {{ getSpecName(row.SpecID) || '—' }}
            </template>
          </el-table-column>
          <el-table-column label="品类" width="100" show-overflow-tooltip>
            <template #default="{ row }">
              {{ getCategoryName(row.CategoryID) || '—' }}
            </template>
          </el-table-column>
          <el-table-column label="验收标准" min-width="150" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.AcceptanceStandard || '—' }}
            </template>
          </el-table-column>
          <el-table-column prop="Sort" label="排序" width="80" align="center" />
          <el-table-column label="创建时间" width="160" show-overflow-tooltip>
            <template #default="{ row }">
              {{ formatDate(row.CreatedAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" align="center" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openEdit(row)" :disabled="!isAdmin">编辑</el-button>
              <el-button link type="danger" size="small" @click="confirmDelete(row)" :disabled="!isAdmin">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="onSearch"
          @current-change="onSearch"
        />
      </div>
    </section>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增商品' : '编辑商品'"
      width="600px"
      @closed="onDialogClosed"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="120px"
        v-loading="submitLoading"
      >
        <el-form-item label="商品名称" prop="name">
          <el-input v-model="form.name" maxlength="128" placeholder="请输入商品名称" clearable />
        </el-form-item>
        <el-form-item label="商品编码" prop="code">
          <el-input v-model="form.code" maxlength="64" placeholder="请输入商品编码/SKU" clearable />
        </el-form-item>
        <el-form-item label="拼音" prop="pinyin">
          <el-input v-model="form.pinyin" maxlength="128" placeholder="留空自动生成" clearable />
        </el-form-item>
        <el-form-item label="商品品类" prop="category_id">
          <el-select v-model="form.category_id" placeholder="请选择商品品类" clearable style="width: 100%">
            <el-option
              v-for="cat in categories"
              :key="cat.ID"
              :label="cat.Name"
              :value="cat.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="规格标准" prop="spec_id">
          <el-select v-model="form.spec_id" placeholder="请选择规格标准" clearable style="width: 100%">
            <el-option
              v-for="spec in specs"
              :key="spec.ID"
              :label="spec.Name"
              :value="spec.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="验收标准" prop="acceptance_standard">
          <el-input
            v-model="form.acceptance_standard"
            type="textarea"
            :rows="3"
            maxlength="512"
            placeholder="请输入验收标准"
            clearable
          />
        </el-form-item>
        <el-form-item label="商品图片URL" prop="image_url">
          <el-input v-model="form.image_url" maxlength="512" placeholder="请输入商品图片URL" clearable />
        </el-form-item>
        <el-form-item label="排序值" prop="sort">
          <el-input-number
            v-model="form.sort"
            :min="0"
            :max="999999"
            controls-position="right"
            placeholder="自动生成"
            style="width: 100%"
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
import { computed, reactive, ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import { GoodsAPI } from '@/api/goods'
import type { GoodsRow, GoodsCreatePayload, GoodsUpdatePayload } from '@/api/goods'
import { CategoryAPI } from '@/api/category'
import type { CategoryRow } from '@/api/category'
import { DictAPI } from '@/api/dict'
import { notifyError } from '@/utils/notify'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'

// JWT & 权限
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const organId = computed(() => jwtPayload.value?.org_id || '')
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

// 品类相关
const categoryKeyword = ref('')
const categories = ref<CategoryRow[]>([])
const selectedCategoryId = ref('')
const categoryLoading = ref(false)

// 规格列表
interface SpecRow {
  ID: string
  Name: string
  Sort: number
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}
const specs = ref<SpecRow[]>([])

// 商品列表
const keyword = ref('')
const goodsList = ref<GoodsRow[]>([])
const tableLoading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 对话框
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const editingGoods = ref<GoodsRow | null>(null)

interface GoodsForm {
  id: string
  name: string
  code: string
  pinyin: string
  category_id: string
  spec_id: string
  acceptance_standard: string
  image_url: string
  sort: number | null
}

const form = reactive<GoodsForm>({
  id: '',
  name: '',
  code: '',
  pinyin: '',
  category_id: '',
  spec_id: '',
  acceptance_standard: '',
  image_url: '',
  sort: null,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入商品编码', trigger: 'blur' }],
  category_id: [{ required: true, message: '请选择商品品类', trigger: 'change' }],
  spec_id: [{ required: true, message: '请选择规格标准', trigger: 'change' }],
}

// 获取品类列表
const fetchCategories = async () => {
  if (!organId.value) {
    categories.value = []
    return
  }
  categoryLoading.value = true
  try {
    const { data } = await CategoryAPI.list({
      org_id: organId.value,
      keyword: categoryKeyword.value.trim() || undefined,
      page: 1,
      page_size: 100,
    })
    categories.value = data?.items || []
  } catch (error) {
    notifyError(error)
  } finally {
    categoryLoading.value = false
  }
}

// 获取规格列表
const fetchSpecs = async () => {
  try {
    const { data } = await DictAPI.listSpecs({ page: 1, page_size: 100 })
    specs.value = data?.items || []
  } catch (error) {
    notifyError(error)
  }
}

// 获取商品列表
const fetchGoods = async () => {
  if (!organId.value) {
    goodsList.value = []
    total.value = 0
    return
  }
  tableLoading.value = true
  try {
    const { data } = await GoodsAPI.list({
      org_id: organId.value,
      keyword: keyword.value.trim() || undefined,
      category_id: selectedCategoryId.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    goodsList.value = data?.items || []
    total.value = data?.total || 0
  } catch (error) {
    notifyError(error)
  } finally {
    tableLoading.value = false
  }
}

// 选择品类
const selectCategory = (id: string) => {
  selectedCategoryId.value = id
  page.value = 1
  fetchGoods()
}

// 品类搜索
const onCategorySearch = () => {
  fetchCategories()
}

// 商品搜索
const onSearch = () => {
  page.value = 1
  fetchGoods()
}

// 打开新增对话框
const openCreate = () => {
  dialogMode.value = 'create'
  editingGoods.value = null
  resetForm()
  dialogVisible.value = true
}

// 打开编辑对话框
const openEdit = (row: GoodsRow) => {
  dialogMode.value = 'edit'
  editingGoods.value = row
  form.id = row.ID
  form.name = row.Name
  form.code = row.Code
  form.pinyin = row.Pinyin || ''
  form.category_id = row.CategoryID
  form.spec_id = row.SpecID
  form.acceptance_standard = row.AcceptanceStandard || ''
  form.image_url = row.ImageURL || ''
  form.sort = row.Sort
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  form.id = ''
  form.name = ''
  form.code = ''
  form.pinyin = ''
  form.category_id = ''
  form.spec_id = ''
  form.acceptance_standard = ''
  form.image_url = ''
  form.sort = null
}

// 对话框关闭回调
const onDialogClosed = () => {
  formRef.value?.resetFields()
  resetForm()
  editingGoods.value = null
}

// 提交表单
const onSubmit = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (dialogMode.value === 'create') {
      const payload: GoodsCreatePayload = {
        name: form.name.trim(),
        code: form.code.trim(),
        org_id: organId.value,
        category_id: form.category_id,
        spec_id: form.spec_id,
        pinyin: form.pinyin.trim() || undefined,
        acceptance_standard: form.acceptance_standard.trim() || undefined,
        image_url: form.image_url.trim() || undefined,
        sort: form.sort !== null ? form.sort : undefined,
      }
      await GoodsAPI.create(payload)
      ElMessage.success('新增商品成功')
    } else {
      const payload: GoodsUpdatePayload = {
        id: form.id,
        name: form.name.trim(),
        code: form.code.trim(),
        category_id: form.category_id,
        spec_id: form.spec_id,
        pinyin: form.pinyin.trim() || undefined,
        acceptance_standard: form.acceptance_standard.trim() || undefined,
        image_url: form.image_url.trim() || undefined,
        sort: form.sort !== null ? form.sort : undefined,
      }
      await GoodsAPI.update(payload)
      ElMessage.success('更新商品成功')
    }
    dialogVisible.value = false
    await fetchGoods()
  } catch (error) {
    notifyError(error)
  } finally {
    submitLoading.value = false
  }
}

// 删除确认
const confirmDelete = async (row: GoodsRow) => {
  try {
    await ElMessageBox.confirm(
      `确认删除商品"${row.Name}"吗？`,
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

  try {
    await GoodsAPI.softDelete(row.ID)
    ElMessage.success('删除成功')
    await fetchGoods()
  } catch (error) {
    notifyError(error)
  }
}

// 格式化日期
const formatDate = (value: string) => {
  if (!value) return '—'
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}

// 获取规格名称
const getSpecName = (specId: string) => {
  const spec = specs.value.find(s => s.ID === specId)
  return spec?.Name || ''
}

// 获取品类名称
const getCategoryName = (categoryId: string) => {
  const cat = categories.value.find(c => c.ID === categoryId)
  return cat?.Name || ''
}

// 监听 organId 变化
watch(
  () => organId.value,
  () => {
    fetchCategories()
    fetchGoods()
  },
  { immediate: true }
)

// 初始化
onMounted(() => {
  fetchSpecs()
})
</script>

<style scoped>
.page-goods {
  display: flex;
  gap: 16px;
  height: calc(100vh - 120px);
  min-height: 600px;
}

.category-panel {
  width: 260px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.panel-search {
  display: flex;
  gap: 8px;
}

.category-list {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.category-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #f8fafc;
  margin-bottom: 8px;
}

.category-item:hover,
.category-item.active {
  background: rgba(64, 158, 255, 0.1);
  border-color: #409eff;
}

.category-avatar {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #409eff, #66b1ff);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
}

.category-name {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.goods-content {
  flex: 1;
  min-width: 0;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.table-container {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
}
</style>
