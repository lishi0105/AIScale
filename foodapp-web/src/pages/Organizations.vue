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
      <!-- 仅管理员能新增 -->
      <el-button type="primary" @click="openCreate" :disabled="!isAdminRef">+ 新增中队</el-button>
    </div>

    <el-table :data="rows" stripe v-loading="tableLoading">
      <el-table-column type="index" label="序号" width="80" />
      <el-table-column prop="Name" label="中队名称" min-width="160" />
      <el-table-column prop="Code" label="中队编码" width="120" />
      <el-table-column prop="ParentName" label="上级中队" min-width="160" />
      <el-table-column prop="Description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="Sort" label="排序" width="80" align="center" />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.IsDeleted === 0 ? 'success' : 'info'">
            {{ row.IsDeleted === 0 ? '正常' : '已删除' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" align="center">
        <template #default="{ row }">
          <el-button link @click="openEdit(row)">编辑</el-button>
          <el-button
            link
            type="danger"
            :disabled="deletingId === row.ID"
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
    <el-dialog v-model="dialogVisible" :title="dialogMode==='create'?'新增中队':'编辑中队'" width="600px">
      <el-form :model="form" label-width="100px" v-loading="submitLoading">
        <el-form-item label="中队名称" required>
          <el-input v-model="form.Name" placeholder="请输入中队名称" />
        </el-form-item>

        <el-form-item label="中队编码">
          <el-input v-model="form.Code" placeholder="留空自动生成" />
        </el-form-item>

        <el-form-item label="上级中队">
          <el-select v-model="form.Parent" placeholder="请选择上级中队" style="width: 100%">
            <el-option label="根节点" value="" />
            <el-option 
              v-for="org in parentOptions" 
              :key="org.ID" 
              :label="org.Name" 
              :value="org.ID"
              :disabled="org.ID === form.ID"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="描述">
          <el-input 
            v-model="form.Description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入中队描述"
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
import { OrganAPI, type Organ } from '@/api/organ'
import { notifyError } from '@/utils/notify'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { getToken } from '@/api/http'

// ====== 类型定义 ======
interface Row extends Organ {
  ParentName?: string
}

// ====== 当前登录用户 ======
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
const isAdminRef = computed(() => (currentUser.value?.Role === 1)) // 1为管理员

// ====== 列表/分页 ======
const rows = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const keyword = ref('')

const tableLoading = ref(false)

// ====== 组织数据 ======
const organOptions = ref<Organ[]>([])
const parentOptions = computed(() => organOptions.value.filter(org => org.IsDeleted === 0))

const fetchList = async () => {
  try {
    tableLoading.value = true
    const limit = pageSize.value
    const offset = (page.value - 1) * pageSize.value
    const body = {
      name_like: keyword.value || undefined,
      limit,
      offset,
    }
    const { data } = await OrganAPI.list(body)
    
    // 处理数据，添加父级名称
    const items = (data.items || []).map((item: Organ) => {
      const parent = organOptions.value.find(org => org.ID === item.Parent)
      return {
        ...item,
        ParentName: parent?.Name || '根节点'
      }
    })
    
    rows.value = items
    total.value = data.total || 0
  } catch (e) {
    notifyError(e)
  } finally {
    tableLoading.value = false
  }
}

const fetchOrgans = async () => {
  try {
    const { data } = await OrganAPI.list({ limit: 1000 })
    organOptions.value = data.items || []
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
  dialogVisible.value = true
}

const openEdit = (row: Row) => {
  dialogMode.value = 'edit'
  form.value = {
    ID: row.ID,
    Name: row.Name,
    Code: row.Code || '',
    Parent: row.Parent,
    Description: row.Description,
    Sort: row.Sort,
  }
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
        code: form.value.Code || undefined,
        parent: form.value.Parent || undefined,
        description: form.value.Description || undefined,
        sort: form.value.Sort || 0,
      })
      ElMessage.success('创建成功')
    } else {
      await OrganAPI.update({
        id: form.value.ID!,
        name: form.value.Name?.trim(),
        code: form.value.Code || undefined,
        parent: form.value.Parent || undefined,
        description: form.value.Description || undefined,
      })
      ElMessage.success('保存成功')
    }
    
    dialogVisible.value = false
    fetchList()
    fetchOrgans() // 刷新组织列表
  } catch (e) {
    notifyError(e)
  } finally {
    submitLoading.value = false
  }
}

// ====== 删除 ======
const deletingId = ref<string | null>(null)
const onDelete = async (row: Row) => {
  if (!isAdminRef.value) {
    ElMessage.warning('仅管理员可删除中队')
    return
  }
  try {
    await ElMessageBox.confirm(`确认删除中队 "${row.Name}" ?`, '提示', { type: 'warning' })
    deletingId.value = row.ID
    await OrganAPI.remove(row.ID)
    ElMessage.success('删除成功')
    fetchList()
    fetchOrgans() // 刷新组织列表
  } catch (e:any) {
    if (e?.message) notifyError(e)
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  fetchOrgans()
  fetchList()
})
</script>

<style scoped>
.toolbar { display:flex; gap:12px; align-items:center; margin-bottom:12px; }
.spacer { flex:1; }
.pager { display:flex; justify-content:flex-end; padding-top:12px; }
</style>