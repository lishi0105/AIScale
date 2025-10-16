<!-- src/pages/Accounts.vue -->
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
      <!-- 仅管理员能新增：按钮禁用 -->
      <el-button type="primary" @click="openCreate" :disabled="!isAdminRef">+ 新增账户</el-button>
    </div>

    <el-table :data="rows" stripe v-loading="tableLoading">
      <el-table-column type="index" label="序号" width="80" />
      <el-table-column prop="Username" label="用户名" min-width="160" />
      <el-table-column prop="OrgName" label="所属中队" min-width="160" />
      <el-table-column label="角色" width="140" align="center">
        <template #default="{ row }">
          <el-tag>{{ roleLabel(row.Role) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="140" align="center">
        <template #default="{ row }">
          <el-tag :type="row.IsDeleted === 0 ? 'success' : 'info'">
            {{ row.IsDeleted === 0 ? '正常' : '已删除' }}
          </el-tag>
        </template>
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
        layout="prev, pager, next, jumper, ->, total"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="(p:number)=>{page=p;fetchList()}"
      />
    </div>

    <!-- 新增/编辑 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode==='create'?'新增账户':'编辑账户'" width="480px">
      <el-form :model="form" label-width="100px" v-loading="submitLoading">
        <el-form-item label="用户名">
          <el-input v-model="form.Username" :disabled="dialogMode==='edit'" />
        </el-form-item>

        <!-- 仅创建时显示密码和确认密码 -->
        <template v-if="dialogMode === 'create'">
          <el-form-item label="设置密码">
            <el-input v-model="form.password" type="password" show-password placeholder="至少8位" />
          </el-form-item>
          <el-form-item label="确认密码">
            <el-input v-model="form.confirmPassword" type="password" show-password placeholder="请再次输入密码" />
          </el-form-item>
        </template>

        <el-form-item label="所属中队">
          <el-select v-model="form.OrgID" placeholder="请选择中队" style="width: 220px">
            <el-option v-for="org in organOptions" :key="org.ID" :label="org.Name" :value="org.ID" />
          </el-select>
        </el-form-item>

        <el-form-item label="角色">
          <el-select v-model="form.Role" style="width: 220px">
            <el-option v-for="opt in roleOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="dialogMode==='edit'" label="状态">
          <el-select v-model="form.IsDeleted" style="width: 220px">
            <el-option label="正常" :value="0" />
            <el-option label="已删除" :value="1" />
          </el-select>
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
        <!-- 只要是“改自己”的密码（包括管理员改自己） => 必须输入旧密码 -->
        <el-form-item v-if="needOldPwd" label="旧密码">
          <el-input v-model="resetForm.old_password" type="password" show-password />
        </el-form-item>

        <el-form-item label="设置密码">
          <el-input v-model="resetForm.new_password" type="password" show-password placeholder="至少8位" />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="resetForm.confirm_new_password" type="password" show-password placeholder="再次输入新密码" />
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
import { OrganAPI, type Organ } from '@/api/organ'
import { notifyError } from '@/utils/notify'
import {
  ROLE_ADMIN, ROLE_USER, ROLE_LABELS, roleLabel
} from '@/utils/role'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { getToken } from '@/api/http'

// ====== 类型定义（与后端 JSON 对齐）======
interface Row {
  ID: string
  Username: string
  OrgID: string    // 所属机构ID
  OrgName?: string // 所属机构名称（用于显示）
  IsDeleted: number // 0正常 1已删除
  Role: number     // 0用户 1管理员
  LastLoginAt?: string | null
  CreatedAt?: string
  UpdatedAt?: string
}

// ====== 当前登录用户（示例：替换为你的实际用户来源）======

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

// 是否管理员（布尔）
const isAdminRef = computed(() => (currentUser.value?.Role === ROLE_ADMIN))

// 下拉选项
const roleOptions = [
  { value: ROLE_ADMIN, label: ROLE_LABELS[ROLE_ADMIN] },
  { value: ROLE_USER,  label: ROLE_LABELS[ROLE_USER]  },
]

// ====== 列表/分页 ======
const rows = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const keyword = ref('')

const tableLoading = ref(false)

// ====== 组织数据 ======
const organOptions = ref<Organ[]>([])
const organLoading = ref(false)

const fetchList = async () => {
  try {
    tableLoading.value = true
    const limit = pageSize.value
    const offset = (page.value - 1) * pageSize.value
    const body = {
      username_like: keyword.value || undefined,
      limit,
      offset,
    }
    const { data } = await AccountAPI.list(body)
    rows.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    notifyError(e)
  } finally {
    tableLoading.value = false
  }
}

const fetchOrgans = async () => {
  try {
    organLoading.value = true
    const { data } = await OrganAPI.list({ limit: 1000 })
    organOptions.value = data.items || []
  } catch (e) {
    notifyError(e)
  } finally {
    organLoading.value = false
  }
}
const onSearch = () => { page.value = 1; fetchList() }

// ====== 新增/编辑（仅管理员可新增）======
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)

const form = ref<Partial<Row> & { password?: string; confirmPassword?: string }>({
  Role: ROLE_USER,
  IsDeleted: 0,
  OrgID: '',
})

const openCreate = () => {
  // 代码兜底：即便按钮被绕过，也不允许非管理员进入新增
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可新增账户')
    return
  }
  dialogMode.value = 'create'
  form.value = {
    Username: '',
    password: '',
    confirmPassword: '',
    Role: ROLE_USER,
    IsDeleted: 0,
    OrgID: ''
  }
  dialogVisible.value = true
}
const openEdit = (row: Row) => {
  dialogMode.value = 'edit'
  form.value = {
    ID: row.ID,
    Username: row.Username,
    Role: row.Role,
    IsDeleted: row.IsDeleted,
    OrgID: row.OrgID
  }
  dialogVisible.value = true
}
const onSubmit = async () => {
  try {
    submitLoading.value = true
    if (dialogMode.value === 'create') {
      const { Username, password, confirmPassword } = form.value
      if (!Username?.trim()) {
        ElMessage.warning('请输入用户名'); return
      }
      if (!password) {
        ElMessage.warning('请输入密码'); return
      }
      if (!confirmPassword) {
        ElMessage.warning('请输入确认密码'); return
      }
      if (password.length < 8) {
        ElMessage.error('密码至少8位'); return
      }
      if (password !== confirmPassword) {
        ElMessage.error('两次输入的密码不一致'); return
      }

      await AccountAPI.create({
        username: Username.trim(),
        password,
        org_id: form.value.OrgID!,
        role: Number(form.value.Role ?? ROLE_USER)
      })
      ElMessage.success('创建成功')
    } else {
      const id = form.value.ID!
      await AccountAPI.update({ 
        id, 
        org_id: form.value.OrgID,
        role: Number(form.value.Role ?? ROLE_USER)
      })
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

// ====== 重置/修改密码 ======
const resetVisible = ref(false)
const resetLoading = ref(false)
const resetForm = ref<{
  id?: string
  username?: string
  old_password?: string
  new_password: string
  confirm_new_password?: string
}>({ new_password: '', confirm_new_password: '' })

// 谁可以打开“修改密码”对话框：管理员或本人
const canResetPassword = (row: Row) => {
  return isAdminRef.value || currentUser.value?.ID === row.ID
}
const openResetGuard = (row: Row) => {
  if (!canResetPassword(row)) {
    ElMessage.warning('无权限修改他人密码')
    return
  }
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

// 规则：只要“改自己”（管理员也包括）=> 需要旧密码；管理员改别人 => 不需要
const needOldPwd = computed(() => {
  const current = currentUser.value?.Username
  const target = resetForm.value.username
  console.log('[调试] 当前用户:', current, '目标用户:', target, '是否为自己:', current === target)
  return current === target
})

const resetDialogTitle = computed(() =>
  needOldPwd.value ? '修改密码' : '重置密码'
)

// 在 <script setup> 中
const isSelf = (row: Row) => {
  return currentUser.value?.ID === row.ID // ✅ 推荐用 ID 比较（唯一）
  // 或：return currentUser.value?.Username === row.Username
}

const isUser = (row: Row) => {
  return row.Role === ROLE_USER
}

const onDoReset = async () => {
  try {
    resetLoading.value = true

    const newPwd = resetForm.value.new_password || ''
    const confirm = resetForm.value.confirm_new_password || ''
    if (newPwd.length < 8) {
      ElMessage.error('新密码至少8位'); return
    }
    if (newPwd !== confirm) {
      ElMessage.error('两次输入的新密码不一致'); return
    }

    if (needOldPwd.value) {
      // 改自己：必须旧密码
      if (!resetForm.value.old_password) {
        ElMessage.warning('请输入旧密码'); return
      }
      await AccountAPI.change_password({
        username: resetForm.value.username!,
        old_password: resetForm.value.old_password!,
        new_password: newPwd,
      })
    } else {
      // 管理员改别人：无需旧密码
      if (!isAdminRef.value) {
        ElMessage.error('无权限修改'); return
      }
      await AccountAPI.update_password({ id: resetForm.value.id!, password: newPwd })
    }

    resetVisible.value = false
    ElMessage.success('密码已更新')
  } catch (e) {
    notifyError(e)
  } finally {
    resetLoading.value = false
  }
}

// ====== 删除（仅管理员）======
const deletingId = ref<string | null>(null)
const onDelete = async (row: Row) => {
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可删除账户')
    return
  }
  try {
    await ElMessageBox.confirm(`确认删除账户 “${row.Username}” ?`, '提示', { type: 'warning' })
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

onMounted(() => {
  fetchList()
  fetchOrgans()
})
</script>

<style scoped>
.toolbar { display:flex; gap:12px; align-items:center; margin-bottom:12px; }
.spacer { flex:1; }
.pager { display:flex; justify-content:flex-end; padding-top:12px; }
</style>
