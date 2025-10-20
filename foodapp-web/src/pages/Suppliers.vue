<!-- src/pages/Suppliers.vue -->
<template>
  <div class="page-suppliers">
    <!-- 左侧：表格 + 工具栏 -->
    <aside class="supplier-panel">
      <div class="list-header">
        <h2 class="list-title">供货商管理</h2>
        <div class="list-tools">
          <el-input
            v-model="keywordInput"
            size="small"
            clearable
            placeholder="请输入关键字"
            @clear="onSearch"
            @keyup.enter="onSearch"
            style="width: 220px"
          />
          <el-button size="small" type="primary" @click="onSearch">查询</el-button>
          <el-button size="small" type="primary" @click="openCreate" :disabled="!isAdmin">
            + 新增供货商
          </el-button>
        </div>
      </div>

      <el-table
        class="supplier-table"
        :data="sortedSuppliers"
        stripe
        v-loading="listLoading"
        highlight-current-row
        :row-key="getRowKey"
        @current-change="onRowChange"
        :style="{ '--el-font-size-base':'clamp(13.5px, 1.0vw, 16px)' }"
      >
        <el-table-column type="index" label="序号" width="70" />

        <el-table-column label="供应商名称" min-width="200">
          <template #header>
            <span class="th-clickable" @click="toggleSort('pinyin')">
              供应商名称
              <i class="caret" :class="caretClass('pinyin')"></i>
            </span>
          </template>
          <template #default="{ row }">
            <el-link type="primary" @click.stop="selectSupplier(row.ID)">
              {{ row.Name }}
            </el-link>
          </template>
        </el-table-column>

        <el-table-column prop="ContactName" label="联系人" min-width="140">
          <template #default="{ row }">
            {{ row.ContactName || '—' }}
          </template>
        </el-table-column>

        <el-table-column prop="ContactPhone" label="联系电话" min-width="160">
          <template #default="{ row }">
            {{ row.ContactPhone || '—' }}
          </template>
        </el-table-column>

        <el-table-column label="浮动比例" width="140" align="center">
          <template #header>
            <span class="th-clickable" @click="toggleSort('ratio')">
              浮动比例
              <i class="caret" :class="caretClass('ratio')"></i>
            </span>
          </template>
          <template #default="{ row }">
            <el-link type="success" :title="formatRatio(row.FloatRatio)">
              {{ formatRatioPct(row.FloatRatio) }}
            </el-link>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="160" align="center">
          <template #default="{ row }">
            <el-button link @click.stop="openEdit(row)" :disabled="!isAdmin">编辑</el-button>
            <el-button
              link
              type="danger"
              @click.stop="confirmDelete(row)"
              :disabled="!isAdmin || deletingId===row.ID"
            >
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
          @current-change="onPageChange"
          @size-change="onPageSizeChange"
        />
      </div>
    </aside>

    <!-- 右侧：单列详情 -->
    <section class="supplier-content" :style="fontVars">
      <div v-if="selectedSupplier" class="detail-card">
        <!-- 顶部摘要 -->
        <div class="detail-header">
          <div class="detail-header-left">
            <div class="detail-name">{{ selectedSupplier.Name }}</div>
            <div class="detail-meta">
              <el-tag size="small" type="info">编码：{{ selectedSupplier.Code || '—' }}</el-tag>
              <el-tag size="small" :type="selectedSupplier.Status === 1 ? 'success' : 'warning'">
                {{ statusLabel(selectedSupplier.Status) }}
              </el-tag>
            </div>
          </div>
          <div class="ratio-hero" :title="'原始值：' + formatRatio(selectedSupplier.FloatRatio)">
            <div class="ratio-hero-value">{{ formatRatioPct(selectedSupplier.FloatRatio) }}</div>
            <div class="ratio-hero-label">浮动比例</div>
          </div>
        </div>

        <!-- 单列详情 -->
        <el-descriptions :column="1" border size="small" class="detail-grid">
          <el-descriptions-item label="拼音">{{ selectedSupplier.Pinyin || '—' }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ selectedSupplier.ContactName || '—' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ selectedSupplier.ContactPhone || '—' }}</el-descriptions-item>
          <el-descriptions-item label="联系邮箱">{{ selectedSupplier.ContactEmail || '—' }}</el-descriptions-item>
          <el-descriptions-item label="联系地址">{{ selectedSupplier.ContactAddress || '—' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(selectedSupplier.CreatedAt) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(selectedSupplier.UpdatedAt) }}</el-descriptions-item>
          <el-descriptions-item label="描述信息">{{ formatDescription(selectedSupplier.Description) }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <el-empty v-else description="请选择左侧供货商" />
    </section>

    <!-- 弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增供货商' : '编辑供货商'"
      width="620px"
      @closed="onDialogClosed"
    >
      <el-form label-width="110px" :model="form" v-loading="submitLoading">
        <el-form-item label="供货商名称">
          <el-input v-model="form.name" maxlength="128" placeholder="请输入供货商名称" />
        </el-form-item>
        <el-form-item label="浮动比例">
          <el-input-number
            v-model="form.floatRatio"
            :min="0.0001"
            :step="0.01"
            :precision="4"
            controls-position="right"
            placeholder="请输入浮动比例"
          />
        </el-form-item>
        <el-form-item label="联系人姓名">
          <el-input v-model="form.contactName" maxlength="64" placeholder="可选" clearable />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.contactPhone" maxlength="32" placeholder="可选" clearable />
        </el-form-item>
        <el-form-item label="联系邮箱">
          <el-input v-model="form.contactEmail" maxlength="128" placeholder="可选" clearable />
        </el-form-item>
        <el-form-item label="联系地址">
          <el-input v-model="form.contactAddress" maxlength="255" placeholder="可选" clearable />
        </el-form-item>
        <el-form-item label="描述信息">
          <el-input
            v-model="form.description"
            type="textarea"
            maxlength="500"
            show-word-limit
            placeholder="可选"
            :rows="3"
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { SupplierAPI } from '@/api/supplier'
import type { SupplierCreatePayload, SupplierRow, SupplierUpdatePayload } from '@/api/supplier'
import { notifyError } from '@/utils/notify'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'

const page = ref(1)
const pageSize = ref(15)
const pageSizes = [10, 15, 20, 50]
const total = ref(0)
const keywordInput = ref('')
const filterKeyword = ref('')
const suppliers = ref<SupplierRow[]>([])
const listLoading = ref(false)
const selectedId = ref('')
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)
const deletingId = ref('')
const editingSupplier = ref<SupplierRow | null>(null)

