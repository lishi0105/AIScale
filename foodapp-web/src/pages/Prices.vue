<template>
  <div class="page-prices">
    <section class="panel">
      <div class="toolbar">
        <el-select v-model="year" placeholder="年份" style="width: 120px" @change="onSearch">
          <el-option v-for="y in yearOptions" :key="y" :label="`${y}年`" :value="y" />
        </el-select>
        <el-select v-model="month" placeholder="月份" style="width: 120px" @change="onSearch">
          <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
        </el-select>
        <el-select v-model="tenDay" placeholder="旬" style="width: 120px" @change="onSearch">
          <el-option :value="1" label="上旬" />
          <el-option :value="2" label="中旬" />
          <el-option :value="3" label="下旬" />
        </el-select>
        <el-input v-model="keyword" placeholder="请输入标题关键词" clearable style="width: 260px" @clear="onSearch" @keyup.enter="onSearch" />
        <el-button @click="onSearch">查询</el-button>
        <div class="spacer" />
        <el-upload
          :show-file-list="false"
          :auto-upload="false"
          accept=".xlsx,.xls"
          @change="onSelectExcel"
        >
          <el-button type="primary" plain>导入excel</el-button>
        </el-upload>
        <el-button disabled>编辑</el-button>
        <el-button disabled>新建</el-button>
        <el-button disabled>导出excel</el-button>
      </div>

      <el-table :data="rows" style="width: 100%" v-loading="loading" :header-cell-style="{ background: '#f3f4f6' }">
        <el-table-column type="index" label="序号" width="70" />
        <el-table-column prop="InquiryTitle" label="询价单标题" min-width="300" />
        <el-table-column label="日期" width="140">
          <template #default="{ row }">{{ row.InquiryDate?.slice(0, 10) }}</template>
        </el-table-column>
        <el-table-column label="年份" width="90">
          <template #default="{ row }">{{ row.InquiryYear ?? '—' }}</template>
        </el-table-column>
        <el-table-column label="月份" width="90">
          <template #default="{ row }">{{ row.InquiryMonth ?? '—' }}</template>
        </el-table-column>
        <el-table-column label="旬" width="90">
          <template #default="{ row }">{{ tenDayLabel(row.InquiryTenDay) }}</template>
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
    </section>
  </div>
  
  <!-- 导入 Excel 成功提示 -->
  <el-dialog v-model="importResultVisible" title="导入结果" width="520px">
    <div>
      <p>导入成功：共 {{ importResult.count }} 条明细，询价单：{{ importResult.title }}</p>
    </div>
    <template #footer>
      <el-button @click="importResultVisible=false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getToken } from '@/api/http'
import { parseJwt, type JwtPayload } from '@/utils/jwt'
import { notifyError } from '@/utils/notify'
import type { InquiryRow } from '@/types/price'
import { PriceAPI } from '@/types/price'

// 登录信息
const jwtPayload = computed<JwtPayload | null>(() => {
  const token = getToken()
  return token ? parseJwt(token) : null
})
const organId = computed(() => jwtPayload.value?.org_id || '')

// 查询参数与数据
const year = ref<number | undefined>()
const month = ref<number | undefined>()
const tenDay = ref<number | undefined>()
const keyword = ref('')
const page = ref(1)
const pageSize = ref(15)
const pageSizes = [10, 15, 20, 50]
const total = ref(0)
const rows = ref<InquiryRow[]>([])
const loading = ref(false)

const yearOptions = Array.from({ length: 8 }).map((_, i) => new Date().getFullYear() - i)

const tenDayLabel = (v?: number | null) => v === 1 ? '上旬' : v === 2 ? '中旬' : v === 3 ? '下旬' : '—'

const handlePageChange = (p: number) => { page.value = p; fetchList() }
const handleSizeChange = (s: number) => { pageSize.value = s; page.value = 1; fetchList() }
const onSearch = () => { page.value = 1; fetchList() }

async function fetchList() {
  if (!organId.value) { rows.value = []; total.value = 0; return }
  loading.value = true
  try {
    const params: any = { org_id: organId.value, page: page.value, page_size: pageSize.value }
    if (year.value) params.year = year.value
    if (month.value) params.month = month.value
    if (tenDay.value) params.ten_day = tenDay.value
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const { data } = await PriceAPI.inquiryList(params)
    rows.value = data?.items || []
    total.value = Number(data?.total || 0)
  } catch (e) {
    notifyError(e)
  } finally { loading.value = false }
}

