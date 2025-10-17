<template>
  <div>
    <h2 style="margin: 8px 0 16px;">商品分类</h2>
    <div class="card">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索分类名称、编码或拼音"
          clearable
          @clear="onSearch"
          @keyup.enter="onSearch"
          style="width: 280px"
        />
        <div class="spacer" />
        <el-button type="primary" @click="openCreate">+ 新增分类</el-button>
      </div>

      <el-table
        :data="rows"
        stripe
        style="width:100%"
        :header-cell-style="{ background: '#f3f4f6' }"
        v-loading="tableLoading"
      >
        <el-table-column type="index" label="序号" width="80" />
        <el-table-column prop="Name" label="分类名称" min-width="120" />
        <el-table-column prop="Code" label="分类编码" min-width="100" />
        <el-table-column prop="Pinyin" label="拼音" min-width="120" />
        <el-table-column prop="CreatedAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
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
          @current-change="(p: number) => { page = p; fetchList(); }"
        />
      </div>

      <el-dialog
        v-model="dialogVisible"
        :title="dialogMode === 'create' ? '新增分类' : '编辑分类'"
        width="500px"
      >
        <el-form :model="form" label-width="100px" v-loading="submitLoading">
          <el-form-item label="分类名称" required>
            <el-input v-model="form.name" maxlength="64" show-word-limit placeholder="例如：蔬菜类" />
          </el-form-item>
          <el-form-item label="分类编码">
            <el-input v-model="form.code" maxlength="64" show-word-limit placeholder="例如：VEG" />
          </el-form-item>
          <el-form-item label="拼音">
            <el-input v-model="form.pinyin" maxlength="64" show-word-limit placeholder="例如：shucai" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false" :disabled="submitLoading">取消</el-button>
          <el-button type="primary" @click="onSubmit" :loading="submitLoading">确定</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessageBox, ElMessage } from 'element-plus'
import { ref, onMounted } from 'vue'
import { notifyError } from '@/utils/notify'
import { CategoryAPI } from '@/api/category'

interface Category {
  ID: string
  Name: string
  Code?: string
  Pinyin?: string
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

interface CategoryForm {
  id?: string
  name: string
  code?: string
  pinyin?: string
}

const rows = ref<Category[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const keyword = ref('')

const tableLoading = ref(false)
const deletingId = ref<string | null>(null)

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)
const form = ref<CategoryForm>({ name: '', code: '', pinyin: '' })

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const fetchList = async () => {
  try {
    tableLoading.value = true
    const res = await CategoryAPI.listCategories({
      keyword: keyword.value?.trim() || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    const data = res?.data ?? res
    rows.value = data?.items || []
    total.value = data?.total || 0
  } catch (err) {
    notifyError(err)
  } finally {
    tableLoading.value = false
  }
}

const onSearch = () => {
  page.value = 1
  fetchList()
}

const openCreate = () => {
  dialogMode.value = 'create'
  form.value = { name: '', code: '', pinyin: '' }
  dialogVisible.value = true
}

const openEdit = (row: Category) => {
  dialogMode.value = 'edit'
  form.value = {
    id: row.ID,
    name: row.Name,
    code: row.Code || '',
    pinyin: row.Pinyin || ''
  }
  dialogVisible.value = true
}

const onSubmit = async () => {
  const name = form.value.name?.trim()
  if (!name) {
    ElMessage.warning('请输入分类名称')
    return
  }

  try {
    submitLoading.value = true
    const data = {
      name,
      code: form.value.code?.trim() || undefined,
      pinyin: form.value.pinyin?.trim() || undefined
    }
    
    if (dialogMode.value === 'create') {
      await CategoryAPI.createCategory(data)
      ElMessage.success('创建成功')
    } else {
      await CategoryAPI.updateCategory({ id: form.value.id!, ...data })
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

const onDelete = async (row: Category) => {
  try {
    await ElMessageBox.confirm(`确认删除分类 "${row.Name}" ?`, '提示', { type: 'warning' })
    deletingId.value = row.ID
    await CategoryAPI.deleteCategory(row.ID)
    ElMessage.success('删除成功')
    fetchList()
  } catch (err) {
    if ((err as any)?.message) notifyError(err)
  } finally {
    deletingId.value = null
  }
}

onMounted(fetchList)
</script>

<style scoped>
.card {
  background: #fff;
  padding: 12px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}
.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}
.spacer {
  flex: 1;
}
.pager {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
}
</style>
