<template>
  <div class="page-market-price">
    <!-- 顶部查询条件 -->
    <div class="search-bar">
      <div class="search-filters">
        <el-select v-model="searchYear" placeholder="选择年份" clearable style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="`${y}年`" :value="y" />
        </el-select>
        <el-select v-model="searchMonth" placeholder="选择月份" clearable style="width: 120px">
          <el-option v-for="m in monthOptions" :key="m" :label="`${m}月`" :value="m" />
        </el-select>
        <el-select v-model="searchTenDay" placeholder="选择旬" clearable style="width: 120px">
          <el-option label="上旬" :value="1" />
          <el-option label="中旬" :value="2" />
          <el-option label="下旬" :value="3" />
        </el-select>
        <el-input
          v-model="keyword"
          placeholder="请输入"
          clearable
          style="width: 280px"
          @clear="onSearch"
          @keyup.enter="onSearch"
        />
        <el-button @click="onSearch">查询</el-button>
      </div>
      <div class="search-actions">
        <el-button @click="onEdit" :disabled="!selectedInquiry || !isAdmin" plain>编辑</el-button>
        <el-button @click="onCreate" :disabled="!isAdmin" plain>+ 新建</el-button>
        <el-button @click="onImport" :disabled="!isAdmin" plain>+ 导入Excel</el-button>
        <el-button @click="onExport" :disabled="!selectedInquiry" plain>导出</el-button>
      </div>
    </div>

    <!-- 询价单列表 -->
    <div class="inquiry-list-section">
      <div class="section-title">询价单列表</div>
      <el-table
        :data="inquiries"
        stripe
        v-loading="inquiryLoading"
        highlight-current-row
        @current-change="onInquirySelect"
        :header-cell-style="{ background: '#f3f4f6' }"
      >
        <el-table-column type="index" label="序号" width="70" :index="indexMethod" />
        <el-table-column prop="InquiryTitle" label="询价单标题" min-width="240" />
        <el-table-column label="业务日期" width="140">
          <template #default="{ row }">{{ formatDate(row.InquiryDate) }}</template>
        </el-table-column>
        <el-table-column label="年份" width="100">
          <template #default="{ row }">{{ row.InquiryYear }}</template>
        </el-table-column>
        <el-table-column label="月份" width="100">
          <template #default="{ row }">{{ row.InquiryMonth }}</template>
        </el-table-column>
        <el-table-column label="旬" width="100">
          <template #default="{ row }">{{ tenDayLabel(row.InquiryTenDay) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" @click="onDeleteInquiry(row)" :disabled="!isAdmin">删除</el-button>
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
    </div>

    <!-- 询价商品明细列表 -->
    <div class="inquiry-items-section" v-if="selectedInquiry">
      <div class="section-title">
        商品价格明细 - {{ selectedInquiry.InquiryTitle }}
      </div>
      <el-table
        :data="inquiryItems"
        stripe
        v-loading="itemsLoading"
        :header-cell-style="{ background: '#f3f4f6' }"
        max-height="500"
      >
        <el-table-column type="index" label="序号" width="70" />
        <el-table-column label="商品图" width="90">
          <template #default>
            <div class="img-ph">无</div>
          </template>
        </el-table-column>
        <el-table-column prop="GoodsNameSnap" label="品名" min-width="140" />
        <el-table-column prop="CategoryNameSnap" label="品类" width="120" />
        <el-table-column prop="SpecNameSnap" label="规格标准" width="120">
          <template #default="{ row }">{{ row.SpecNameSnap || '—' }}</template>
        </el-table-column>
        <el-table-column prop="UnitNameSnap" label="单位" width="100">
          <template #default="{ row }">{{ row.UnitNameSnap || '—' }}</template>
        </el-table-column>
        <el-table-column prop="GuidePrice" label="指导价" width="100">
          <template #default="{ row }">{{ formatPrice(row.GuidePrice) }}</template>
        </el-table-column>
        <el-table-column prop="LastMonthAvgPrice" label="上期均价" width="100">
          <template #default="{ row }">{{ formatPrice(row.LastMonthAvgPrice) }}</template>
        </el-table-column>
        <el-table-column prop="CurrentAvgPrice" label="本期均价" width="100">
          <template #default="{ row }">{{ formatPrice(row.CurrentAvgPrice) }}</template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 导入Excel弹窗 -->
    <el-dialog v-model="importDialogVisible" title="导入Excel" width="520px">
      <div class="import-dialog-content">
        <el-upload
          ref="uploadRef"
          :auto-upload="false"
          :limit="1"
          accept=".xlsx,.xls"
          :on-change="onFileChange"
          :file-list="fileList"
          drag
        >
          <el-icon class="upload-icon"><UploadFilled /></el-icon>
          <div class="upload-text">将文件拖到此处，或<em>点击上传</em></div>
          <template #tip>
            <div class="upload-tip">只能上传 xlsx/xls 文件</div>
          </template>
        </el-upload>
      </div>
      <template #footer>
        <el-button @click="importDialogVisible=false" :disabled="uploading">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="onSubmitImport">确定导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'