// 处理 Excel 选择并导入
const importResultVisible = ref(false)
const importResult = ref<{ title: string; count: number }>({ title: '', count: 0 })

async function onSelectExcel(file: any) {
  try {
    const f: File = file.raw || file
    const buf = await f.arrayBuffer()
    const wb = await (await import('xlsx')).read(buf, { type: 'array' })
    const sheet = wb.Sheets[wb.SheetNames[0]]
    const json = (await import('xlsx')).utils.sheet_to_json<any>(sheet, { defval: '' })
    if (!json.length) { ElMessage.warning('Excel 为空'); return }

    // 粗略取第一行第一格为标题
    const first = json[0] as any
    const firstKey = first ? Object.keys(first)[0] : ''
    const titleValue = (firstKey && first[firstKey]) ? String(first[firstKey]) : ''
    const title = titleValue || `导入${new Date().toISOString().slice(0,10)}`

    // 从标题提取年月旬
    let y = year.value || new Date().getFullYear()
    let m = month.value || (new Date().getMonth() + 1)
    let td = tenDay.value || 1
    const m1 = title.match(/(\d{4})年(\d{1,2})月(上旬|中旬|下旬)/)
    if (m1) {
      y = Number(m1[1])
      m = Number(m1[2])
      td = m1[3] === '上旬' ? 1 : m1[3] === '中旬' ? 2 : 3
    }

    const d = td === 1 ? 5 : td === 2 ? 15 : 25
    const dateStr = `${y}-${String(m).padStart(2,'0')}-${String(d).padStart(2,'0')}`

    if (!organId.value) { ElMessage.warning('缺少中队信息'); return }

    // 创建询价单
    const { data: inq } = await PriceAPI.inquiryCreate({ org_id: organId.value, inquiry_title: title, inquiry_date: dateStr })
    const inquiryId = inq?.ID
    if (!inquiryId) throw new Error('创建询价单失败')

    // 表头定位：尝试在前几行中找包含“品名”的行
    const headIndex = Math.max(0, json.findIndex(r => Object.values(r).some((v: any) => String(v).includes('品名'))))
    const dataRows = json.slice(headIndex + 1)

    let count = 0
    for (const r of dataRows) {
      const goodsName = String(r['品名'] || r['品名/简称'] || r['品种'] || r['品目'] || r['名称'] || '').trim()
      if (!goodsName) continue
      const guide = parseFloat(String(r['发改委指导价'] || r['指导价'] || ''))
      const lastAvg = parseFloat(String(r['上期均价'] || r['上月均价'] || ''))
      const currAvg = parseFloat(String(r['本期均价'] || r['平均价'] || r['本期平均价'] || ''))
      const specName = String(r['规格标准'] || r['规格'] || '').trim() || undefined
      const unitName = String(r['单位'] || '').trim() || undefined
      const categoryName = String(r['品类'] || r['类别'] || '蔬菜类').trim() || '蔬菜类'

      await PriceAPI.inquiryItemCreate({
        inquiry_id: inquiryId,
        goods_id: '00000000-0000-0000-0000-000000000000',
        category_id: '00000000-0000-0000-0000-000000000000',
        spec_id: null,
        unit_id: null,
        goods_name_snap: goodsName,
        category_name_snap: categoryName,
        spec_name_snap: specName || null,
        unit_name_snap: unitName || null,
        guide_price: Number.isFinite(guide) ? guide : null,
        last_month_avg_price: Number.isFinite(lastAvg) ? lastAvg : null,
        current_avg_price: Number.isFinite(currAvg) ? currAvg : null,
      })
      count++
    }

    importResult.value = { title, count }
    importResultVisible.value = true
    ElMessage.success(`导入完成：${count} 条`)
    onSearch()
  } catch (e) { notifyError(e) }
}

onMounted(() => { fetchList() })
</script>

<style scoped>
.panel { background:#fff; border:1px solid #ebeef5; border-radius:8px; padding:16px; display:flex; flex-direction:column }
.toolbar { display:flex; gap:12px; align-items:center; margin-bottom:12px; }
.spacer { flex: 1 }
.pager { display:flex; justify-content:flex-end; padding-top:12px; }
</style>
