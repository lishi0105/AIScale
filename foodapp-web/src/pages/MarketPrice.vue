<template>
  <div class="page-market-price">
    <!-- 顶部：标题(左) + 年月旬/搜索(中) + 功能按钮(右) 三组等间距 -->
    <div class="header-bar">
      <!-- 左：标题（name/sub 竖排） -->
      <div class="title">
        <span class="name">{{ selectedInquiry?.InquiryTitle }}</span>
        <span class="sub">（最近更新时间：{{ formatDateTime(selectedInquiry?.UpdatedAt || '') }}）</span>
      </div>

      <!-- 中：筛选/搜索 -->
      <div class="filters">
        <el-select v-model="searchYear" placeholder="选择年份" clearable style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="`${y}年`" :value="y" />
        </el-select>
        <el-select v-model="searchMonth" placeholder="选择月份" clearable style="width: 120px">
          <el-option v-for="m in monthOptions" :key="m" :label="`${m}月`" :value="m" />
        </el-select>
        <el-select v-model="searchTenDay" placeholder="选择旬" clearable style="width: 120px">
          <el-option v-for="t in tenDayOptions" :key="t" :label="tenDayLabel(t)" :value="t" />
        </el-select>
        <el-input
          v-model="keyword"
          placeholder="请输入"
          clearable
          style="width: 240px"
          @clear="onSearch"
          @keyup.enter="onSearch"
        />
        <el-button @click="onSearch">查询</el-button>
      </div>

      <!-- 右：动作按钮 -->
      <div class="actions">
        <el-button type="primary" @click="onCreateClick" plain>新增</el-button>
        <el-button type="primary" @click="onEditClick" plain>编辑</el-button>
        <el-button type="primary" @click="onImportClick" plain>导入</el-button>
        <el-button type="primary" @click="onExportClick" plain>导出</el-button>
      </div>
    </div>

    <!-- 询价商品明细列表 -->
    <div class="inquiry-items-section" v-if="selectedInquiry">
      <div class="detail-body">
        <!-- 左侧：品类面板 -->
        <aside class="category-panel">
          <div class="panel-title-line">商品品类</div>
          <div class="category-list">
            <div class="category-row" :class="{ active: (selectedCategory || 'ALL')==='ALL' }" @click="onCategorySelect('ALL')">
              <span class="name">全部</span>
            </div>
            <el-scrollbar>
              <div
                v-for="c in categoryList"
                :key="c || 'blank'"
                class="category-row"
                :class="{ active: selectedCategory===c }"
                @click="onCategorySelect(c || '')"
              >
                <span class="name">{{ c || '未分类' }}</span>
              </div>
            </el-scrollbar>
          </div>
        </aside>

        <div class="list-pane">
          <el-table
            :data="inquiryItems"
            stripe
            v-loading="itemsLoading"
            :header-cell-style="{ background: '#f3f4f6' }"
            height="100%"
          >
            <el-table-column type="index" label="序号" width="70" :index="indexMethod" />
            <el-table-column label="商品图" width="90">
              <template #default>
                <div class="img-ph">无</div>
              </template>
            </el-table-column>
            <el-table-column prop="GoodsNameSnap" label="品名" min-width="160" />
            <el-table-column label="拼音" width="120">
              <template #default>—</template>
            </el-table-column>
            <el-table-column prop="SpecNameSnap" label="规格" width="120">
              <template #default="{ row }">{{ row.SpecNameSnap || '—' }}</template>
            </el-table-column>
            <el-table-column prop="UnitNameSnap" label="单位" width="100">
              <template #default="{ row }">{{ row.UnitNameSnap || '—' }}</template>
            </el-table-column>
            <el-table-column prop="LastMonthAvgPrice" label="上期均价" width="110">
              <template #default="{ row }">{{ formatPrice(row.LastMonthAvgPrice) }}</template>
            </el-table-column>
            <el-table-column prop="CurrentAvgPrice" label="本期均价" width="110">
              <template #default="{ row }">{{ formatPrice(row.CurrentAvgPrice) }}</template>
            </el-table-column>
            <el-table-column label="供应商价" width="110">
              <template #default>—</template>
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
      </div>
    </div>

    <!-- 导入Excel弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="询价单导入"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <InquiryImport
        :org-id="organId"
        @close="handleImportClose"
        @import-success="handleImportSuccess"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { InquiryAPI, InquiryItemAPI, type InquiryRow, type InquiryItemRow } from '@/api/inquiry'
import { CategoryAPI, type CategoryRow } from '@/api/category'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { notifyError } from '@/utils/notify'
import InquiryImport from './InquiryImport.vue'

const indexMethod = (rowIndex: number) =>
  (page.value - 1) * pageSize.value + rowIndex + 1
// 登录信息
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const organId = computed(() => jwtPayload.value?.org_id || '')

// 查询条件
const searchYear = ref<number | undefined>()
const searchMonth = ref<number | undefined>()
const searchTenDay = ref<number | undefined>()
const keyword = ref('')

// 年/月/旬选项
const yearOptions = ref<number[]>([])
const monthOptions = ref<number[]>([])
const tenDayOptions = ref<number[]>([])

// 从列表数据提取筛选项
const buildFilterOptions = (items: InquiryRow[]) => {
  const years = new Set<number>()
  const months = new Set<number>()
  const tens = new Set<number>()
  items.forEach(i => {
    if (i.InquiryYear) years.add(i.InquiryYear)
    if (i.InquiryMonth) months.add(i.InquiryMonth)
    if (i.InquiryTenDay) tens.add(i.InquiryTenDay)
  })
  yearOptions.value = Array.from(years).sort((a, b) => b - a)
  monthOptions.value = Array.from(months).sort((a, b) => a - b)
  tenDayOptions.value = Array.from(tens).sort((a, b) => a - b)
}