import { InquiryAPI, InquiryItemAPI, type InquiryRow, type InquiryItemRow } from '@/api/market'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { ROLE_ADMIN } from '@/utils/role'
import { notifyError } from '@/utils/notify'
const indexMethod = (rowIndex: number) =>
  (page.value - 1) * pageSize.value + rowIndex + 1
// 登录信息
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const organId = computed(() => jwtPayload.value?.org_id || '')
const isAdmin = computed(() => jwtPayload.value?.role === ROLE_ADMIN)

// 查询条件
const searchYear = ref<number | undefined>()
const searchMonth = ref<number | undefined>()
const searchTenDay = ref<number | undefined>()
const keyword = ref('')

// 年份和月份选项
const currentYear = new Date().getFullYear()
const yearOptions = Array.from({ length: 10 }, (_, i) => currentYear - i)
const monthOptions = Array.from({ length: 12 }, (_, i) => i + 1)

// 询价单列表
const page = ref(1)
const pageSize = ref(15)
const pageSizes = [10, 15, 20, 50]
const total = ref(0)
const inquiries = ref<InquiryRow[]>([])
const inquiryLoading = ref(false)
const selectedInquiry = ref<InquiryRow | null>(null)

// 询价商品明细列表
const inquiryItems = ref<InquiryItemRow[]>([])
const itemsLoading = ref(false)

// 分页处理
const handlePageChange = (p: number) => {
  page.value = p
  fetchInquiries()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  page.value = 1
  fetchInquiries()
}

// 获取询价单列表
const fetchInquiries = async () => {
  if (!organId.value) {
    inquiries.value = []
    total.value = 0
    return
  }
  inquiryLoading.value = true
  try {
    const params: any = {
      org_id: organId.value,
      page: page.value,
      page_size: pageSize.value,
    }
    if (searchYear.value) params.year = searchYear.value
    if (searchMonth.value) params.month = searchMonth.value
    if (searchTenDay.value) params.ten_day = searchTenDay.value
    if (keyword.value.trim()) params.keyword = keyword.value.trim()

    const { data } = await InquiryAPI.list(params)
    inquiries.value = data?.items || []
    total.value = Number(data?.total || 0)

    // 如果当前选中的询价单不在列表中，清除选择
    if (selectedInquiry.value && !inquiries.value.some(i => i.ID === selectedInquiry.value?.ID)) {
      selectedInquiry.value = null
      inquiryItems.value = []
    }
  } catch (e) {
    notifyError(e)
  } finally {
    inquiryLoading.value = false
  }
}

// 获取询价商品明细
const fetchInquiryItems = async (inquiryId: string) => {
  itemsLoading.value = true
  try {
    const params = {
      inquiry_id: inquiryId,
      page: 1,
      page_size: 1000,
    }
    const { data } = await InquiryItemAPI.list(params)
    inquiryItems.value = data?.items || []
  } catch (e) {
    notifyError(e)
  } finally {
    itemsLoading.value = false
  }
}

