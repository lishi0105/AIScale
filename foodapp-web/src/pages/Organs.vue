<template>
  <div class="page-organ">
    <h2 class="page-title">中队管理</h2>

    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索中队名称"
        clearable
        @clear="onSearch"
        @keyup.enter="onSearch"
        style="width: 260px"
      />
      <el-button class="toolbar-btn" type="primary" @click="openCreate" :disabled="!isAdmin">
        + 新增中队
      </el-button>
    </div>

    <el-table :data="rows" stripe v-loading="tableLoading">
      <el-table-column type="index" width="70" label="序号" />
      <el-table-column prop="Name" label="中队名称" min-width="180" />
      <el-table-column prop="Code" label="编码" min-width="120">
        <template #default="{ row }">
          <span>{{ row.Code || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="上级中队" min-width="180">
        <template #default="{ row }">
          <span>{{ parentName(row.Parent) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="Sort" label="排序" width="100" align="center" />
      <el-table-column prop="Description" label="描述" min-width="220">
        <template #default="{ row }">
          <span>{{ row.Description || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="120" align="center">
        <template #default="{ row }">
          <el-tag :type="row.IsDeleted === 0 ? 'success' : 'info'">
            {{ row.IsDeleted === 0 ? '正常' : '已删除' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" align="center">
        <template #default="{ row }">
          <el-button link @click="openEdit(row)" :disabled="!isAdmin">编辑</el-button>
          <el-button
            link
            type="danger"
            :disabled="!isAdmin || row.IsDeleted === 1 || deletingId === row.ID"
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

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <el-form :model="form" label-width="110px" v-loading="submitLoading">
        <el-form-item label="中队名称">
          <el-input v-model="form.name" placeholder="请输入中队名称" maxlength="64" />
        </el-form-item>

        <el-form-item label="上级中队">
          <el-select
            v-model="form.parent"
            placeholder="请选择上级中队"
            :loading="parentLoading"
            filterable
            style="width: 260px"
          >
            <el-option :value="''" label="（无上级）" />
            <el-option
              v-for="item in parentOptions"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="编码">
          <el-input
            v-model="form.code"
            placeholder="留空则自动生成"
            maxlength="32"
            clearable
          />
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            placeholder="可填写中队简介"
            :rows="3"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>

        <el-form-item
          v-if="dialogMode === 'create'"
          label="排序码"
          placeholder="留空则自动生成"
        >
          <el-input-number v-model="form.sort" :min="0" :max="9999" controls-position="right" />
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
import { computed, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { OrganAPI, type OrganRow } from '@/api/organ'
import type { OrganListParams } from '@/api/organ'
import { notifyError } from '@/utils/notify'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'

interface FormState {
  id: string
  name: string
  parent: string
  code: string
  description: string
  sort: number | null
}

const rows = ref<OrganRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const keyword = ref('')
const tableLoading = ref(false)
const deletingId = ref('')
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)
const parentLoading = ref(false)

const form = reactive<FormState>({
  id: '',
  name: '',
  parent: '',
  code: '',
  description: '',
  sort: null,
})

const allParents = ref<OrganRow[]>([])

const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

const dialogTitle = computed(() =>
  dialogMode.value === 'create' ? '新增中队' : '编辑中队'
)

const parentMap = computed<Map<string, OrganRow>>(() => {
  const map = new Map<string, OrganRow>()
  ;[...allParents.value, ...rows.value].forEach(item => map.set(item.ID, item))
  return map
})

const parentOptions = computed(() =>
  allParents.value
    .filter(item => item.ID !== form.id)
    .map(item => ({ value: item.ID, label: item.Name }))
)

const parentName = (id: string) => {
  if (!id) return '—'
  return parentMap.value.get(id)?.Name || '—'
}

const resetForm = () => {
  form.id = ''
  form.name = ''
  form.parent = ''
  form.code = ''
  form.description = ''
  form.sort = null
}

const fetchParents = async () => {
  try {
    parentLoading.value = true
    const { data } = await OrganAPI.list({ page: 1, page_size: 200, is_deleted: 0 })
    allParents.value = Array.isArray(data?.items) ? data.items : []
  } catch (error) {
    notifyError(error)
  } finally {
    parentLoading.value = false
  }
}

const fetchList = async () => {
  try {
    tableLoading.value = true
    const params: OrganListParams= {
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
      is_deleted: 0,
    }
    const { data } = await OrganAPI.list(params)
    rows.value = Array.isArray(data?.items) ? data.items : []
    total.value = Number(data?.total || 0)
  } catch (error) {
    notifyError(error)
  } finally {
    tableLoading.value = false
  }
}

const onSearch = () => {
  page.value = 1
  fetchList()
}
const onPageChange = (p: number) => {
  page.value = p
  fetchList()
}

const openCreate = async () => {
  if (!isAdmin.value) return
  dialogMode.value = 'create'
  resetForm()
  await fetchParents()
  dialogVisible.value = true
}

const openEdit = async (row: OrganRow) => {
  if (!isAdmin.value) return
  dialogMode.value = 'edit'
  await fetchParents()
  form.id = row.ID
  form.name = row.Name
  form.parent = row.Parent || ''
  form.code = row.Code || ''
  form.description = row.Description || ''
  form.sort = row.Sort
  dialogVisible.value = true
}

const buildCreatePayload = () => {
  const name = form.name.trim()
  if (!name) {
    ElMessage.error('请输入中队名称')
    return null
  }
  return {
    name,
    parent: form.parent || undefined,
    code: form.code.trim() || undefined,
    description: form.description.trim() || undefined,
    sort: form.sort ?? undefined,
  }
}

const buildUpdatePayload = () => {
  const name = form.name.trim()
  if (!name) {
    ElMessage.error('请输入中队名称')
    return null
  }
  return {
    id: form.id,
    name,
    parent: form.parent || '',
    code: form.code.trim() || undefined,
    description: form.description.trim() || undefined,
  }
}

const onSubmit = async () => {
  try {
    submitLoading.value = true
    if (dialogMode.value === 'create') {
      const payload = buildCreatePayload()
      if (!payload) return
      await OrganAPI.create(payload)
      ElMessage.success('新增成功')
    } else {
      const payload = buildUpdatePayload()
      if (!payload) return
      await OrganAPI.update(payload)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    await Promise.all([fetchList(), fetchParents()])
  } catch (error) {
    notifyError(error)
  } finally {
    submitLoading.value = false
  }
}

const onDelete = async (row: OrganRow) => {
  try {
    await ElMessageBox.confirm(`确认删除中队“${row.Name}”吗？`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    deletingId.value = row.ID
    await OrganAPI.softDelete(row.ID)
    ElMessage.success('删除成功')
    await Promise.all([fetchList(), fetchParents()])
  } catch (error) {
    notifyError(error)
  } finally {
    deletingId.value = ''
  }
}

fetchList()
fetchParents()
</script>

<style scoped>
.page-organ {
  display: flex;
  flex-direction: column;
}
.page-title {
  margin: 8px 0 16px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.toolbar-btn {
  margin-left: auto;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