interface SupplierForm {
  id: string
  name: string
  floatRatio: number | null
  contactName: string
  contactPhone: string
  contactEmail: string
  contactAddress: string
  description: string
}
const form = reactive<SupplierForm>({
  id: '',
  name: '',
  floatRatio: null,
  contactName: '',
  contactPhone: '',
  contactEmail: '',
  contactAddress: '',
  description: '',
})

const onPageChange = (p: number) => {
  page.value = p
  fetchSuppliers()
}

const onPageSizeChange = (ps: number) => {
  pageSize.value = ps
  page.value = 1
  fetchSuppliers()
}

const getRowKey = (row: SupplierRow) => row.ID
const onRowChange = (row: SupplierRow | undefined) => {
  if (row) selectSupplier(row.ID)
}

// 登录信息
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const organId = computed(() => jwtPayload.value?.org_id || '')
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

// 关键字本地过滤
const normalizedKeyword = computed(() => filterKeyword.value.trim().toLowerCase())
const toLower = (v: string | null | undefined) => (v ? v.toLowerCase() : '')

const displayedSuppliers = computed(() => {
  const kw = normalizedKeyword.value
  if (!kw) return suppliers.value
  return suppliers.value.filter(item => {
    const name = toLower(item.Name)
    const pinyin = toLower(item.Pinyin)
    const contact = toLower(item.ContactName)
    const phone = toLower(item.ContactPhone)
    const address = toLower(item.ContactAddress)
    return [name, pinyin, contact, phone, address].some(f => f.includes(kw))
  })
})

// 排序：pinyin / ratio
const sortKey = ref<'pinyin' | 'ratio' | null>('pinyin')
const sortOrder = ref<'asc' | 'desc'>('asc')

const toggleSort = (key: 'pinyin' | 'ratio') => {
  if (sortKey.value !== key) {
    sortKey.value = key
    sortOrder.value = 'asc'
  } else {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  }
}
const caretClass = (key: 'pinyin' | 'ratio') => {
  if (sortKey.value !== key) return ''
  return sortOrder.value === 'asc' ? 'asc' : 'desc'
}

const sortedSuppliers = computed(() => {
  const list = [...displayedSuppliers.value]
  if (!sortKey.value) return list
  const asc = sortOrder.value === 'asc' ? 1 : -1
  if (sortKey.value === 'pinyin') {
    return list.sort(
      (a, b) =>
        (toLower(a.Pinyin || a.Name) > toLower(b.Pinyin || b.Name) ? 1 : -1) * asc
    )
  }
  return list.sort(
    (a, b) => ((Number(a.FloatRatio) || 0) - (Number(b.FloatRatio) || 0)) * asc
  )
})