// 选择询价单
const onInquirySelect = (row: InquiryRow | null) => {
  selectedInquiry.value = row
  if (row) {
    fetchInquiryItems(row.ID)
  } else {
    inquiryItems.value = []
  }
}

// 查询
const onSearch = () => {
  page.value = 1
  fetchInquiries()
}

// 编辑
const onCreate = () => {
  ElMessage.info('新建功能暂未开放')
}

// 新建
const onEdit = () => {
  ElMessage.info('编辑功能暂未开放')
}

// 导出
const onExport = () => {
  ElMessage.info('导出功能暂未开放')
}

// 删除询价单
const onDeleteInquiry = async (row: InquiryRow) => {
  try {
    await ElMessageBox.confirm(`确认删除询价单 "${row.InquiryTitle}" ?`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await InquiryAPI.remove(row.ID)
    ElMessage.success('删除成功')
    if (selectedInquiry.value?.ID === row.ID) {
      selectedInquiry.value = null
      inquiryItems.value = []
    }
    await fetchInquiries()
  } catch (e) {
    notifyError(e)
  }
}

// 导入Excel
const importDialogVisible = ref(false)
const fileList = ref<UploadFile[]>([])
const uploadRef = ref()
const uploading = ref(false)

const onImport = () => {
  fileList.value = []
  importDialogVisible.value = true
}

const onFileChange = (file: UploadFile) => {
  fileList.value = [file]
}

const onSubmitImport = async () => {
  if (fileList.value.length === 0) {
    ElMessage.warning('请选择要上传的文件')
    return
  }

  const uploadFile = fileList.value[0]
  if (!uploadFile || !uploadFile.raw) {
    ElMessage.warning('文件不存在')
    return
  }
  
  const file = uploadFile.raw

  if (!organId.value) {
    ElMessage.warning('缺少组织信息')
    return
  }

  const formData = new FormData()
  formData.append('file', file)
  formData.append('org_id', organId.value)

  uploading.value = true
  try {
    // 调用导入接口
    const response = await fetch('/api/v1/inquiry_import/import_inquiry', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getToken()}`,
      },
      body: formData,
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || '导入失败')
    }

    const result = await response.json()
    ElMessage.success(result.message || '导入成功')
    importDialogVisible.value = false
    fileList.value = []
    await fetchInquiries()
  } catch (e: any) {
    ElMessage.error(e.message || '导入失败')
  } finally {
    uploading.value = false
  }
}

// 格式化日期
const formatDate = (dateStr: string) => {
  if (!dateStr) return '—'
  return dateStr.split('T')[0]
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '—'
  return dateStr.replace('T', ' ').substring(0, 19)
}

const tenDayLabel = (tenDay?: number) => {
  if (!tenDay) return '—'
  return ['上旬', '中旬', '下旬'][tenDay - 1] || '—'
}

const formatPrice = (price?: number | null) => {
  if (price === null || price === undefined) return '—'
  return price.toFixed(2)
}

// 初始化
onMounted(() => {
  fetchInquiries()
})
</script>

<style scoped>
.page-market-price {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: calc(100vh - 120px);
  min-height: 520px;
}

.search-bar {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.search-filters {
  display: flex;
  gap: 12px;
  align-items: center;
  flex: 1;
}

.search-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.inquiry-list-section,
.inquiry-items-section {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.inquiry-list-section {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.section-title {
  font-weight: 600;
  font-size: 16px;
  color: #333;
  padding: 4px 0;
}

.pager {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
}

.img-ph {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  background: #f5f7fa;
  color: #909399;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.import-dialog-content {
  padding: 20px 0;
}

.upload-icon {
  font-size: 67px;
  color: #c0c4cc;
  margin-bottom: 16px;
}

.upload-text {
  color: #606266;
  font-size: 14px;
  text-align: center;
}

.upload-text em {
  color: #409eff;
  font-style: normal;
}

.upload-tip {
  color: #909399;
  font-size: 12px;
  text-align: center;
  margin-top: 8px;
}
</style>
