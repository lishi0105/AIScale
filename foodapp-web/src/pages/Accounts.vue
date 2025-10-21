<template>
  <div>
    <h2 style="margin:8px 0 16px;">账户管理</h2>

    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索用户名"
        clearable
        @clear="onSearch"
        @keyup.enter="onSearch"
        style="width:240px"
      />
      <div class="spacer" />
      <el-button type="primary" @click="openCreate" :disabled="!isAdminRef">
        + 新增账户
      </el-button>
    </div>

    <el-table :data="rows" stripe v-loading="tableLoading">
      <el-table-column type="index" label="序号" width="70" :index="indexMethod" />
      <el-table-column prop="Username" label="用户名" min-width="160" />
      <el-table-column label="所属机构" min-width="180">
        <template #default="{ row }">{{ orgName(row.OrgID) }}</template>
      </el-table-column>
      <el-table-column label="角色" width="140" align="center">
        <template #default="{ row }">
          <el-tag>{{ roleLabel(row.Role) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="描述" min-width="220">
        <template #default="{ row }">{{ row.Description || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="360" align="center">
        <template #default="{ row }">
          <el-button link @click="openEdit(row)">编辑</el-button>
          <el-button
            link
            type="warning"
            :disabled="!canResetPassword(row)"
            @click="openResetGuard(row)"
          >
            {{ isSelf(row) ? '修改密码' : '重置密码' }}
          </el-button>
          <el-button
            link
            type="danger"
            :disabled="!isUser(row) || deletingId === row.ID"
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
        layout="sizes, prev, pager, next, jumper, ->, total"
        :page-sizes="pageSizes"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>

    <!-- 新增/编辑 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode==='create'?'新增账户':'编辑账户'"
      width="480px"
    >
      <el-form :model="form" label-width="100px" v-loading="submitLoading">
        <el-form-item label="用户名">
          <el-input v-model="form.Username" />
        </el-form-item>

        <template v-if="dialogMode === 'create'">
          <el-form-item label="设置密码">
            <el-input
              v-model="form.password"
              type="password"
              show-password
              placeholder="至少8位"
            />
          </el-form-item>
          <el-form-item label="确认密码">
            <el-input
              v-model="form.confirmPassword"
              type="password"
              show-password
              placeholder="请再次输入密码"
            />
          </el-form-item>
        </template>

        <el-form-item label="所属机构">
          <el-select
            v-model="form.OrgID"
            filterable
            placeholder="请选择机构"
            style="width: 220px"
          >
            <el-option
              v-for="opt in orgOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="角色">
          <el-select v-model="form.Role" style="width: 220px">
            <el-option
              v-for="opt in roleOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="form.Description"
            type="textarea"
            :rows="3"
            placeholder="可选"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false" :disabled="submitLoading">取消</el-button>
        <el-button type="primary" @click="onSubmit" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 重置/修改密码 -->
    <el-dialog v-model="resetVisible" :title="resetDialogTitle" width="460px">
      <el-form :model="resetForm" label-width="110px" v-loading="resetLoading">
        <el-form-item v-if="needOldPwd" label="旧密码">
          <el-input v-model="resetForm.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="设置密码">
          <el-input v-model="resetForm.new_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="resetForm.confirm_new_password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetVisible=false" :disabled="resetLoading">取消</el-button>
        <el-button type="primary" @click="onDoReset" :loading="resetLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { AccountAPI } from '@/api/acl'
import type { AccountListParams, AccountUpdatePayload } from '@/api/acl'
import { notifyError } from '@/utils/notify'
import {
  ROLE_ADMIN, ROLE_USER, ROLE_LABELS, roleLabel,
} from '@/utils/role'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { getToken } from '@/api/http'
import { OrganAPI } from '@/api/organ'

interface Row {
  ID: string
  Username: string
  OrgID: string
  Description?: string | null
  Role: number
  LastLoginAt?: string | null
  CreatedAt?: string
  UpdatedAt?: string
}
const indexMethod = (rowIndex: number) =>
  (page.value - 1) * pageSize.value + rowIndex + 1

const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})

const currentUser = computed(() => {
  if (!jwtPayload.value) return null
  return {
    ID: jwtPayload.value.sub,
    Username: jwtPayload.value.usr,
    Role: jwtPayload.value.role,
  }
})

const isAdminRef = computed(() => currentUser.value?.Role === ROLE_ADMIN)

const roleOptions = [
  { value: ROLE_ADMIN, label: ROLE_LABELS[ROLE_ADMIN] },
  { value: ROLE_USER, label: ROLE_LABELS[ROLE_USER] },
]
const orgOptions = ref<Array<{ value: string; label: string }>>([])
const orgNameMap = ref<Record<string, string>>({})

const rows = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const pageSizes = [10, 15, 20, 50]
const keyword = ref('')
const tableLoading = ref(false)

const fetchList = async () => {
  try {
    tableLoading.value = true
    const params: AccountListParams = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (keyword.value.trim()) {
      params.keyword = keyword.value.trim()
    }
    const { data } = await AccountAPI.list(params)
    rows.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    notifyError(e)
  } finally {
    tableLoading.value = false
  }
}
const onSearch = () => { page.value = 1; fetchList() }

const handlePageChange = (p: number) => {
  page.value = p
  fetchList()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  page.value = 1
  fetchList()
}

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)