// 右侧详情选中
const selectedSupplier = computed(
  () => suppliers.value.find(i => i.ID === selectedId.value) || null
)

const statusLabel = (status: number) => (status === 1 ? '正常' : '禁用')

const formatRatio = (ratio: number) => {
  if (ratio === undefined || ratio === null) return '—'
  return Number(ratio).toFixed(4)
}
const formatRatioPct = (ratio: number | null | undefined) => {
  if (ratio === null || ratio === undefined) return '—'
  const n = Number(ratio)
  if (Number.isNaN(n)) return '—'
  return (n * 100).toFixed(2) + '%'
}
const formatDate = (value: string | null) => {
  if (!value) return '—'
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}
const formatDescription = (desc: string) => {
  const t = desc ? desc.trim() : ''
  return t || '—'
}

const ensureSelection = (list: SupplierRow[] | undefined) => {
  const first = list?.[0]
  if (!first) {
    selectedId.value = ''
    return
  }
  if (!list!.some(item => item.ID === selectedId.value)) {
    selectedId.value = first.ID
  }
}

// org_id 变化 & 首次加载
watch(
  () => organId.value,
  val => {
    if (!val) {
      ElMessage.error('缺少中队信息（org_id），无法加载供货商列表')
      suppliers.value = []
      selectedId.value = ''
      return
    }
    page.value = 1  
    fetchSuppliers()
  },
  { immediate: false }
)

onMounted(() => {
  if (!organId.value) {
    ElMessage.error('缺少中队信息（org_id），无法加载供货商列表')
    suppliers.value = []
    selectedId.value = ''
    return
  }
  fetchSuppliers()
})

// 根据排序后的列表保证选中项
watch(
  () => sortedSuppliers.value,
  list => ensureSelection(list),
  { immediate: true }
)

const fetchSuppliers = async () => {
  if (!organId.value) {
    suppliers.value = []
    selectedId.value = ''
    total.value = 0
    return
  }
  listLoading.value = true
  try {
    const { data } = await SupplierAPI.list({
      org_id: organId.value,
      page: page.value,          // ✅ 当前页
      page_size: pageSize.value, // ✅ 每页大小
    })
    suppliers.value = Array.isArray(data?.items) ? data.items : []
    total.value = Number(data?.total || 0)       // ✅ 总数
  } catch (error) {
    notifyError(error)
  } finally {
    listLoading.value = false
  }
}

const selectSupplier = (id: string) => {
  selectedId.value = id
}

const onSearch = () => {
  filterKeyword.value = keywordInput.value.trim()
  if (!organId.value) {
    ElMessage.error('缺少中队信息（org_id），无法搜索供货商')
    return
  }
   page.value = 1
   fetchSuppliers()
}

const openCreate = () => {
  dialogMode.value = 'create'
  resetForm()
  editingSupplier.value = null
  dialogVisible.value = true
}
const openEdit = (row: SupplierRow) => {
  dialogMode.value = 'edit'
  editingSupplier.value = row
  form.id = row.ID
  form.name = row.Name
  form.floatRatio = row.FloatRatio
  form.contactName = row.ContactName || ''
  form.contactPhone = row.ContactPhone || ''
  form.contactEmail = row.ContactEmail || ''
  form.contactAddress = row.ContactAddress || ''
  form.description = row.Description || ''
  dialogVisible.value = true
}

const optionalString = (v: string) => {
  const t = v.trim()
  return t || undefined
}
const buildUpdateString = (v: string, o: string | null | undefined) => {
  const t = v.trim()
  const ot = o ? o.trim() : ''
  if (t === ot) return undefined
  return t || null
}

