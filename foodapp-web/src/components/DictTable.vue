<template>
  <div class="card">
    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索名称"
        clearable
        @clear="onSearch"
        @keyup.enter="onSearch"
        style="width: 240px"
      />
      <div class="spacer" />
      <el-button type="primary" @click="openCreate">+ 新增{{ label }}</el-button>
    </div>

    <el-table
      :data="rows"
      stripe
      style="width:100%"
      :header-cell-style="{ background: '#f3f4f6' }"
      v-loading="tableLoading"
    >
      <el-table-column type="index" label="序号" width="80" />
      <el-table-column prop="Name" :label="label" min-width="160" />
      <el-table-column prop="Sort" label="排序码" width="120" />
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
        layout="sizes, prev, pager, next, jumper, ->, total"
        :page-sizes="pageSizes"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="onPageChange"
        @size-change="onPageSizeChange"
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增' + label : '编辑' + label"
      width="420px"
    >
      <el-form :model="form" label-width="80px" v-loading="submitLoading">
        <el-form-item label="名称">
          <el-input v-model="form.Name" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="排序码">
          <el-input-number v-model="form.Sort" :min="0" :step="1" />
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
import { ElMessageBox, ElMessage } from 'element-plus'
import { ref, onMounted } from 'vue'
import { notifyError } from '@/utils/notify'

// 字段名与后端保持一致（首字母大写）
interface Row {
  ID: string
  Name: string
  Sort: number
}

const props = defineProps<{
  label: string
  // 保持你的 props 形状：父组件传入具体 API
  list: (params: { keyword?: string; page?: number; page_size?: number }) => Promise<any>
  create: (data: { Name: string; Sort?: number }) => Promise<any>
  update: (data: { ID: string; Name: string; Sort?: number }) => Promise<any>
  remove: (id: string) => Promise<any>
}>()

const emit = defineEmits<{
  // 提交成功后通知父组件（可用于外层刷新等）
  (e: 'changed'): void
}>()

const rows = ref<Row[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const pageSizes = [10, 15, 20, 50]
const keyword = ref('')

const tableLoading = ref(false)
const deletingId = ref<string | null>(null)

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const submitLoading = ref(false)
const form = ref<Partial<Row>>({ Name: '', Sort: 0 })

const fetchList = async () => {
  try {
    tableLoading.value = true
    const res = await props.list({
      keyword: keyword.value?.trim() || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    // 兼容两种返回结构
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

const onPageChange = (p: number) => {
  page.value = p
  fetchList()
}

const onPageSizeChange = (size: number) => {
  pageSize.value = size
  page.value = 1
  fetchList()
}

const openCreate = () => {
  dialogMode.value = 'create'
  form.value = { Name: '', Sort: 0 }
  dialogVisible.value = true
}

const openEdit = (row: Row) => {
  dialogMode.value = 'edit'
  form.value = { ...row }
  dialogVisible.value = true
}

const onSubmit = async () => {
  // 简单前置校验
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

  try {
    submitLoading.value = true
    if (dialogMode.value === 'create') {
      await props.create({ Name: name, Sort: sort })
      ElMessage.success('创建成功')
    } else {
      await props.update({ ID: String(form.value.ID), Name: name, Sort: sort })
      ElMessage.success('保存成功')
    }
    dialogVisible.value = false
    emit('changed')
    fetchList()
  } catch (err) {
    notifyError(err)
  } finally {
    submitLoading.value = false
  }
}

const onDelete = async (row: Row) => {
  try {
    await ElMessageBox.confirm(`确认删除 “${row.Name}” ?`, '提示', { type: 'warning' })
    deletingId.value = row.ID
    await props.remove(row.ID)
    ElMessage.success('删除成功')
    emit('changed')
    fetchList()
  } catch (err) {
    // 用户点击取消不会有 err；只有接口错误才提示
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
