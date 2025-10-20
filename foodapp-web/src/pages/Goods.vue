<template>
  <div class="page-goods">
    <!-- 左侧：品类列表 -->
    <aside class="category-panel">
      <div class="panel-title-line">商品品类</div>
      <div class="panel-search">
        <el-input
          v-model="categoryKeyword"
          size="small"
          clearable
          placeholder="搜索品类名称/拼音/编码"
          @clear="fetchCategories"
          @keyup.enter="fetchCategories"
        />
        <el-button size="small" @click="fetchCategories">搜索</el-button>
      </div>

      <div class="category-list" v-loading="categoryLoading">
        <div
          class="category-row"
          :class="{ active: selectedCategoryId === '' }"
          @click="selectCategory('')"
        >
          <span class="name">全部商品</span>
        </div>

        <el-empty v-if="!categoryLoading && !categories.length" description="暂无品类" />
        <el-scrollbar v-else>
          <div
            v-for="cat in categories"
            :key="cat.ID"
            class="category-row"
            :class="{ active: cat.ID === selectedCategoryId }"
            @click="selectCategory(cat.ID)"
          >
            <span class="name" :title="cat.Name">{{ cat.Name }}</span>
            <div class="ops" @click.stop>
              <el-dropdown trigger="click">
                <span class="more">···</span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click.stop="openEditCategory(cat)">编辑</el-dropdown-item>
                    <el-dropdown-item divided type="danger" @click.stop="onDeleteCategory(cat)">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-scrollbar>
      </div>
      <div class="category-footer">
        <el-button size="small" type="primary" plain @click="openCreateCategory" :disabled="!isAdmin">
          + 新增品类
        </el-button>
      </div>
    </aside>

    <!-- 右侧：商品表格 -->
    <section class="goods-content">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="请输入"
          clearable
          @clear="onSearch"
          @keyup.enter="onSearch"
          style="width: 280px"
        />
        <el-button @click="onSearch">查询</el-button>
        <div class="spacer" />
        <el-button @click="onImport" plain>导入商品库</el-button>
        <el-button type="primary" @click="openCreate" :disabled="!isAdmin">+ 新增商品</el-button>
      </div>

      <el-table
        :data="rows"
        stripe
        v-loading="tableLoading"
        style="width:100%"
        :header-cell-style="{ background: '#f3f4f6' }"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column type="index" label="序号" width="70" />
        <el-table-column label="商品图" width="90">
          <template #default="{ row }">
            <el-image v-if="row.ImageURL" :src="row.ImageURL" fit="cover" style="width:48px;height:48px;border-radius:6px" />
            <div v-else class="img-ph">无</div>
          </template>
        </el-table-column>
        <el-table-column prop="Name" label="品名" min-width="140" />
        <el-table-column label="规格标准" width="120">
          <template #default="{ row }">{{ specName(row.SpecID) }}</template>
        </el-table-column>
        <el-table-column label="单位" width="100">
          <template #default="{ row }">{{ unitName(row.UnitID) }}</template>
        </el-table-column>
        <el-table-column prop="Pinyin" label="拼音首字母代码" width="160">
          <template #default="{ row }">{{ row.Pinyin || '—' }}</template>
        </el-table-column>
        <el-table-column label="验收标准" min-width="280" show-overflow-tooltip>
          <template #default="{ row }">{{ row.AcceptanceStandard || '—' }}</template>
        </el-table-column>
        <el-table-column prop="Code" label="商品编码" width="140" />
        <el-table-column prop="Sort" label="排序码" width="100" />
        <el-table-column label="操作" width="160" fixed="right">
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
          layout="sizes, prev, pager, next, jumper, ->, total"
          :page-sizes="pageSizes"
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </section>

    <!-- 弹窗：新增/编辑商品 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode==='create' ? '新增商品' : '编辑商品'" width="640px">
      <el-form :model="form" label-width="110px" v-loading="submitLoading" class="goods-form">
        <el-form-item label="商品名称">
          <div class="field-inline">
            <el-input v-model="form.name" maxlength="128" />
            <span class="required-mark">*</span>
          </div>
        </el-form-item>
        <el-form-item label="编码(SKU)">
          <div class="field-inline">
            <el-input v-model="form.code" placeholder="缺省自动生成" maxlength="64" />
          </div>
        </el-form-item>
        <el-form-item label="所属品类">
          <div class="field-inline">
            <el-select v-model="form.category_id" placeholder="选择品类" style="width:100%">
              <el-option v-for="c in categories" :key="c.ID" :label="c.Name" :value="c.ID" />
            </el-select>
            <span class="required-mark">*</span>
          </div>
        </el-form-item>
        <el-form-item label="规格">
          <div class="field-inline">
            <el-select v-model="form.spec_id" placeholder="选择规格" style="width:100%">
              <el-option v-for="s in specs" :key="s.ID" :label="s.Name" :value="s.ID" />
            </el-select>
            <span class="required-mark">*</span>
          </div>
        </el-form-item>
        <el-form-item label="单位">
          <div class="field-inline">
            <el-select v-model="form.unit_id" placeholder="选择单位" style="width:100%">
              <el-option v-for="u in units" :key="u.ID" :label="u.Name" :value="u.ID" />
            </el-select>
            <span class="required-mark">*</span>
          </div>
        </el-form-item>
        <el-form-item label="排序码">
          <div class="field-inline">
            <el-input-number v-model="form.sort" placeholder="缺省自动生成" :min="0" :step="1" />
          </div>
        </el-form-item>
        <el-form-item label="拼音">
          <div class="field-inline">
            <el-input v-model="form.pinyin" placeholder="缺省自动生成" maxlength="128" clearable />
          </div>
        </el-form-item>
        <el-form-item label="图片">
          <el-input v-model="form.image_url" maxlength="512" clearable />
        </el-form-item>
        <el-form-item label="验收标准">
          <el-input v-model="form.acceptance_standard" placeholder="可选" type="textarea" :rows="3" maxlength="512" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false" :disabled="submitLoading">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>
    <!-- 弹窗：编辑品类 -->
    <el-dialog v-model="catDialogVisible" :title="catDialogTitle" width="420px">
      <el-form :model="catForm" label-width="80px" v-loading="catSubmitLoading">
        <el-form-item label="品类名称">
          <el-input v-model="catForm.name" maxlength="64" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="catDialogVisible=false" :disabled="catSubmitLoading">取消</el-button>
        <el-button type="primary" :loading="catSubmitLoading" @click="onSubmitCategory">保存</el-button>
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
const categoryKeyword = ref('')
const categories = ref<CategoryRow[]>([])
const categoryLoading = ref(false)
const selectedCategoryId = ref('')

