<!-- src/pages/Organizations.vue -->
<template>
  <div>
    <h2 style="margin:8px 0 16px;">中队管理</h2>

    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索中队名称"
        clearable
        @clear="onSearch"
        @keyup.enter="onSearch"
        style="width:240px"
      />
      <div class="spacer" />
      <el-button type="primary" @click="openCreate" :disabled="!isAdminRef">+ 新增中队</el-button>
    </div>

    <el-table :data="rows" stripe v-loading="tableLoading">
      <el-table-column type="index" label="序号" width="80" />
      <el-table-column prop="Name" label="中队名称" min-width="180" />
      <el-table-column prop="Code" label="编码" width="120" />
      <el-table-column prop="Parent" label="上级组织ID" width="180" />
      <el-table-column prop="Description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="Sort" label="排序" width="100" align="center" />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.IsDeleted === 0 ? 'success' : 'info'">
            {{ row.IsDeleted === 0 ? '正常' : '已删除' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" align="center" fixed="right">
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
        @current-change="(p:number)=>{page=p;fetchList()}"
      />
    </div>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode==='create'?'新增中队':'编辑中队'" width="560px">
      <el-form :model="form" label-width="100px" v-loading="submitLoading">
        <el-form-item label="中队名称" required>
          <el-input v-model="form.Name" placeholder="请输入中队名称" />
        </el-form-item>

        <el-form-item label="组织编码">
          <el-input v-model="form.Code" placeholder="留空则自动生成" />
        </el-form-item>

        <el-form-item label="上级组织">
          <el-select v-model="form.Parent" clearable placeholder="请选择上级组织（留空为根组织）" style="width:100%">
            <el-option 
              v-for="org in allOrgs" 
              :key="org.ID" 
              :label="org.Name" 
              :value="org.ID"
              :disabled="dialogMode === 'edit' && org.ID === form.ID"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="描述">
          <el-input 
            v-model="form.Description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>

        <el-form-item label="排序">
          <el-input-number v-model="form.Sort" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false" :disabled="submitLoading">取消</el-button>
        <el-button type="primary" @click="onSubmit" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { OrganAPI } from '@/api/acl'
import { notifyError } from '@/utils/notify'
import { ROLE_ADMIN } from '@/utils/role'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { getToken } from '@/api/http'

// ====== 类型定义 ======
interface Organ {
  ID: string
  Name: string
  Code?: string | null
  Parent: string
  Description?: string
  Sort: number
  IsDeleted: number
  CreatedAt?: string
  UpdatedAt?: string
}

// ====== 权限控制 ======
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})

const isAdminRef = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

// ====== 列表/分页 ======
const rows = ref<Organ[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const keyword = ref('')
const tableLoading = ref(false)

// 所有组织列表（用于上级组织下拉）
const allOrgs = ref<Organ[]>([])

const fetchList = async () => {
  try {
    tableLoading.value = true
    const limit = pageSize.value
    const offset = (page.value - 1) * pageSize.value
    const body = {
      name_like: keyword.value || undefined,
      is_deleted: 0, // 只显示未删除的
      limit,
      offset,
    }
    const { data } = await OrganAPI.list(body)
    rows.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    notifyError(e)
  } finally {
    tableLoading.value = false
  }
}

// 获取所有组织（用于上级组织下拉）
const fetchAllOrgs = async () => {
  try {
    const { data } = await OrganAPI.list({ limit: 1000, offset: 0, is_deleted: 0 })
    allOrgs.value = data.items || []
  } catch (e) {
    notifyError(e)
  }
}

const onSearch = () => { page.value = 1; fetchList() }

// ====== 新增/编辑 ======
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)

const form = ref<Partial<Organ>>({
  Name: '',
  Code: '',
  Parent: '',
  Description: '',
  Sort: 0,
})

const openCreate = () => {
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可新增中队')
    return
  }
  dialogMode.value = 'create'
  form.value = {
    Name: '',
    Code: '',
    Parent: '',
    Description: '',
    Sort: 0,
  }
  fetchAllOrgs()
  dialogVisible.value = true
}

const openEdit = (row: Organ) => {
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可编辑中队')
    return
  }
  dialogMode.value = 'edit'
  form.value = {
    ID: row.ID,
    Name: row.Name,
    Code: row.Code || '',
    Parent: row.Parent || '',
    Description: row.Description || '',
    Sort: row.Sort,
  }
  fetchAllOrgs()
  dialogVisible.value = true
}

const onSubmit = async () => {
  try {
    submitLoading.value = true
    
    if (!form.value.Name?.trim()) {
      ElMessage.warning('请输入中队名称')
      return
    }

    if (dialogMode.value === 'create') {
      await OrganAPI.create({
        name: form.value.Name.trim(),
        parent: form.value.Parent || '',
        code: form.value.Code || undefined,
        description: form.value.Description || undefined,
        sort: form.value.Sort || 0,
      })
      ElMessage.success('创建成功')
    } else {
      await OrganAPI.update({
        id: form.value.ID!,
        name: form.value.Name.trim(),
        parent: form.value.Parent || '',
        code: form.value.Code || undefined,
        description: form.value.Description || undefined,
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

// ====== 删除 ======
const deletingId = ref<string | null>(null)
const onDelete = async (row: Organ) => {
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可删除中队')
    return
  }
  try {
    await ElMessageBox.confirm(`确认删除中队 "${row.Name}" ?`, '提示', { type: 'warning' })
    deletingId.value = row.ID
    await OrganAPI.delete(row.ID)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e: any) {
    if (e?.message) notifyError(e)
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  fetchList()
  fetchAllOrgs()
})
</script>

<style scoped>
.toolbar { display:flex; gap:12px; align-items:center; margin-bottom:12px; }
.spacer { flex:1; }
.pager { display:flex; justify-content:flex-end; padding-top:12px; }
</style>
