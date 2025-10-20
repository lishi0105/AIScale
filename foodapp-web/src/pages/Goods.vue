<template>
  <div class="page-goods">
    <!-- 左侧：品类列表 -->
    <aside class="category-panel">
      <div class="panel-header">
        <h3 class="panel-title">商品品类</h3>
      </div>

      <div class="category-list" v-loading="categoryLoading">
        <el-scrollbar>
          <!-- 全部商品选项 -->
          <div
            class="category-item"
            :class="{ active: selectedCategoryId === '' }"
            @click="selectCategory('')"
          >
            <span class="category-name">全部商品</span>
          </div>
          <!-- 品类列表 -->
          <div
            v-for="item in categories"
            :key="item.ID"
            class="category-item"
            :class="{ active: item.ID === selectedCategoryId }"
          >
            <span class="category-name" @click="selectCategory(item.ID)">{{ item.Name }}</span>
            <div class="category-actions" v-if="isAdmin">
              <el-button link size="small" @click.stop="openEditCategory(item)">编辑</el-button>
              <el-button link size="small" type="danger" @click.stop="onDeleteCategory(item)">删除</el-button>
            </div>
          </div>
        </el-scrollbar>
      </div>
    </aside>

    <!-- 右侧：商品表格 -->
    <section class="goods-content">
      <h3 class="content-title">商品库列表</h3>
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="请输入"
          clearable
          @clear="onSearch"
          @keyup.enter="onSearch"
          style="width: 200px"
        />
        <el-button type="primary" @click="onSearch">查询</el-button>
        <div class="spacer" />
        <el-button type="primary" @click="openCreate" :disabled="!isAdmin">+ 新增商品</el-button>
        <el-button @click="onImportGoods" :disabled="!isAdmin">导入商品库</el-button>
      </div>

      <el-table
        :data="rows"
        stripe
        v-loading="tableLoading"
        style="width:100%"
        :header-cell-style="{ background: '#f3f4f6' }"
      >
        <el-table-column type="index" label="序号" width="70" />
        <el-table-column label="商品图" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.ImageURL"
              :src="row.ImageURL"
              fit="cover"
              style="width: 50px; height: 50px; border-radius: 4px"
              :preview-src-list="[row.ImageURL]"
            />
            <div v-else style="width: 50px; height: 50px; background: #f5f5f5; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: #ccc;">
              暂无
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="Name" label="品名" min-width="120" />
        <el-table-column label="规格标准" width="120">
          <template #default="{ row }">
            {{ specName(row.SpecID) }}
          </template>
        </el-table-column>
        <el-table-column label="单位" width="100">
          <template #default="{ row }">
            {{ unitName(row.UnitID) }}
          </template>
        </el-table-column>
        <el-table-column prop="Pinyin" label="拼音首字母代码" width="140">
          <template #default="{ row }">
            {{ row.Pinyin || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="商品描述" min-width="150">
          <template #default="{ row }">
            {{ row.AcceptanceStandard || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="Code" label="商品编码" width="120" />
        <el-table-column prop="Sort" label="排序码" width="100" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link @click="openEdit(row)" :disabled="!isAdmin">编辑</el-button>
            <el-button link type="danger" @click="onDelete(row)" :disabled="!isAdmin || deletingId===row.ID">
              <span v-if="deletingId===row.ID">删除中…</span>
              <span v-else>删除</span>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          background
          layout="prev, pager, next, jumper, ->, total"
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          @current-change="(p:number)=>{ page=p; fetchGoods(); }"
        />
      </div>
    </section>

    <!-- 弹窗：新增/编辑商品 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode==='create' ? '新增商品' : '编辑商品'" width="640px">
      <el-form :model="form" label-width="110px" v-loading="submitLoading">
        <el-form-item label="商品名称">
          <el-input v-model="form.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="编码(SKU)">
          <el-input v-model="form.code" maxlength="64" />
        </el-form-item>
        <el-form-item label="所属品类">
          <el-select v-model="form.category_id" placeholder="选择品类" style="width:100%">
            <el-option v-for="c in categories" :key="c.ID" :label="c.Name" :value="c.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="规格">
          <el-select v-model="form.spec_id" placeholder="选择规格" style="width:100%">
            <el-option v-for="s in specs" :key="s.ID" :label="s.Name" :value="s.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="单位">
          <el-select v-model="form.unit_id" placeholder="选择单位" style="width:100%">
            <el-option v-for="u in units" :key="u.ID" :label="u.Name" :value="u.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序码">
          <el-input-number v-model="form.sort" :min="0" :step="1" />
        </el-form-item>
        <el-form-item label="拼音">
          <el-input v-model="form.pinyin" maxlength="128" clearable />
        </el-form-item>
        <el-form-item label="图片URL">
          <el-input v-model="form.image_url" maxlength="512" clearable />
        </el-form-item>
        <el-form-item label="验收标准">
          <el-input v-model="form.acceptance_standard" type="textarea" :rows="3" maxlength="512" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false" :disabled="submitLoading">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 弹窗：新增/编辑品类 -->
    <el-dialog v-model="categoryDialogVisible" :title="categoryDialogMode==='create' ? '新增品类' : '编辑品类'" width="500px">
      <el-form :model="categoryForm" label-width="100px" v-loading="categorySubmitLoading">
        <el-form-item label="品类名称">
          <el-input v-model="categoryForm.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="品类编码">
          <el-input v-model="categoryForm.code" maxlength="64" />
        </el-form-item>
        <el-form-item label="拼音">
          <el-input v-model="categoryForm.pinyin" maxlength="128" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="categoryForm.sort" :min="0" :step="1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible=false" :disabled="categorySubmitLoading">取消</el-button>
        <el-button type="primary" :loading="categorySubmitLoading" @click="onSubmitCategory">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CategoryAPI, type CategoryRow, type CategoryListParams } from '@/api/category'
import { DictAPI } from '@/api/dict'
import { GoodsAPI, type GoodsRow, type GoodsCreatePayload, type GoodsUpdatePayload } from '@/api/goods'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'
import { notifyError } from '@/utils/notify'

// 登录信息
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const organId = computed(() => jwtPayload.value?.org_id || '')
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

// 左侧：品类
const categories = ref<CategoryRow[]>([])
const categoryLoading = ref(false)
const selectedCategoryId = ref('')

const selectCategory = (id: string) => {
  selectedCategoryId.value = id
  page.value = 1
  fetchGoods()
}

const fetchCategories = async () => {
  if (!organId.value) {
    categories.value = []
    selectedCategoryId.value = ''
    return
  }
  categoryLoading.value = true
  try {
    const params: CategoryListParams = { org_id: organId.value, page: 1, page_size: 100 }
    const { data } = await CategoryAPI.list(params)
    categories.value = data?.items || []
  } catch (e) {
    notifyError(e)
  } finally {
    categoryLoading.value = false
  }
}

watch(() => organId.value, () => { fetchCategories() }, { immediate: true })
watch(() => categories.value, (list: CategoryRow[]) => {
  if (!list?.length) {
    selectedCategoryId.value = '';
    return;
  }
  if (selectedCategoryId.value !== '' && !list.some(i => i.ID === selectedCategoryId.value)) {
    selectedCategoryId.value = '';
  }
}, { deep: true });

// 规格
interface SpecRow { ID: string; Name: string }
const specs = ref<SpecRow[]>([])
const fetchSpecs = async () => {
  try {
    const { data } = await DictAPI.listSpecs({ page: 1, page_size: 200 })
    specs.value = data?.items || []
  } catch (e) { /* ignore */ }
}
const specName = (id: string) => specs.value.find(s=>s.ID===id)?.Name || '—'

// 单位
interface UnitRow { ID: string; Name: string }
const units = ref<UnitRow[]>([])
const fetchUnits = async () => {
  try {
    const { data } = await DictAPI.listUnits({ page: 1, page_size: 200 })
    units.value = data?.items || []
  } catch (e) { /* ignore */ }
}
const unitName = (id: string) => units.value.find(u=>u.ID===id)?.Name || '—'

// 右侧：表格
const page = ref(1)
const pageSize = ref(15)
const total = ref(0)
const keyword = ref('')
const filterSpecId = ref<string|undefined>()
const rows = ref<GoodsRow[]>([])
const tableLoading = ref(false)
const deletingId = ref('')

const fetchGoods = async () => {
  if (!organId.value) { rows.value=[]; total.value=0; return }
  tableLoading.value = true
  try {
    const params: any = {
      org_id: organId.value,
      page: page.value,
      page_size: pageSize.value,
    }
    if (selectedCategoryId.value) params.category_id = selectedCategoryId.value
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    if (filterSpecId.value) params.spec_id = filterSpecId.value
    const { data } = await GoodsAPI.list(params)
    rows.value = data?.items || []
    total.value = Number(data?.total || 0)
  } catch (e) {
    notifyError(e)
  } finally {
    tableLoading.value = false
  }
}

const onSearch = () => { page.value = 1; fetchGoods() }

// 弹窗表单
const dialogVisible = ref(false)
const dialogMode = ref<'create'|'edit'>('create')
const submitLoading = ref(false)
const editingRow = ref<GoodsRow | null>(null)

interface GoodsForm {
  id: string
  name: string
  code: string
  category_id: string
  spec_id: string
  unit_id: string
  sort: number | null
  pinyin: string
  image_url: string
  acceptance_standard: string
}

const form = reactive<GoodsForm>({
  id: '', name: '', code: '', category_id: '', spec_id: '', unit_id: '', sort: null, pinyin: '', image_url: '', acceptance_standard: ''
})

const resetForm = () => {
  form.id=''; form.name=''; form.code='';
  form.category_id = selectedCategoryId.value || ''
  form.spec_id=''; form.unit_id=''; form.sort=null; form.pinyin=''; form.image_url=''; form.acceptance_standard=''
}

const openCreate = () => {
  dialogMode.value = 'create'
  editingRow.value = null
  resetForm()
  dialogVisible.value = true
}

const openEdit = (row: GoodsRow) => {
  dialogMode.value = 'edit'
  editingRow.value = row
  form.id = row.ID
  form.name = row.Name
  form.code = row.Code
  form.category_id = row.CategoryID
  form.spec_id = row.SpecID
  form.unit_id = row.UnitID || ''
  form.sort = row.Sort
  form.pinyin = row.Pinyin || ''
  form.image_url = row.ImageURL || ''
  form.acceptance_standard = row.AcceptanceStandard || ''
  dialogVisible.value = true
}

const onSubmit = async () => {
  const name = form.name.trim()
  const code = form.code.trim()
  if (!name || !code) { ElMessage.warning('请输入商品名称和编码'); return }
  if (!organId.value) { ElMessage.warning('缺少中队信息'); return }

  submitLoading.value = true
  try {
    if (dialogMode.value === 'create') {
      const payload: GoodsCreatePayload = {
        name, code, org_id: organId.value,
        category_id: form.category_id,
        spec_id: form.spec_id,
      }
      if (form.unit_id) payload.unit_id = form.unit_id
      if (form.sort !== null && form.sort !== undefined) payload.sort = Number(form.sort)
      if (form.pinyin.trim()) payload.pinyin = form.pinyin.trim()
      if (form.image_url.trim()) payload.image_url = form.image_url.trim()
      if (form.acceptance_standard.trim()) payload.acceptance_standard = form.acceptance_standard.trim()
      const { data } = await GoodsAPI.create(payload)
      ElMessage.success('创建成功')
      await fetchGoods()
      if (data?.ID) {
        // 切到该商品所在品类
        selectedCategoryId.value = data.CategoryID || selectedCategoryId.value
      }
    } else if (editingRow.value) {
      const payload: GoodsUpdatePayload = { id: form.id }
      if (name !== editingRow.value.Name) payload.name = name
      if (code !== editingRow.value.Code) payload.code = code
      if (form.category_id && form.category_id !== editingRow.value.CategoryID) payload.category_id = form.category_id
      if (form.spec_id && form.spec_id !== editingRow.value.SpecID) payload.spec_id = form.spec_id
      if (form.unit_id && form.unit_id !== (editingRow.value.UnitID || '')) payload.unit_id = form.unit_id
      const sortNum = form.sort === null ? null : Number(form.sort)
      if (sortNum !== null && sortNum !== editingRow.value.Sort) payload.sort = sortNum
      const pinyinTrim = form.pinyin.trim()
      if (pinyinTrim !== (editingRow.value.Pinyin || '')) payload.pinyin = pinyinTrim || null
      const imgTrim = form.image_url.trim()
      if (imgTrim !== (editingRow.value.ImageURL || '')) payload.image_url = imgTrim || null
      const asTrim = form.acceptance_standard.trim()
      if (asTrim !== (editingRow.value.AcceptanceStandard || '')) payload.acceptance_standard = asTrim || null
      if (Object.keys(payload).length === 1) { ElMessage.info('未检测到需要保存的修改'); return }
      await GoodsAPI.update(payload)
      ElMessage.success('保存成功')
      await fetchGoods()
    }
    dialogVisible.value = false
  } catch (e) {
    notifyError(e)
  } finally {
    submitLoading.value = false
  }
}

const onDelete = async (row: GoodsRow) => {
  try {
    await ElMessageBox.confirm(`确认删除 “${row.Name}” ?`, '提示', { type: 'warning' })
  } catch { return }
  deletingId.value = row.ID
  try {
    await GoodsAPI.remove(row.ID)
    ElMessage.success('删除成功')
    await fetchGoods()
  } catch (e) { notifyError(e) }
  finally { deletingId.value = '' }
}

// 品类管理
const categoryDialogVisible = ref(false)
const categoryDialogMode = ref<'create'|'edit'>('create')
const categorySubmitLoading = ref(false)
const editingCategory = ref<CategoryRow | null>(null)

interface CategoryForm {
  id: string
  name: string
  code: string
  pinyin: string
  sort: number | null
}

const categoryForm = reactive<CategoryForm>({
  id: '', name: '', code: '', pinyin: '', sort: null
})

const resetCategoryForm = () => {
  categoryForm.id = ''
  categoryForm.name = ''
  categoryForm.code = ''
  categoryForm.pinyin = ''
  categoryForm.sort = null
}

const openEditCategory = (row: CategoryRow) => {
  categoryDialogMode.value = 'edit'
  editingCategory.value = row
  categoryForm.id = row.ID
  categoryForm.name = row.Name
  categoryForm.code = row.Code || ''
  categoryForm.pinyin = row.Pinyin || ''
  categoryForm.sort = row.Sort
  categoryDialogVisible.value = true
}

const onSubmitCategory = async () => {
  const name = categoryForm.name.trim()
  if (!name) { ElMessage.warning('请输入品类名称'); return }
  if (!organId.value) { ElMessage.warning('缺少中队信息'); return }

  categorySubmitLoading.value = true
  try {
    if (categoryDialogMode.value === 'edit' && editingCategory.value) {
      const payload: any = { id: categoryForm.id, name }
      if (categoryForm.code.trim()) payload.code = categoryForm.code.trim()
      if (categoryForm.pinyin.trim()) payload.pinyin = categoryForm.pinyin.trim()
      if (categoryForm.sort !== null) payload.sort = Number(categoryForm.sort)
      await CategoryAPI.update(payload)
      ElMessage.success('保存成功')
    }
    await fetchCategories()
    categoryDialogVisible.value = false
  } catch (e) {
    notifyError(e)
  } finally {
    categorySubmitLoading.value = false
  }
}

const onDeleteCategory = async (row: CategoryRow) => {
  try {
    await ElMessageBox.confirm(`确认删除品类 "${row.Name}" ?`, '提示', { type: 'warning' })
  } catch { return }
  try {
    await CategoryAPI.remove(row.ID)
    ElMessage.success('删除成功')
    await fetchCategories()
  } catch (e) { notifyError(e) }
}

// 导入商品库
const onImportGoods = () => {
  ElMessage.info('导入功能开发中...')
}

onMounted(() => { fetchSpecs(); fetchUnits(); })
watch(() => selectedCategoryId.value, () => { page.value=1; fetchGoods() })
</script>

<style scoped>
.page-goods { display: flex; gap: 16px; height: calc(100vh - 120px); min-height: 520px; }
.category-panel { width: 240px; background: #fff; border: 1px solid #ebeef5; border-radius: 4px; padding: 0; display: flex; flex-direction: column; }
.panel-header { padding: 16px; border-bottom: 1px solid #ebeef5; }
.panel-title { font-size: 16px; font-weight: 600; margin: 0; color: #303133; }
.category-list { flex: 1; min-height: 0; padding: 8px 0; }
.category-item { 
  display: flex; 
  align-items: center; 
  justify-content: space-between; 
  padding: 12px 20px; 
  transition: all .2s ease; 
  cursor: pointer;
  border-left: 3px solid transparent;
}
.category-item:hover { background: #f5f7fa; }
.category-item.active { 
  background: #ecf5ff; 
  border-left-color: #409eff;
}
.category-name { 
  flex: 1;
  font-size: 14px; 
  color: #606266;
  white-space: nowrap; 
  overflow: hidden; 
  text-overflow: ellipsis;
}
.category-item.active .category-name { 
  color: #409eff;
  font-weight: 500;
}
.category-actions { 
  display: flex; 
  gap: 4px;
  opacity: 0;
  transition: opacity .2s;
}
.category-item:hover .category-actions { 
  opacity: 1;
}
.goods-content { flex: 1; min-width: 0; background: #fff; border: 1px solid #ebeef5; border-radius: 4px; padding: 16px; display: flex; flex-direction: column; }
.content-title { font-size: 16px; font-weight: 600; margin: 0 0 16px; color: #303133; }
.toolbar { display: flex; gap: 12px; align-items: center; margin-bottom: 16px; }
.spacer { flex: 1 }
.pager { display: flex; justify-content: flex-end; padding-top: 12px; }
</style>