// 额外的磁贴样式已移除，直接展示列表

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
    if (categoryKeyword.value.trim()) params.keyword = categoryKeyword.value.trim()
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
    selectedCategoryId.value = ''
    return
  }
  // 如果当前选择为空（全部商品），保持为空；否则如果选中的品类不在列表中，回退到第一个
  if (!selectedCategoryId.value) return
  if (!list.some(i => i.ID === selectedCategoryId.value)) {
    selectedCategoryId.value = list[0]!.ID
  }
}, { deep: true })

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
const pageSizes = [10, 15, 20, 50]
const total = ref(0)
const keyword = ref('')
const filterSpecId = ref<string|undefined>()
const rows = ref<GoodsRow[]>([])
const tableLoading = ref(false)
const deletingId = ref('')

const handlePageChange = (p: number) => {
  page.value = p
  fetchGoods()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  page.value = 1
  fetchGoods()
}

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

// 导入
const onImport = () => {
  ElMessage.info('导入功能暂未开放')
}

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
  form.unit_id = row.UnitID
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
  if (!form.category_id) { ElMessage.warning('请选择品类'); return }
  if (!form.spec_id) { ElMessage.warning('请选择规格'); return }
  if (!form.unit_id) { ElMessage.warning('请选择单位'); return }
  if (!organId.value) { ElMessage.warning('缺少中队信息'); return }

  submitLoading.value = true
  try {
    if (dialogMode.value === 'create') {
      const payload: GoodsCreatePayload = {
        name,
        code,
        org_id: organId.value,
        category_id: form.category_id,
        spec_id: form.spec_id,
        unit_id: form.unit_id,
      }
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
      if (form.unit_id && form.unit_id !== editingRow.value.UnitID) payload.unit_id = form.unit_id
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

// 品类编辑/删除
const catDialogVisible = ref(false)
const catSubmitLoading = ref(false)
const catEditing = ref<CategoryRow | null>(null)
const catDialogTitle = computed(() => (catEditing.value ? '编辑品类' : '新增品类'))
const catForm = reactive<{ id: string; name: string }>({ id: '', name: '' })

const openCreateCategory = () => {
  catEditing.value = null
  catForm.id = ''
  catForm.name = ''
  catDialogVisible.value = true
}

const openEditCategory = (row: CategoryRow) => {
  catEditing.value = row
  catForm.id = row.ID
  catForm.name = row.Name
  catDialogVisible.value = true
}

const onSubmitCategory = async () => {
  const name = catForm.name.trim()
  if (!name) { ElMessage.warning('请输入品类名称'); return }
  catSubmitLoading.value = true
  try {
    let createdId: string | undefined
    if (catEditing.value) {
      await CategoryAPI.update({ id: catForm.id, name })
      ElMessage.success('保存成功')
    } else {
      if (!organId.value) { ElMessage.warning('缺少组织信息'); return }
      const { data } = await CategoryAPI.create({ name, org_id: organId.value })
      ElMessage.success('新增成功')
      createdId = data?.ID
    }
    catDialogVisible.value = false
    await fetchCategories()
    if (createdId) {
      selectedCategoryId.value = createdId
    }
  } catch (e) { notifyError(e) }
  finally { catSubmitLoading.value = false }
}

const onDeleteCategory = async (row: CategoryRow) => {
  try {
    await ElMessageBox.confirm(`确认删除品类 “${row.Name}” ?`, '提示', { type: 'warning' })
  } catch { return }
  try {
    await CategoryAPI.remove(row.ID)
    ElMessage.success('删除成功')
    if (selectedCategoryId.value === row.ID) selectedCategoryId.value = ''
    await fetchCategories()
  } catch (e) { notifyError(e) }
}

onMounted(() => { fetchSpecs(); fetchUnits(); })
watch(() => organId.value, () => { fetchGoods() }, { immediate: true })
watch(() => selectedCategoryId.value, () => { page.value=1; fetchGoods() }, { immediate: true })
</script>

<style scoped>
.page-goods { display: flex; gap: 16px; height: calc(100vh - 120px); min-height: 520px; }
.category-panel { width: 260px; background: #fff; border: 1px solid #ebeef5; border-radius: 8px; padding: 12px; display: flex; flex-direction: column; gap: 10px; }
.panel-title-line { font-weight: 600; padding: 4px 8px; background: #f5f7fa; border-radius: 6px; color: #333; }
.panel-search { display:flex; gap:8px; }
.category-list { flex:1; min-height:0; overflow: hidden; }
.category-row { display:flex; align-items:center; justify-content:space-between; padding:10px 10px; cursor:pointer; border-radius:6px; margin: 4px 0; }
.category-row:hover { background:#f6f7fb; }
.category-row.active { background:#409eff; color:#fff; }
.category-row .name { white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.category-row .ops .more { display:inline-block; width:18px; text-align:center; font-weight:600; color:#909399; }
.category-footer { text-align:center; padding-top:4px; }
.goods-content { flex:1; min-width:0; background:#fff; border:1px solid #ebeef5; border-radius:8px; padding:16px; display:flex; flex-direction:column }
.toolbar { display:flex; gap:12px; align-items:center; margin-bottom:12px; }
.spacer { flex:1 }
.img-ph { width:48px; height:48px; border-radius:6px; background:#f5f7fa; color:#909399; display:flex; align-items:center; justify-content:center; font-size:12px; }
.pager { display:flex; justify-content:flex-end; padding-top:12px; }

/* 商品表单样式 */
.goods-form .field-inline { display:flex; align-items:center; gap:8px; }
.goods-form .field-inline :deep(.el-input),
.goods-form .field-inline :deep(.el-select),
.goods-form .field-inline :deep(.el-input-number) { flex:1; }

.goods-form .field-inline :deep(.el-select) {width: 100%;}
.goods-form .field-inline :deep(.el-select .el-input__wrapper) {width: 100%;}
.goods-form .field-inline :deep(.el-input-number) { width:100%; }
.required-mark { color:#f56c6c; font-size:18px; line-height:1; }
.optional-hint { color:#909399; font-size:12px; }
</style>
