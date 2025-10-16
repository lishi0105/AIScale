<template>
  <div>
    <h2 style="margin: 8px 0 16px;">中队管理</h2>

    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索名称或编码"
        clearable
        @clear="onSearch"
        @keyup.enter="onSearch"
        style="width: 240px"
      />
      <el-select
        v-model="statusFilter"
        placeholder="全部状态"
        clearable
        @change="onStatusChange"
        style="width: 150px; margin-left: 12px;"
      >
        <el-option
          v-for="opt in statusOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <div class="spacer" />
      <el-button type="primary" :disabled="!isAdminRef" @click="openCreate">
        + 新增中队
      </el-button>
    </div>

    <el-table :data="rows" stripe v-loading="tableLoading" style="width: 100%">
      <el-table-column type="index" label="序号" width="70" />
      <el-table-column prop="Code" label="编码" width="150">
        <template #default="{ row }">
          {{ row.Code || '—' }}
        </template>
      </el-table-column>
      <el-table-column prop="Name" label="名称" min-width="160" />
      <el-table-column prop="Leader" label="负责人" min-width="120" />
      <el-table-column prop="Phone" label="联系电话" min-width="140" />
      <el-table-column label="状态" width="120" align="center">
        <template #default="{ row }">
          <el-tag :type="row.Status === STATUS_ENABLED ? 'success' : 'info'">
            {{ statusLabel(row.Status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="Sort" label="排序码" width="100" />
      <el-table-column prop="Remark" label="备注" min-width="200">
        <template #default="{ row }">
          {{ row.Remark || '—' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" align="center">
        <template #default="{ row }">
          <el-button link @click="openEdit(row)" :disabled="!isAdminRef">编辑</el-button>
          <el-button
            link
            type="danger"
            :disabled="!isAdminRef || deletingId === row.ID"
            @click="onDelete(row)"
          >
            <span v-if="deletingId === row.ID">删除中…</span>
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
        @current-change="onPageChange"
      />
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <el-form :model="form" label-width="100px" v-loading="submitLoading">
        <el-form-item label="名称">
          <el-input v-model="form.Name" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="编码">
          <el-input
            v-model="form.Code"
            maxlength="32"
            placeholder="留空将按后端规则处理"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.Leader" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.Phone" maxlength="32" show-word-limit />
        </el-form-item>
        <el-form-item label="排序码">
          <el-input-number v-model="form.Sort" :min="0" :step="1" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.Status">
            <el-radio :label="STATUS_ENABLED">启用</el-radio>
            <el-radio :label="STATUS_DISABLED">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            type="textarea"
            v-model="form.Remark"
            maxlength="255"
            show-word-limit
            rows="3"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false" :disabled="submitLoading">取消</el-button>
        <el-button type="primary" @click="onSubmit" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { OrganAPI } from '@/api/acl'
import { notifyError } from '@/utils/notify'
import {
  STATUS_DISABLED,
  STATUS_ENABLED,
  statusLabel,
  STATUS_LABELS,
  ROLE_ADMIN,
} from '@/utils/role'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'

interface OrganRow {
  ID: string
  Name: string
  Code?: string | null
  Leader?: string
  Phone?: string
  Sort: number
  Status: number
  Remark?: string | null
}

const rows = ref<OrganRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const keyword = ref('')
const statusFilter = ref<number | null>(null)

const tableLoading = ref(false)
const deletingId = ref<string | null>(null)

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)

type FormState = {
  ID?: string
  Name: string
  Code?: string
  Leader?: string
  Phone?: string
  Sort: number
  Status: number
  Remark?: string
}

const form = ref<FormState>({
  Name: '',
  Code: '',
  Leader: '',
  Phone: '',
  Sort: 0,
  Status: STATUS_ENABLED,
  Remark: '',
})

const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const isAdminRef = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

const statusOptions = computed(() => [
  { value: STATUS_ENABLED, label: STATUS_LABELS[STATUS_ENABLED] },
  { value: STATUS_DISABLED, label: STATUS_LABELS[STATUS_DISABLED] },
])

const dialogTitle = computed(() => (dialogMode.value === 'create' ? '新增中队' : '编辑中队'))

const fetchList = async () => {
  try {
    tableLoading.value = true
    const payload: Record<string, any> = {
      keyword: keyword.value.trim() || undefined,
      limit: pageSize.value,
      offset: (page.value - 1) * pageSize.value,
    }
    if (statusFilter.value !== null && statusFilter.value !== undefined) {
      payload.status = statusFilter.value
    }
    const { data } = await OrganAPI.list(payload)
    rows.value = data?.items || []
    total.value = data?.total || 0
  } catch (err) {
    notifyError(err)
  } finally {
    tableLoading.value = false
  }
}

const onPageChange = (p: number) => {
  page.value = p
  fetchList()
}

const onSearch = () => {
  page.value = 1
  fetchList()
}

const onStatusChange = () => {
  page.value = 1
  fetchList()
}

const openCreate = () => {
  if (!isAdminRef.value) return
  dialogMode.value = 'create'
  form.value = { Name: '', Code: '', Leader: '', Phone: '', Sort: 0, Status: STATUS_ENABLED, Remark: '' }
  dialogVisible.value = true
}

const openEdit = (row: OrganRow) => {
  if (!isAdminRef.value) return
  dialogMode.value = 'edit'
  form.value = {
    ID: row.ID,
    Name: row.Name,
    Code: row.Code ?? '',
    Leader: row.Leader ?? '',
    Phone: row.Phone ?? '',
    Sort: row.Sort ?? 0,
    Status: row.Status,
    Remark: row.Remark ?? '',
  }
  dialogVisible.value = true
}

const onSubmit = async () => {
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可操作')
    return
  }
  const name = form.value.Name?.trim()
  if (!name) {
    ElMessage.warning('请输入名称')
    return
  }
  const sort = Number(form.value.Sort ?? 0)
  if (!Number.isInteger(sort) || sort < 0) {
    ElMessage.warning('排序码需为非负整数')
    return
  }

  const payload: any = {
    name,
    code: form.value.Code?.trim() || undefined,
    leader: form.value.Leader?.trim() || undefined,
    phone: form.value.Phone?.trim() || undefined,
    sort,
    status: form.value.Status,
    remark: form.value.Remark?.trim() || undefined,
  }

  try {
    submitLoading.value = true
    if (dialogMode.value === 'create') {
      await OrganAPI.create(payload)
      ElMessage.success('创建成功')
    } else {
      payload.id = form.value.ID
      await OrganAPI.update(payload)
      ElMessage.success('保存成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (err) {
    notifyError(err)
  } finally {
    submitLoading.value = false
  }
}

const onDelete = async (row: OrganRow) => {
  if (!isAdminRef.value) return
  try {
    await ElMessageBox.confirm(`确认删除【${row.Name}】吗？删除后将无法恢复`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    deletingId.value = row.ID
    await OrganAPI.remove(row.ID)
    ElMessage.success('删除成功')
    fetchList()
  } catch (err) {
    notifyError(err)
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.toolbar .spacer {
  flex: 1;
}

.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