const onSubmit = async () => {
  const name = form.name.trim()
  if (!name) {
    ElMessage.warning('请输入供货商名称')
    return
  }
  if (!organId.value) {
    ElMessage.warning('缺少中队信息，无法提交')
    return
  }
  if (
    form.floatRatio === null ||
    form.floatRatio === undefined ||
    Number(form.floatRatio) <= 0
  ) {
    ElMessage.warning('请输入正确的浮动比例')
    return
  }

  submitLoading.value = true
  try {
    if (dialogMode.value === 'create') {
      const payload: SupplierCreatePayload = {
        name,
        org_id: organId.value,
        float_ratio: Number(form.floatRatio),
        description: form.description.trim() || ' '
      }
      const cn = optionalString(form.contactName)
      if (cn !== undefined) payload.contact_name = cn
      const cp = optionalString(form.contactPhone)
      if (cp !== undefined) payload.contact_phone = cp
      const ce = optionalString(form.contactEmail)
      if (ce !== undefined) payload.contact_email = ce
      const ca = optionalString(form.contactAddress)
      if (ca !== undefined) payload.contact_address = ca
      const { data } = await SupplierAPI.create(payload)
      ElMessage.success('新增供货商成功')
      await fetchSuppliers()
      if (data?.ID) {
        selectedId.value = data.ID
      }
    } else if (editingSupplier.value) {
      const payload: SupplierUpdatePayload = { id: form.id }
      if (name !== editingSupplier.value.Name) payload.name = name
      const ratio = Number(form.floatRatio)
      if (!Number.isNaN(ratio) && ratio > 0 && ratio !== editingSupplier.value.FloatRatio) {
        payload.float_ratio = ratio
      }
      const cn = buildUpdateString(form.contactName, editingSupplier.value.ContactName)
      if (cn !== undefined) payload.contact_name = cn
      const cp = buildUpdateString(form.contactPhone, editingSupplier.value.ContactPhone)
      if (cp !== undefined) payload.contact_phone = cp
      const ce = buildUpdateString(form.contactEmail, editingSupplier.value.ContactEmail)
      if (ce !== undefined) payload.contact_email = ce
      const ca = buildUpdateString(form.contactAddress, editingSupplier.value.ContactAddress)
      if (ca !== undefined) payload.contact_address = ca
      const desc = form.description.trim()
      if (desc !== (editingSupplier.value.Description || '').trim()) {
        payload.description = desc
      }
      if (Object.keys(payload).length === 1) {
        ElMessage.info('未检测到需要保存的修改')
        return
      }
      await SupplierAPI.update(payload)
      ElMessage.success('更新供货商成功')
      await fetchSuppliers()
      selectedId.value = form.id
    }
    dialogVisible.value = false
  } catch (e) {
    notifyError(e)
  } finally {
    submitLoading.value = false
  }
}

const confirmDelete = async (row: SupplierRow) => {
  if (deletingId.value) return
  try {
    await ElMessageBox.confirm(
      `确认删除供货商“${row.Name}”吗？`,
      '删除确认',
      {
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    return
  }

  deletingId.value = row.ID
  try {
    await SupplierAPI.remove(row.ID)
    ElMessage.success('删除成功')
    if (selectedId.value === row.ID) {
      selectedId.value = ''
    }
    await fetchSuppliers()
  } catch (e) {
    notifyError(e)
  } finally {
    deletingId.value = ''
  }
}

const resetForm = () => {
  form.id = ''
  form.name = ''
  form.floatRatio = null
  form.contactName = ''
  form.contactPhone = ''
  form.contactEmail = ''
  form.contactAddress = ''
  form.description = ''
}
const onDialogClosed = () => {
  resetForm()
  editingSupplier.value = null
}

// 右侧字体自适应（EP 变量）
const fontVars = computed(() => ({
  '--el-font-size-base': 'clamp(14px, 1.15vw, 17px)',
  '--el-font-size-small': 'clamp(13px, 1.0vw, 16px)',
  '--el-font-size-extra-small': 'clamp(12px, 0.9vw, 14px)'
}))
</script>

<style scoped>
.page-suppliers {
  display: flex;
  gap: 16px;
  height: calc(100vh - 120px);
  min-height: 520px;
}

.pager-suppliers { display:flex; justify-content:flex-end; padding-top:12px; }

/* 左侧区域 */
.supplier-panel {
  width: 60%;
  min-width: 620px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.list-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.list-tools {
  display: flex;
  gap: 8px;
  align-items: center;
}

.supplier-table :deep(.el-link) {
  cursor: pointer;
}
.pager {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
}
/* 表头点击效果与箭头 */
.th-clickable {
  cursor: pointer;
  user-select: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.caret {
  border: 5px solid transparent;
  margin-left: 2px;
}
.caret.asc {
  border-bottom-color: #303133;
  transform: translateY(-2px);
}
.caret.desc {
  border-top-color: #303133;
  transform: translateY(2px);
}

/* 右侧详情 */
.supplier-content {
  flex: 1;
  min-width: 0;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.detail-card {
  width: 100%;
}
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #fafcff 100%);
}
.detail-header-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}
.detail-name {
  font-size: clamp(18px, 1.4vw, 22px);
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.detail-meta {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 英雄比例徽章 */
.ratio-hero {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  min-width: 120px;
  user-select: none;
}
.ratio-hero-value {
  font-size: clamp(22px, 2vw, 28px);
  font-weight: 800;
  line-height: 1;
  color: #67c23a;
}
.ratio-hero-label {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}

/* 单列描述里标签列更清晰 */
.detail-grid :deep(.el-descriptions__label) {
  width: 120px;
  font-weight: 600;
}
</style>