const form = ref<Partial<Row> & { password?: string; confirmPassword?: string }>({
  Role: ROLE_USER,
})

const openCreate = () => {
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可新增账户')
    return
  }
  dialogMode.value = 'create'
  form.value = {
    Username: '',
    OrgID: '',
    Description: '',
    password: '',
    confirmPassword: '',
    Role: ROLE_USER,
  }
  dialogVisible.value = true
}

const openEdit = (row: Row) => {
  dialogMode.value = 'edit'
  form.value = {
    ID: row.ID,
    Username: row.Username,
    OrgID: row.OrgID,
    Description: row.Description || '',
    Role: row.Role,
  }
  dialogVisible.value = true
}

const onSubmit = async () => {
  try {
    submitLoading.value = true
    if (dialogMode.value === 'create') {
      const { Username, password, confirmPassword } = form.value
      if (!Username?.trim()) return ElMessage.warning('请输入用户名')
      if (!password) return ElMessage.warning('请输入密码')
      if (!confirmPassword) return ElMessage.warning('请输入确认密码')
      if (password.length < 8) return ElMessage.error('密码至少8位')
      if (password !== confirmPassword) return ElMessage.error('两次输入的密码不一致')
      if (!form.value.OrgID) return ElMessage.error('请选择所属机构')

      await AccountAPI.create({
        username: Username.trim(),
        password,
        org_id: form.value.OrgID!,
        role: Number(form.value.Role ?? ROLE_USER),
        description: (form.value.Description || '').trim() || undefined,
      })
      ElMessage.success('创建成功')
    } else {
      const payload: AccountUpdatePayload = {
        id: form.value.ID!,
        username: form.value.Username?.trim(),
        org_id: form.value.OrgID,
        description: (form.value.Description || '').trim() || null,
        role: Number(form.value.Role ?? ROLE_USER),
      }
      await AccountAPI.update(payload)
      ElMessage.success('保存成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    notifyError(e)
  } finally {
    submitLoading.value = false
  }
}

const resetVisible = ref(false)
const resetLoading = ref(false)
const resetForm = ref({
  id: '',
  username: '',
  old_password: '',
  new_password: '',
  confirm_new_password: '',
})

const canResetPassword = (row: Row) =>
  isAdminRef.value || currentUser.value?.ID === row.ID

const openResetGuard = (row: Row) => {
  if (!canResetPassword(row)) return ElMessage.warning('无权限修改他人密码')
  openReset(row)
}

const openReset = (row: Row) => {
  resetForm.value = {
    id: row.ID,
    username: row.Username,
    old_password: '',
    new_password: '',
    confirm_new_password: '',
  }
  resetVisible.value = true
}

const needOldPwd = computed(() =>
  currentUser.value?.Username === resetForm.value.username
)
const resetDialogTitle = computed(() =>
  needOldPwd.value ? '修改密码' : '重置密码'
)
const isSelf = (row: Row) => currentUser.value?.ID === row.ID
const isUser = (row: Row) => row.Role === ROLE_USER

const onDoReset = async () => {
  try {
    resetLoading.value = true
    const newPwd = resetForm.value.new_password
    const confirm = resetForm.value.confirm_new_password
    if (newPwd.length < 8) return ElMessage.error('新密码至少8位')
    if (newPwd !== confirm) return ElMessage.error('两次输入的新密码不一致')

    if (needOldPwd.value) {
      if (!resetForm.value.old_password)
        return ElMessage.warning('请输入旧密码')
      await AccountAPI.change_password({
        id: resetForm.value.id,
        old_password: resetForm.value.old_password,
        new_password: newPwd,
      })
    } else {
      if (!isAdminRef.value) return ElMessage.error('无权限修改')
      await AccountAPI.update_password({ id: resetForm.value.id, password: newPwd })
    }
    resetVisible.value = false
    ElMessage.success('密码已更新')
  } catch (e) {
    notifyError(e)
  } finally {
    resetLoading.value = false
  }
}

const deletingId = ref<string | null>(null)
const onDelete = async (row: Row) => {
  if (!isAdminRef.value) return ElMessage.warning('仅管理员可删除账户')
  try {
    await ElMessageBox.confirm(`确认删除账户 “${row.Username}”？`, '提示', { type: 'warning' })
    deletingId.value = row.ID
    await AccountAPI.remove(row.ID)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e:any) {
    if (e?.message) notifyError(e)
  } finally {
    deletingId.value = null
  }
}

const orgName = (id: string) => orgNameMap.value[id] || id || '-'

const loadOrgs = async () => {
  try {
    const { data } = await OrganAPI.list({ page: 1, page_size: 500, is_deleted: 0 })
    const opts = (data.items || []).map((it: any) => ({
      value: it.ID, label: it.Name,
    }))
    orgOptions.value = opts
    const map: Record<string, string> = {}
    for (const o of opts) map[o.value] = o.label
    orgNameMap.value = map
  } catch (e) {}
}

onMounted(() => {
  fetchList()
  loadOrgs()
})
</script>

<style scoped>
.toolbar { display:flex; gap:12px; align-items:center; margin-bottom:12px; }
.spacer { flex:1; }
.pager { display:flex; justify-content:flex-end; padding-top:12px; }
</style>
