<template>
  <div class="page-suppliers">
    <aside class="supplier-panel">
      <div class="panel-header">
        <div class="panel-title">
          <h2>供货商列表</h2>
          <p class="panel-sub">点击左侧供货商查看详情</p>
        </div>
        <el-button size="small" type="primary" @click="openCreate" :disabled="!isAdmin">+ 新增供货商</el-button>
      </div>
      <div class="panel-search">
        <el-input
          v-model="keywordInput"
          size="small"
          clearable
          placeholder="搜索拼音/名称/联系人/地址/电话"
          @clear="onSearch"
          @keyup.enter="onSearch"
        />
        <el-button size="small" @click="onSearch">搜索</el-button>
      </div>
      <div class="supplier-list" v-loading="listLoading">
        <el-empty v-if="!listLoading && !cardSuppliers.length" description="暂无供货商" />
        <el-scrollbar v-else>
          <div
            v-for="item in cardSuppliers"
            :key="item.id"
            class="supplier-item"
            :class="{ active: item.id === selectedId }"
            @click="selectSupplier(item.id)"
          >
            <div class="item-info">
              <div class="item-avatar">{{ item.avatar }}</div>
              <div class="item-text">
                <div class="item-name" :title="item.name">{{ item.name }}</div>
                <div class="item-sub">
                  <template v-if="item.fallback">
                    {{ item.fallback }}
                  </template>
                  <template v-else>
                    <span v-if="item.contact" class="item-contact" :title="item.contact">联系人：{{ item.contact }}</span>
                    <span v-if="item.phone" class="item-phone" :title="item.phone">电话：{{ item.phone }}</span>
                    <span v-else-if="item.address" class="item-address" :title="item.address">地址：{{ item.address }}</span>
                  </template>
                </div>
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

    <section class="supplier-content">
      <div v-if="selectedSupplier" class="detail-card">
        <h3 class="detail-title">供货商详情</h3>
        <el-descriptions :column="2" border size="small" class="detail-grid">
          <el-descriptions-item label="供货商名称">{{ selectedSupplier.Name }}</el-descriptions-item>
          <el-descriptions-item label="供货商编码">
            {{ selectedSupplier.Code || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="拼音">
            {{ selectedSupplier.Pinyin || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="浮动比例">{{ formatRatio(selectedSupplier.FloatRatio) }}</el-descriptions-item>
          <el-descriptions-item label="联系人">
            {{ selectedSupplier.ContactName || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="联系电话">
            {{ selectedSupplier.ContactPhone || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="联系邮箱">
            {{ selectedSupplier.ContactEmail || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="联系地址">
            {{ selectedSupplier.ContactAddress || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">{{ statusLabel(selectedSupplier.Status) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(selectedSupplier.CreatedAt) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(selectedSupplier.UpdatedAt) }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions :column="1" border size="small" class="detail-description">
          <el-descriptions-item label="描述信息">
            {{ formatDescription(selectedSupplier.Description) }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <el-empty v-else description="请选择左侧供货商" />
    </section>

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
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled } from '@element-plus/icons-vue'
import { SupplierAPI } from '@/api/supplier'
import type { SupplierCreatePayload, SupplierRow, SupplierUpdatePayload } from '@/api/supplier'
import { notifyError } from '@/utils/notify'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'

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

type DropdownCommand = 'edit' | 'delete'

const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})

const organId = computed(() => jwtPayload.value?.org_id || '')
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

const normalizedKeyword = computed(() => filterKeyword.value.trim().toLowerCase())

const toLower = (value: string | null | undefined) => (value ? value.toLowerCase() : '')

const displayedSuppliers = computed(() => {
  const kw = normalizedKeyword.value
  if (!kw) return suppliers.value
  return suppliers.value.filter(item => {
    const name = toLower(item.Name)
    const pinyin = toLower(item.Pinyin)
    const contact = toLower(item.ContactName)
    const phone = toLower(item.ContactPhone)
    const address = toLower(item.ContactAddress)
    return [name, pinyin, contact, phone, address].some(field => field.includes(kw))
  })
})

const cardSuppliers = computed(() =>
  displayedSuppliers.value.map(item => {
    const contact = (item.ContactName || '').trim()
    const phone = (item.ContactPhone || '').trim()
    const address = (item.ContactAddress || '').trim()
    const fallback = contact || phone || address ? '' : '暂无联系信息'
    return {
      id: item.ID,
      name: item.Name,
      avatar: item.Name?.charAt(0)?.toUpperCase() || '#',
      contact,
      phone,
      address,
      fallback,
      raw: item,
    }
  })
)

const selectedSupplier = computed(() =>
  suppliers.value.find(item => item.ID === selectedId.value) || null
)

const statusLabel = (status: number) => (status === 1 ? '正常' : '禁用')

const formatRatio = (ratio: number) => {
  if (ratio === undefined || ratio === null) return '—'
  return Number(ratio).toFixed(4)
}

const formatDate = (value: string | null) => {
  if (!value) return '—'
  try {
    return new Date(value).toLocaleString()
  } catch (error) {
    return value
  }
}

const formatDescription = (desc: string) => {
  const trimmed = desc ? desc.trim() : ''
  return trimmed || '—'
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

watch(
  () => organId.value,
  () => {
    fetchSuppliers()
  },
  { immediate: true }
)

watch(
  () => displayedSuppliers.value,
  (list) => {
    ensureSelection(list)
  },
  { immediate: true }
)

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

const fetchSuppliers = async () => {
  if (!organId.value) {
    suppliers.value = []
    selectedId.value = ''
    return
  }
  listLoading.value = true
  try {
    const { data } = await SupplierAPI.list({
      org_id: organId.value,
      page: 1,
      page_size: 500,
    })
    suppliers.value = data?.items || []
  } catch (error) {
    notifyError(error)
  } finally {
    listLoading.value = false
  }
}

const selectSupplier = (id: string) => {
  selectedId.value = id
}

const onCardCommand = (command: DropdownCommand, row: SupplierRow) => {
  selectedId.value = row.ID
  if (command === 'edit') {
    openEdit(row)
  } else if (command === 'delete') {
    confirmDelete(row)
  }
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

const onSearch = () => {
  filterKeyword.value = keywordInput.value.trim()
  fetchSuppliers()
}

const optionalString = (value: string) => {
  const trimmed = value.trim()
  return trimmed || undefined
}

const buildUpdateString = (value: string, original: string | null | undefined) => {
  const trimmed = value.trim()
  const originalTrimmed = original ? original.trim() : ''
  if (trimmed === originalTrimmed) {
    return undefined
  }
  return trimmed || null
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
  if (form.floatRatio === null || form.floatRatio === undefined || Number(form.floatRatio) <= 0) {
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
        description: form.description.trim() || ' ',
      }
      const contactName = optionalString(form.contactName)
      if (contactName !== undefined) payload.contact_name = contactName
      const contactPhone = optionalString(form.contactPhone)
      if (contactPhone !== undefined) payload.contact_phone = contactPhone
      const contactEmail = optionalString(form.contactEmail)
      if (contactEmail !== undefined) payload.contact_email = contactEmail
      const contactAddress = optionalString(form.contactAddress)
      if (contactAddress !== undefined) payload.contact_address = contactAddress
      const { data } = await SupplierAPI.create(payload)
      ElMessage.success('新增供货商成功')
      await fetchSuppliers()
      if (data?.ID) {
        selectedId.value = data.ID
      }
    } else if (editingSupplier.value) {
      const payload: SupplierUpdatePayload = { id: form.id }

      if (name !== editingSupplier.value.Name) {
        payload.name = name
      }

      const ratio = Number(form.floatRatio)
      if (!Number.isNaN(ratio) && ratio > 0 && ratio !== editingSupplier.value.FloatRatio) {
        payload.float_ratio = ratio
      }

      const contactName = buildUpdateString(form.contactName, editingSupplier.value.ContactName)
      if (contactName !== undefined) {
        payload.contact_name = contactName
      }
      const contactPhone = buildUpdateString(form.contactPhone, editingSupplier.value.ContactPhone)
      if (contactPhone !== undefined) {
        payload.contact_phone = contactPhone
      }
      const contactEmail = buildUpdateString(form.contactEmail, editingSupplier.value.ContactEmail)
      if (contactEmail !== undefined) {
        payload.contact_email = contactEmail
      }
      const contactAddress = buildUpdateString(form.contactAddress, editingSupplier.value.ContactAddress)
      if (contactAddress !== undefined) {
        payload.contact_address = contactAddress
      }

      const description = form.description.trim()
      if (description !== (editingSupplier.value.Description || '').trim()) {
        payload.description = description
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
  } catch (error) {
    notifyError(error)
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
        type: 'warning',
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
  } catch (error) {
    notifyError(error)
  } finally {
    deletingId.value = ''
  }
}
</script>

<style scoped>
.page-suppliers {
  display: flex;
  gap: 16px;
  height: calc(100vh - 120px);
  min-height: 520px;
}

.supplier-panel {
  width: 320px;
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

.supplier-list {
  flex: 1;
  min-height: 0;
}

.supplier-item {
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

.supplier-item:last-child {
  margin-bottom: 0;
}

.supplier-item:hover,
.supplier-item.active {
  border-color: #67c23a;
  background: rgba(103, 194, 58, 0.12);
  box-shadow: 0 4px 12px rgba(103, 194, 58, 0.18);
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
  background: linear-gradient(135deg, #67c23a, #95d475);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  box-shadow: 0 4px 10px rgba(103, 194, 58, 0.35);
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
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-contact,
.item-phone,
.item-address {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-phone {
  color: #67c23a;
}

.item-meta {
  display: flex;
  align-items: center;
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
  background: rgba(103, 194, 58, 0.15);
  color: #67c23a;
}

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

.detail-title {
  margin: 0 0 16px;
  font-size: 18px;
  font-weight: 600;
}

.detail-grid .el-descriptions__label {
  width: 120px;
}

.detail-description {
  margin-top: 8px;
}
</style>