// 分页与数据
const page = ref(1)
const pageSize = ref(15)
const pageSizes = [10, 15, 20, 50]
const total = ref(0)

const inquiries = ref<InquiryRow[]>([])
const inquiryLoading = ref(false)
const selectedInquiry = ref<InquiryRow | null>(null)

const inquiryItems = ref<InquiryItemRow[]>([])
const itemsLoading = ref(false)
const selectedCategory = ref<string | null>(null)

// 品类列表
const categories = ref<CategoryRow[]>([])
const categoryList = computed(() => {
  return categories.value
    .filter(c => c.Name)
    .sort((a, b) => (a.Sort || 0) - (b.Sort || 0))
    .map(c => c.Name)
})

// 分页（明细）
const handlePageChange = (p: number) => {
  page.value = p
  fetchInquiryItems()
}
const handleSizeChange = (size: number) => {
  pageSize.value = size
  page.value = 1
  fetchInquiryItems()
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
    buildFilterOptions(inquiries.value)

    // 选中项已不在结果中则清空
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

// 获取品类
const fetchCategories = async () => {
  if (!organId.value) {
    categories.value = []
    return
  }
  try {
    const params = { org_id: organId.value, page: 1, page_size: 100 }
    const { data } = await CategoryAPI.list(params)
    categories.value = data?.items || []
  } catch (e) {
    notifyError(e)
  }
}

// 获取询价商品明细
const fetchInquiryItems = async () => {
  if (!selectedInquiry.value) {
    inquiryItems.value = []
    total.value = 0
    return
  }

  itemsLoading.value = true
  try {
    const params: any = {
      inquiry_id: selectedInquiry.value.ID,
      page: page.value,
      page_size: pageSize.value,
    }
    // 分类筛选
    if (selectedCategory.value && selectedCategory.value !== 'ALL') {
      const category = categories.value.find(c => c.Name === selectedCategory.value)
      if (category) {
        params.category_id = category.ID
      }
    }

    const { data } = await InquiryItemAPI.list(params)
    inquiryItems.value = data?.items || []
    total.value = Number(data?.total || 0)
  } catch (e) {
    notifyError(e)
  } finally {
    itemsLoading.value = false
  }
}

// 交互
const onSearch = () => {
  page.value = 1
  fetchInquiries()
}
const onCategorySelect = (key: string) => {
  selectedCategory.value = key
  page.value = 1
  fetchInquiryItems()
}

// 顶部按钮
const onCreateClick = () => { /* TODO: 新增弹窗或路由 */ }
const onEditClick = () => { /* TODO: 编辑模式 */ }
const importDialogVisible = ref(false)
const onImportClick = () => { importDialogVisible.value = true }
const onExportClick = () => { /* TODO: 导出入口 */ }

// 导入弹窗事件
const handleImportClose = () => { importDialogVisible.value = false }
const handleImportSuccess = () => {
  importDialogVisible.value = false
  fetchInquiries()
}

// 工具
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
  fetchCategories()
  fetchInquiries().then(async () => {
    if (!selectedInquiry.value && inquiries.value.length > 0) {
      const latest = inquiries.value[0]
      if (latest) {
        selectedInquiry.value = latest
        page.value = 1
        await fetchInquiryItems()
      }
    }
  })
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

/* 头部三组等间距 */
.header-bar{
  display:flex;
  align-items:flex-start;      /* 让标题两行顶对齐更自然 */
  gap:12px;
  background:#fff;
  border:1px solid #ebeef5;
  border-radius:8px;
  padding:12px;
  justify-content: space-between; /* 三组横向等间距分布 */
  /* 若希望两侧边距也与组间距相等，用 space-evenly：
     justify-content: space-evenly;
  */
}

/* 左：标题竖排 */
.header-bar .title {
  font-weight: 600;
  color: #333;
  display: flex;
  flex-direction: column;   /* 竖排 */
  align-items: flex-start;  /* 左对齐 */
  gap: 2px;
  margin-right: 12px;
  flex: 0 0 auto;           /* 不被挤压 */
}
.header-bar .title .name { font-size: 16px; line-height: 1.2; }
.header-bar .title .sub  { color: #909399; font-size: 12px; }

/* 中：筛选区可伸缩，占中间空间；窄屏可换行 */
.header-bar .filters {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1 1 auto;          /* 中间这组可以拉伸 */
  min-width: 360px;        /* 视情况调整最小宽度 */
  flex-wrap: wrap;         /* 可换行（小屏更友好） */
}

/* 右：动作按钮固定宽度 */
.header-bar .actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 0 0 auto;
}

/* 内容区域 */
.detail-body { display:flex; gap:16px; height: calc(100vh - 220px); }
.category-panel { width: 260px; background:#fff; border:1px solid #ebeef5; border-radius:8px; padding:12px; display:flex; flex-direction:column; gap:10px; }
.panel-title-line { font-weight:600; padding:4px 8px; background:#f5f7fa; border-radius:6px; color:#333; }
.category-list { flex:1; min-height:0; overflow:hidden; }
.category-row { display:flex; align-items:center; justify-content:space-between; padding:10px 10px; cursor:pointer; border-radius:6px; margin:4px 0; }
.category-row:hover { background:#f6f7fb; }
.category-row.active { background:#409eff; color:#fff; }
.category-row .name { white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }

.list-pane {
  flex: 1;
  min-width: 0;
  background:#fff;
  border:1px solid #ebeef5;
  border-radius:8px;
  padding:12px;
  display: flex;
  flex-direction: column;
}

/* 表尾分页 */
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

/* 详情容器 */
.inquiry-items-section {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
